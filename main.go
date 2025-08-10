package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"jishin-api/db"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

const (
	JMAQuakeURL = "https://www.jma.go.jp/bosai/quake/data/list.json"
)

// Earthquake is the final, complete struct for an earthquake report.
type Earthquake struct {
	ReportId     string
	OriginTime   time.Time
	ArrivalTime  time.Time
	Latitude     float64
	Longitude    float64
	DepthKm      int
	Magnitude    float64
	MaxIntensity string
	JpLocation   string
	EnLocation   string
	TsunamiRisk  string
	JpComment    string
	EnComment    string
}

// QuakeSummary holds the data from the list of earthquakes.
type QuakeSummary struct {
	ID         string `json:"eid"`
	DetailJSON string `json:"json"`
}

// DetailQuakeReport is the intermediate struct for parsing the detailed JSON.
type DetailQuakeReport struct {
	Control json.RawMessage `json:"Control"`
	Head    json.RawMessage `json:"Head"`
	Body    struct {
		Earthquake JsonEarthquake `json:"Earthquake"`
		Intensity  JsonIntensity  `json:"Intensity"`
		Comments   JsonComments   `json:"Comments"`
	} `json:"Body"`
}

// JsonEarthquake contains the core quake info, matching the JSON exactly.
type JsonEarthquake struct {
	OriginTime  string `json:"OriginTime"`
	ArrivalTime string `json:"ArrivalTime"`
	Magnitude   string `json:"Magnitude"`
	Hypocenter  struct {
		Area struct {
			Coordinate string `json:"Coordinate"`
			JpName     string `json:"Name"`
			EnName     string `json:"enName"`
		} `json:"Area"`
	} `json:"Hypocenter"`
}

// JsonComments holds the comment data.
type JsonComments struct {
	ForecastComment struct {
		Text   string `json:"Text"`
		Code   string `json:"Code"`
		EnText string `json:"enText"`
	} `json:"ForecastComment"`
	VarComment struct {
		Text   string `json:"Text"`
		Code   string `json:"Code"`
		EnText string `json:"enText"`
	} `json:"VarComment"`
}

// JsonIntensity is for getting the maximum seismic intensity.
type JsonIntensity struct {
	Observation struct {
		MaxIntensity string `json:"MaxInt"`
	} `json:"Observation"`
}

// dbConnect loads environment variables and connects to the database.
func dbConnect() (*pgx.Conn, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	conn, err := db.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return conn, nil
}

// fetchQuakeData fetches the list of earthquake summaries from the JMA.
func fetchQuakeData() ([]byte, error) {
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

// parseQuakeData unmarshals the list of quake summaries.
func parseQuakeData(data []byte) ([]QuakeSummary, error) {
	var events []QuakeSummary
	err := json.Unmarshal(data, &events)
	return events, err
}

// fetchDetailQuakeData fetches the detailed JSON for a specific earthquake.
func fetchDetailQuakeData(detailJSON string) ([]byte, error) {
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

// parseDetailQuakeData takes the raw detailed JSON and builds the final Earthquake struct.
func parseDetailQuakeData(id string, data []byte) (*Earthquake, error) {
	quake := &Earthquake{}
	var detailData DetailQuakeReport
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

func main() {
	log.Println("Starting Jishin API...")

	conn, err := dbConnect()
	if err != nil {
		log.Fatalf("Application startup failed: %v", err)
	}
	defer conn.Close(context.Background())
	fmt.Println("Successfully connected to the database!")

	http.HandleFunc("/earthquakes", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

		data, err := fetchQuakeData()
		if err != nil {
			log.Printf("Error fetching data: %v", err)
			http.Error(rw, "Error fetching summary data", http.StatusInternalServerError)
			return
		}

		events, err := parseQuakeData(data)
		if err != nil {
			log.Printf("Error parsing data: %v", err)
			http.Error(rw, "Error parsing summary data", http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			detailData, err := fetchDetailQuakeData(event.DetailJSON)
			if err != nil {
				log.Printf("Error fetching detailed data for ID %s: %v", event.ID, err)
				continue
			}

			earthquake, err := parseDetailQuakeData(event.ID, detailData)
			if err != nil {
				log.Printf("Error parsing data for ID %s: %v", event.ID, err)
				continue
			}

			fmt.Fprintln(rw, "------------------------------------")
			fmt.Fprintln(rw, "Parsed Earthquake Data:")
			fmt.Fprintf(rw, "  Report ID: %s\n", earthquake.ReportId)
			fmt.Fprintf(rw, "  Origin Time: %s\n", earthquake.OriginTime.Format(time.RFC3339))
			fmt.Fprintf(rw, "  Arrival Time: %s\n", earthquake.ArrivalTime.Format(time.RFC3339))
			fmt.Fprintf(rw, "  Latitude: %f\n", earthquake.Latitude)
			fmt.Fprintf(rw, "  Longitude: %f\n", earthquake.Longitude)
			fmt.Fprintf(rw, "  Depth: %d km\n", earthquake.DepthKm)
			fmt.Fprintf(rw, "  Magnitude: %f\n", earthquake.Magnitude)
			fmt.Fprintf(rw, "  Max Intensity: %s\n", earthquake.MaxIntensity)
			fmt.Fprintf(rw, "  English Location: %s\n", earthquake.EnLocation)
			fmt.Fprintf(rw, "  Japanese Location: %s\n", earthquake.JpLocation)

			if earthquake.TsunamiRisk != "" {
				fmt.Fprintf(rw, "  Tsunami Risk: %s\n", earthquake.TsunamiRisk)
			} else {
				fmt.Fprintf(rw, "  Tsunami Risk: No information available\n")
			}

			if earthquake.JpComment != "" {
				fmt.Fprintf(rw, "  Japanese Comment: %s\n", earthquake.JpComment)
			} else {
				fmt.Fprintf(rw, "  Japanese Comment: No comment available\n")
			}
			if earthquake.EnComment != "" {
				fmt.Fprintf(rw, "  English Comment: %s\n", earthquake.EnComment)
			} else {
				fmt.Fprintf(rw, "  English Comment: No comment available\n")
			}
			fmt.Fprintln(rw, "------------------------------------")
		}
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
