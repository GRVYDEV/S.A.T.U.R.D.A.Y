package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Transcription struct {
	From           uint32
	Transcriptions []TranscriptionSegment
}

type TranscriptionPython struct {
	Transcriptions []TranscriptionSegment `json:"transcriptions"`
}

type TranscriptionSegment struct {
	StartTimestamp uint32 `json:"startTimestamp"`
	EndTimestamp   uint32 `json:"endTimestamp"`
	Text           string `json:"text"`
}

func callTranscriptionAPI(audioData []float32) (error, Transcription) {

	room := os.Getenv("TRASCRIPTION_SERVICE")
	if room == "" {
		room = "http://localhost:8000"
	}
	url := room + "/transcribe" // Replace with the appropriate API URL

	payloadBytes, err := json.Marshal(audioData)
	if err != nil {
		return err, Transcription{}
	}

	// Send POST request to the API
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err, Transcription{}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, Transcription{}
	}

	// Check the response status code
	if resp.StatusCode == http.StatusOK {
		transcription := TranscriptionPython{}
		err = json.Unmarshal(body, &transcription)
		if err != nil {
			return err, Transcription{}
		}
		retValue := Transcription{}
		retValue.Transcriptions = transcription.Transcriptions
		return nil, retValue
	} else {
		return fmt.Errorf("Error: %s", body), Transcription{}
	}
}
