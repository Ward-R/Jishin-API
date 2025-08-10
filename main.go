package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"jishin-api/db"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

const JMAQuakeURL = "https://www.jma.go.jp/bosai/quake/data/list.json"

// type Earthquake struct { // final complete struct for reference from db file
// 	ReportId     string
// 	OriginTime   time.Time
// 	ArrivalTime  time.Time
// 	Latitude     float64
// 	Longitude    float64
// 	DepthKm      int
// 	Magnitude    float64
// 	MaxIntensity float64
// 	JpLocation   string
// 	EnLocation   string
// 	TsunamiRisk  string
// }

// The JMA has two json files. A list of each earthquake summary and inside
// a secondary json link to a detail json of the specific earthquake.
// We need to have intermediary stucts to hold this data as we parse it.

// This holds the data from the list of earthquakes.
type QuakeSummary struct {
	ID         string `json:"eid"`
	EnLocation string `json:"en_anm"`
	Magnitude  string `json:"mag"`
	DetailJSON string `json:"json"`
}

// The following structs are the details of one json via DetailJSON:
type EarthquakeBody struct {
	Earthquake JsonEarthquake `json:"Earthquake"`
	Comments   JsonComments   `json:"Comments"`
}

type JsonEarthquake struct {
	OriginTime  string `json:OriginTime"`
	ArrivalTime string `json:ArrivalTime`
	Hypocenter  struct {
		Coordinate string `json:"Coordinate"`
		EnName     string `json:"enName"`
	} `json:"Hypocenter"`
	Magnitude string `json:"Magnitude"`
}

type JsonComments struct {
	ForecastComment struct {
		JpComment string `json:"Text"`
		EnComment string `json:"enText"`
	} `json:"ForecastComment"`
}

func dbConnect() (*pgx.Conn, error) {
	// Load environment variables from .env file.
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

func fetchQuakeData() ([]byte, error) {
	res, err := http.Get(JMAQuakeURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 400 {
		return nil, fmt.Errorf("Response failed with status code: %d", res.StatusCode)
	}
	return body, err
}

func parseQuakeData(data []byte) ([]QuakeSummary, error) {
	var events []QuakeSummary
	err := json.Unmarshal(data, &events)
	return events, err
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
		// rw.Header().Set("Content-Type", "application/json") use this if we want to output to the server in json
		rw.Header().Set("Content-Type", "text/plain")

		data, err := fetchQuakeData()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error fetching data: %v", err)
			return
		}

		events, err := parseQuakeData(data)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error parsing data: %v", err)
			return
		}

		limit := 10
		for i, e := range events {
			if i >= limit {
				break
			}
			fmt.Fprintf(rw, "#%v) ID: %v, Location: %v, Magnitude: %v\n", i, e.ID, e.EnLocation, e.Magnitude)
		}
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
