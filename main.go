package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const JMAQuakeURL = "https://www.jma.go.jp/bosai/quake/data/list.json"

type QuakeSummary struct {
	// can add more from the JMA JSON, just putting a couple to get it working for now.
	ID         string `json:"eid"`
	EnLocation string `json:"en_anm"`
	Magnitude  string `json:"mag"`
}

func main() {
	fmt.Println("Hello, Jishin API!")

	res, err := http.Get(JMAQuakeURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 400 {
		log.Fatalf("Response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshall body into events slice of QuakeSummary structs

	var events []QuakeSummary
	err = json.Unmarshal([]byte(body), &events)

	fmt.Printf("%+v", events)
}
