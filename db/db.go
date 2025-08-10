package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

// type Earthquake struct { // final complete struct
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

// type Earthquake struct {
// 	// can add more from the JMA JSON, just putting a couple to get it working for now.
// 	ReportId    string `json:"eid"`
// 	OriginTime  int    `json:"at"`
// 	ArrivalTime time.Time
// 	Latitude
// 	Longitude
// 	DepthKm      int    `json:"`
// 	Magnitude    int    `json:"mag"`
// 	MaxIntensity int    `json:"maxi"`
// 	JpLocation   string `json:"anm"`
// 	EnLocation   string `json:"en_anm"`
// 	TsunamiRisk  string `json:"cod"`
// }
