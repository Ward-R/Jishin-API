package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/types"
	"github.com/jackc/pgx/v4"
)

const (
	JMAQuakeURL = "https://www.jma.go.jp/bosai/quake/data/list.json"
)

// fetchQuakeData fetches the list of earthquake summaries from the JMA.
func FetchQuakeData() ([]byte, error) {
	res, err := http.Get(JMAQuakeURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch data: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 400 {
		return nil, fmt.Errorf("Response failed with status code: %d", res.StatusCode)
	}
	return body, err
}

// fetchDetailQuakeData fetches the detailed JSON for a specific earthquake.
func FetchDetailQuakeData(detailJSON string) ([]byte, error) {
	rootURL := "https://www.jma.go.jp/bosai/quake/data/"
	detailURL := rootURL + detailJSON
	res, err := http.Get(detailURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 400 {
		return nil, fmt.Errorf("Response failed with status code: %d", res.StatusCode)
	}
	return body, err
}

// Helper function:
// parseCoordinate parses the concatenated coordinate string into lat, long, depth, and tsunami risk.
func parseCoordinate(cod string) (latitude, longitude, depthKm float64, tsunamiRisk string, err error) {
	// The regex pattern is updated to make the depth component optional.
	re := regexp.MustCompile(`([+-]?\d+\.\d+)([+-]?\d+\.\d+)([+-]\d+)?\/?(\d*)?`)
	matches := re.FindStringSubmatch(cod)

	if len(matches) < 3 {
		return 0, 0, 0, "", fmt.Errorf("failed to parse coordinate string: %s", cod)
	}

	latitude, err = strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("failed to parse latitude: %w", err)
	}

	longitude, err = strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("failed to parse longitude: %w", err)
	}

	// Check if the depth group was captured before attempting to parse it.
	if len(matches) > 3 && matches[3] != "" {
		depthMeters, err := strconv.ParseFloat(matches[3], 64)
		if err != nil {
			return 0, 0, 0, "", fmt.Errorf("failed to parse depth: %w", err)
		}
		depthKm = depthMeters / 1000
	}

	// Check if the tsunami risk group was captured.
	if len(matches) > 4 {
		tsunamiRisk = matches[4]
	}

	return latitude, longitude, depthKm, tsunamiRisk, nil
}

// parseQuakeData unmarshals the list of quake summaries.
func ParseQuakeData(data []byte) ([]types.QuakeSummary, error) {
	var events []types.QuakeSummary
	err := json.Unmarshal(data, &events)
	return events, err
}

// parseDetailQuakeData takes the raw detailed JSON and builds the final Earthquake struct.
func ParseDetailQuakeData(id string, data []byte) (*types.Earthquake, error) {
	quake := &types.Earthquake{}
	var detailData types.DetailQuakeReport
	err := json.Unmarshal(data, &detailData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling detailed data: %w", err)
	}

	quake.ReportId = id

	if detailData.Body.Earthquake.OriginTime != "" {
		parsedOriginTime, err := time.Parse(time.RFC3339, detailData.Body.Earthquake.OriginTime)
		if err != nil {
			return nil, fmt.Errorf("error parsing origin time: %w", err)
		}
		quake.OriginTime = parsedOriginTime
	}

	if detailData.Body.Earthquake.ArrivalTime != "" {
		parsedArrivalTime, err := time.Parse(time.RFC3339, detailData.Body.Earthquake.ArrivalTime)
		if err != nil {
			return nil, fmt.Errorf("error parsing arrival time: %w", err)
		}
		quake.ArrivalTime = parsedArrivalTime
	}

	coordinate := strings.TrimSpace(detailData.Body.Earthquake.Hypocenter.Area.Coordinate)
	if coordinate != "" {
		lat, long, depth, tsunami, err := parseCoordinate(coordinate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coordinates for ID %s: %w", id, err)
		}
		quake.Latitude = lat
		quake.Longitude = long
		quake.DepthKm = int(depth)
		quake.TsunamiRisk = tsunami
	}

	if detailData.Body.Earthquake.Magnitude != "" {
		quake.Magnitude, _ = strconv.ParseFloat(detailData.Body.Earthquake.Magnitude, 64)
	}
	quake.EnLocation = detailData.Body.Earthquake.Hypocenter.Area.EnName
	quake.JpLocation = detailData.Body.Earthquake.Hypocenter.Area.JpName
	quake.MaxIntensity = detailData.Body.Intensity.Observation.MaxIntensity
	quake.JpComment = detailData.Body.Comments.ForecastComment.Text
	quake.EnComment = detailData.Body.Comments.ForecastComment.EnText

	return quake, nil
}

// function syncs database with JMA earthquake data when called. If it exists,
// in the db already it skips adding it based on recordID.
func SyncEarthquakes(conn *pgx.Conn) (int, error) {
	data, err := FetchQuakeData()
	recordsAdded := 0
	if err != nil {
		return 0, fmt.Errorf("error fetching summary data: %w", err)
	}

	events, err := ParseQuakeData(data)
	if err != nil {
		return 0, fmt.Errorf("error parsing summary data: %w", err)
	}

	for _, event := range events {
		detailData, err := FetchDetailQuakeData(event.DetailJSON)
		if err != nil {
			log.Printf("Error fetching detailed data for ID %s: %v", event.ID, err)
			continue
		}

		earthquake, err := ParseDetailQuakeData(event.ID, detailData)
		if err != nil {
			log.Printf("Error parsing data for ID %s: %v", event.ID, err)
			continue
		}

		// Check if earthquake already exists
		exists, err := db.EarthquakeExists(conn, earthquake.ReportId)
		if err != nil {
			// Log error but continue with next earthquake
			continue
		}

		if !exists {
			// Insert new earthquake
			err = db.InsertEarthquake(conn, earthquake)
			if err != nil {
				// Log error but continue
				continue
			}
			recordsAdded++ // Increment counter
		}
	}
	return recordsAdded, nil
}
