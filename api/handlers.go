package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/service"
	"github.com/jackc/pgx/v4"
)

func GetEarthquakesHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Call db function
		earthquakes, err := db.GetEarthquakes(conn)
		if err != nil {
			http.Error(w, "error fetching earthquakes", http.StatusInternalServerError)
			return
		}

		// Step 2: Set JSON header
		w.Header().Set("Content-Type", "application/json")

		// Step 3: Encode to JSON
		json.NewEncoder(w).Encode(earthquakes)
	}
}

func SyncEarthquakesHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call service function
		log.Println("Manually syncing earthquake data from JMA")
		recordsAdded, err := service.SyncEarthquakes(conn)
		if err != nil {
			log.Printf("Error syncing earthquake data: %v", err)
			http.Error(w, "Error syncing earthquake data:", http.StatusInternalServerError)
		}

		log.Printf("Sync complete: added %d new earthquake records", recordsAdded)

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message":       "Sync completed successfully",
			"records_added": recordsAdded,
		}
		json.NewEncoder(w).Encode(response)

	}
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"name":        "Jishin API",
		"version":     "1.0.0",
		"description": "Real-time earthquake data from Japan Meterological Agency (JMA)",
		"endpoints": map[string]string{
			"GET /":                 "API information",
			"GET /health":           "Health check",
			"GET /earthquakes":      "List earthquakes (latest 50)",
			"GET /allEarthquakes":   "List all earthquakes in database",
			"GET /earthquakes/{id}": "Get specific earthquake by report ID",
			"POST /sync":            "Manually sync with JMA data",
		},
		"data_source": "Japan Meterological Agency (JMA)",
		"github":      "https://github.com/Ward-R/Jishin-API",
	}

	json.NewEncoder(w).Encode(response)
}
