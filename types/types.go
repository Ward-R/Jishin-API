package types

import (
	"encoding/json"
	"time"
)

type Earthquake struct {
	ReportId     string
	OriginTime   time.Time
	ArrivalTime  time.Time
	Magnitude    float64
	DepthKm      int
	Latitude     float64
	Longitude    float64
	MaxIntensity string
	JpLocation   string
	EnLocation   string
	JpComment    string
	EnComment    string
	TsunamiRisk  string
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
