package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/service"
	"github.com/jackc/pgx/v4"
)

func GetEarthquakesHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		magnitudeStr := r.URL.Query().Get("magnitude")
		dateStr := r.URL.Query().Get("date")

		// Convert to proper types with defaults
		limit := 0       // Will use the default (50) in DB function
		magnitude := 0.0 // will mean no filter in DB function

		if limitStr != "" {
			limit, _ = strconv.Atoi(limitStr)
		}
		if magnitudeStr != "" {
			magnitude, _ = strconv.ParseFloat(magnitudeStr, 64)
		}
		// Step 1: Call db function
		earthquakes, err := db.GetEarthquakes(conn, limit, magnitude, dateStr)
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

func GetEarthquakeByIdHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path like "/earthquakes/20250812113450"
		path := r.URL.Path
		// Remove prefix to get id
		id := strings.TrimPrefix(path, "/earthquake/")

		if id == "" || id == path { // No ID provided or invalid path
			http.Error(w, "Earthquake ID required", http.StatusBadRequest)
			return
		}

		// Call database function to get single earthquake
		earthquake, err := db.GetEarthquakeById(conn, id)
		if err != nil {
			http.Error(w, "Earthquake not found", http.StatusNotFound)
			return
		}
		// Step 2: Set JSON header
		w.Header().Set("Content-Type", "application/json")

		// Step 3: Encode to JSON
		json.NewEncoder(w).Encode(earthquake)
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
			"GET /":                            "API information",
			"GET /health":                      "Health check",
			"GET /earthquakes":                 "Default 50 earthquakes",
			"GET /earthquakes/recent":          "Gets all earthquakes in last 24 hours",
			"GET /earthquakes/stats":           "Summary statistics and data overview",
			"GET /earthquakes?limit=10":        "10 earthquakes",
			"GET /earthquakes?limit=-1":        "ALL earthquakes",
			"GET /earthquakes?magnitude=5.0":   "Earthquakes 5.0+ magnitude",
			"GET /earthquakes?date=2025-08-12": "Earthquakes from specific date (YYYY-MM-DD)",
			"GET /earthquakes?limit=5&magnitude=4.0&date=2025-08-12": "Combined filters example",
			"GET /earthquake/{id}": "Get specific earthquake by report ID",
			"POST /sync":           "Manually sync with JMA data",
		},
		"data_source": "Japan Meterological Agency (JMA)",
		"github":      "https://github.com/Ward-R/Jishin-API",
	}
	json.NewEncoder(w).Encode(response)
}

func HealthHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Test database connection
		err := conn.Ping(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			response := map[string]interface{}{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// All good
		response := map[string]interface{}{
			"status":    "healthy",
			"database":  "connected",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}
}

func GetRecentEarthquakesHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		earthquakes, err := db.GetRecentEarthquakes(conn)
		if err != nil {
			http.Error(w, "Error fetching recent earthquakes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Check if no earthquakes found
		if len(earthquakes) == 0 {
			response := map[string]interface{}{
				"message":   "No earthquakes found in the last 24 hours",
				"count":     0,
				"timeframe": "24 hours",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return earthquakes with count info
		response := map[string]interface{}{
			"count":       len(earthquakes),
			"timeframe":   "24 hours",
			"earthquakes": earthquakes,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func GetEarthquakeStatsHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := db.GetEarthquakeStats(conn)
		if err != nil {
			http.Error(w, "Error fetching earthquake statistics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
