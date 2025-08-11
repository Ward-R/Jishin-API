package main

import (
	"context"
	"net/http"

	"fmt"
	"log"

	"time"

	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/service"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

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

		data, err := service.FetchQuakeData()
		if err != nil {
			log.Printf("Error fetching data: %v", err)
			http.Error(rw, "Error fetching summary data", http.StatusInternalServerError)
			return
		}

		events, err := service.ParseQuakeData(data)
		if err != nil {
			log.Printf("Error parsing data: %v", err)
			http.Error(rw, "Error parsing summary data", http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			detailData, err := service.FetchDetailQuakeData(event.DetailJSON)
			if err != nil {
				log.Printf("Error fetching detailed data for ID %s: %v", event.ID, err)
				continue
			}

			earthquake, err := service.ParseDetailQuakeData(event.ID, detailData)
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
