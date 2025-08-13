package main

import (
	"context"
	"log"
	"strings"

	"github.com/Ward-R/Jishin-API/api"
	"github.com/Ward-R/Jishin-API/db"
	"github.com/Ward-R/Jishin-API/service"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v4"
)

// Global db connection for lambda invocations
var dbConn *pgx.Conn

func init() {
	// Initialize database connection once
	var err error
	dbConn, err = db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Sync earthquake data on startup
	log.Println("Syncing earthquake data from JMA on startup")
	recordsAdded, err := service.SyncEarthquakes(dbConn)
	if err != nil {
		log.Printf("Error syncing earthquake data: %v", err)
	} else {
		log.Printf("Startup sync complete: added %d new earthquake records", recordsAdded)
	}
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Route based on path and method
	path := request.Path
	method := request.HTTPMethod

	switch {
	// Simple routes
	case path == "/" && method == "GET":
		return api.HandleRoot(dbConn)
	case path == "/health" && method == "GET":
		return api.HandleHealth(dbConn)
	case path == "/earthquakes/stats" && method == "GET":
		return api.HandleStats(dbConn)
	case path == "/earthquakes/recent" && method == "GET":
		return api.HandleRecent(dbConn)
	case path == "/earthquakes/largest/today" && method == "GET":
		return api.HandleLargestToday(dbConn)
	case path == "/earthquakes/largest/week" && method == "GET":
		return api.HandleLargestWeek(dbConn)
	// Complex routes
	case path == "/earthquakes" && method == "GET":
		return api.HandleEarthquakes(dbConn, request) // Needs ?Limit=X&magnitude=Y
	case strings.HasPrefix(path, "/earthquake/") && method == "GET":
		return api.HandleEarthquakeById(dbConn, request) // Needs ID from path
	case path == "/sync" && method == "POST":
		return api.HandleSync(dbConn)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       `{"error": "Not found"}`,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
