package db

import (
	"context"
	"fmt"
	"os"

	"github.com/Ward-R/Jishin-API/types"
	"github.com/jackc/pgx/v4"
)

// Note to self. to login via terminal
// psql -U admin -d jishin_db -h localhost

// Connect establishes a connection to the PostgreSQL database.
func Connect() (*pgx.Conn, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return conn, nil
}

// EarthquakeExists checks if an earthquake record already exists in the database
func EarthquakeExists(conn *pgx.Conn, reportID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM earthquakes WHERE report_id = $1)`
	err := conn.QueryRow(context.Background(), query, reportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking earthquake existence: %w", err)
	}
	return exists, nil
}

func InsertEarthquake(conn *pgx.Conn, quake *types.Earthquake) error {
	query := `
        INSERT INTO earthquakes (
            report_id, origin_time, arrival_time, magnitude,
            depth_km, latitude, longitude, max_intensity,
            jp_location, en_location, jp_comment, en_comment,
            tsunami_risk
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := conn.Exec(context.Background(), query,
		quake.ReportId,
		quake.OriginTime,
		quake.ArrivalTime,
		quake.Magnitude,
		quake.DepthKm,
		quake.Latitude,
		quake.Longitude,
		quake.MaxIntensity,
		quake.JpLocation,
		quake.EnLocation,
		quake.JpComment,
		quake.EnComment,
		quake.TsunamiRisk,
	)

	if err != nil {
		return fmt.Errorf("error inserting earthquake: %w", err)
	}
	return nil
}

func GetEarthquakes(conn *pgx.Conn) ([]types.Earthquake, error) {
	query := `
			SELECT report_id, origin_time, arrival_time, magnitude, depth_km,
      	latitude, longitude, max_intensity, jp_location, en_location,
      	jp_comment, en_comment, tsunami_risk
			FROM earthquakes
      ORDER BY origin_time
			DESC LIMIT 50`

	// Execute the query
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("error getting earthquakes: %w", err)
	}
	defer rows.Close()

	// Create slice to hold the results
	var earthquakes []types.Earthquake

	// Loop through rows and scan into structs
	for rows.Next() {
		var eq types.Earthquake
		err := rows.Scan(
			&eq.ReportId, &eq.OriginTime, &eq.ArrivalTime, &eq.Magnitude,
			&eq.DepthKm, &eq.Latitude, &eq.Longitude, &eq.MaxIntensity,
			&eq.JpLocation, &eq.EnLocation, &eq.JpComment, &eq.EnComment,
			&eq.TsunamiRisk,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		earthquakes = append(earthquakes, eq)
	}

	return earthquakes, nil
}
