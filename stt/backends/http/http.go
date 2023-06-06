package stt_http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"
)

// ensure this satisfies the interface
var _ engine.Transcriber = (*STTHttpBackend)(nil)

type STTHttpBackend struct {
	url string
}

type TranscribeResponse struct {
	Transcriptions []engine.TranscriptionSegment `json:"transcriptions"`
}

func New(url string) (*STTHttpBackend, error) {
	if url == "" {
		return nil, errors.New(fmt.Sprintf("invalid url for STTHttpBackend %s", url))
	}
	return &STTHttpBackend{
		url: url,
	}, nil
}

func (s *STTHttpBackend) Transcribe(audioData []float32) (engine.Transcription, error) {
	payloadBytes, err := json.Marshal(audioData)
	if err != nil {
		return engine.Transcription{}, err
	}

	// Send POST request to the API
	resp, err := http.Post(s.url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return engine.Transcription{}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return engine.Transcription{}, err
	}

	// Check the response status code
	if resp.StatusCode == http.StatusCreated {
		transcription := TranscribeResponse{}
		err = json.Unmarshal(body, &transcription)
		if err != nil {
			return engine.Transcription{}, err
		}
		retValue := engine.Transcription{}
		retValue.Transcriptions = transcription.Transcriptions
		return retValue, err
	} else {
		return engine.Transcription{}, fmt.Errorf("Error: %s", body)
	}
}
