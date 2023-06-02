package faster_whisper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"S.A.T.U.R.D.A.Y/stt/engine"
)

// ensure this satisfies the interface
var _ engine.Transcriber = (*FasterWhisperApi)(nil)

type FasterWhisperApi struct {
	url string
}

type transcriptionPython struct {
	Transcriptions []engine.TranscriptionSegment `json:"transcriptions"`
}

func New(url string) (*FasterWhisperApi, error) {
	if url == "" {
		return nil, errors.New(fmt.Sprintf("invalid url for FasterWhisperApi %s", url))
	}
	return &FasterWhisperApi{
		url: url,
	}, nil
}

func (f *FasterWhisperApi) Transcribe(audioData []float32) (engine.Transcription, error) {
	payloadBytes, err := json.Marshal(audioData)
	if err != nil {
		return engine.Transcription{}, err
	}

	// Send POST request to the API
	resp, err := http.Post(f.url, "application/json", bytes.NewBuffer(payloadBytes))
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
	if resp.StatusCode == http.StatusOK {
		transcription := transcriptionPython{}
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
