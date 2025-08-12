package main

import (
	"context"
	"net/http"

	"fmt"
	"log"

	"github.com/Ward-R/Jishin-API/api"
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

	// Connect to postgres database
	conn, err := dbConnect()
	if err != nil {
		log.Fatalf("Application startup failed: %v", err)
	}
	defer conn.Close(context.Background())
	fmt.Println("Successfully connected to the database!")

	// Forced data sync
	log.Println("Syncing earthquake data from JMA")
	recordsAdded, err := service.SyncEarthquakes(conn)
	if err != nil {
		log.Printf("Error syncing earthquake data: %v", err)
	} else {
		log.Printf("Sync complete: added %d new earthquake records", recordsAdded)
	}

	// Routes
	http.HandleFunc("/", api.RootHandler)
	http.HandleFunc("/earthquakes", api.GetEarthquakesHandler(conn))
	http.HandleFunc("/sync", api.SyncEarthquakesHandler(conn))

	// Start localhost
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
