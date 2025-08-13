package api

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/service"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jackc/pgx/v4"
)

func HandleEarthquakes(dbConn *pgx.Conn, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse query parameters (Lambda way)
	limitStr := request.QueryStringParameters["limit"]
	magnitudeStr := request.QueryStringParameters["magnitude"]
	dateStr := request.QueryStringParameters["date"]

	// Convert to proper types with defaults
	limit := 0       // Will use the default (50) in DB function
	magnitude := 0.0 // will mean no filter in DB function

	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	if magnitudeStr != "" {
		magnitude, _ = strconv.ParseFloat(magnitudeStr, 64)
	}

	// Call db function
	earthquakes, err := db.GetEarthquakes(dbConn, limit, magnitude, dateStr)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Error fetching earthquakes"}`,
		}, nil
	}

	// Marshal to JSON
	body, _ := json.Marshal(earthquakes)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleRoot(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	response := map[string]interface{}{
		"name":        "Jishin API",
		"version":     "1.0.0",
		"description": "Real-time earthquake data from Japan Meteorological Agency (JMA)",
		"endpoints": map[string]string{
			"GET /":                                                  "API information",
			"GET /health":                                            "Health check",
			"GET /earthquakes":                                       "Default 50 earthquakes",
			"GET /earthquakes/largest/today":                         "Strongest earthquake today",
			"GET /earthquakes/largest/week":                          "Strongest earthquake this week",
			"GET /earthquakes/recent":                                "Gets all earthquakes in last 24 hours",
			"GET /earthquakes/stats":                                 "Summary statistics and data overview",
			"GET /earthquakes?limit=10":                              "10 earthquakes",
			"GET /earthquakes?limit=-1":                              "ALL earthquakes",
			"GET /earthquakes?magnitude=5.0":                         "Earthquakes 5.0+ magnitude",
			"GET /earthquakes?date=2025-08-12":                       "Earthquakes from specific date (YYYY-MM-DD)",
			"GET /earthquakes?limit=5&magnitude=4.0&date=2025-08-12": "Combined filters example",
			"GET /earthquake/{id}":                                   "Get specific earthquake by report ID",
			"POST /sync":                                             "Manually sync with JMA data",
		},
		"data_source": "Japan Meteorological Agency (JMA)",
		"github":      "https://github.com/Ward-R/Jishin-API",
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleHealth(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	// Test db connection
	err := dbConn.Ping(context.Background())
	if err != nil {
		response := map[string]interface{}{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		}
		body, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 503,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       string(body),
		}, nil
	}

	// All good
	response := map[string]interface{}{
		"status":    "healthy",
		"database":  "connected",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleRecent(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	earthquakes, err := db.GetRecentEarthquakes(dbConn)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Error fetching recent earthquakes"}`,
		}, nil
	}

	// Check if no earthquakes found
	if len(earthquakes) == 0 {
		response := map[string]interface{}{
			"message":   "No earthquakes found in the last 24 hours",
			"count":     0,
			"timeframe": "24 hours",
		}
		body, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       string(body),
		}, nil
	}

	// Return earthquakes with count info
	response := map[string]interface{}{
		"count":       len(earthquakes),
		"timeframe":   "24 hours",
		"earthquakes": earthquakes,
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleStats(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	stats, err := db.GetEarthquakeStats(dbConn)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Error fetching earthquake statistics"}`,
		}, nil
	}

	body, _ := json.Marshal(stats)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleLargestToday(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	earthquake, err := db.GetLargestEarthquakeToday(dbConn)
	if err != nil {
		response := map[string]interface{}{
			"message": "No earthquakes found today",
			"period":  "today",
		}
		body, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       string(body),
		}, nil
	}

	response := map[string]interface{}{
		"period":             "today",
		"largest_earthquake": earthquake,
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleLargestWeek(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	earthquake, err := db.GetLargestEarthquakeThisWeek(dbConn)
	if err != nil {
		response := map[string]interface{}{
			"message": "No earthquakes found this week",
			"period":  "this week",
		}
		body, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       string(body),
		}, nil
	}

	response := map[string]interface{}{
		"period":             "this week",
		"largest_earthquake": earthquake,
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleSync(dbConn *pgx.Conn) (events.APIGatewayProxyResponse, error) {
	log.Println("Manually syncing earthquake data from JMA")
	recordsAdded, err := service.SyncEarthquakes(dbConn)
	if err != nil {
		log.Printf("Error syncing earthquake data: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Error syncing earthquake data"}`,
		}, nil
	}

	log.Printf("Sync complete: added %d new earthquake records", recordsAdded)
	response := map[string]interface{}{
		"message":       "Sync completed successfully",
		"records_added": recordsAdded,
	}
	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}

func HandleEarthquakeById(dbConn *pgx.Conn, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract ID from URL path like "/earthquake/20250812113450"
	path := request.Path
	id := strings.TrimPrefix(path, "/earthquake/")

	if id == "" || id == path {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Earthquake ID required"}`,
		}, nil
	}

	earthquake, err := db.GetEarthquakeById(dbConn, id)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
			Body:       `{"error": "Earthquake not found"}`,
		}, nil
	}

	body, _ := json.Marshal(earthquake)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(body),
	}, nil
}