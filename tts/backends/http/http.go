package tts_http

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/engine"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/util"
)

// ensure this satisfies the interface
var _ engine.Synthesizer = (*TTSHttpBackend)(nil)
var Logger = logr.New()

// Data format that the http endpoint must return
type SynthesizeResponse struct {
	// This must be base64 encoded float32 little endian binary data
	Data         string `json:"data"`
	SampleRate   int    `json:"sample_rate"`
	ChannelCount int    `json:"channel_count"`
}

// Data format that the http endpoint must accept
type SynthesisRequest struct {
	Text string `json:"text"`
}

// TTSHttpBackend conforms to the engine.Synthesizer interface and posts
// text to a remote server for TTS. It expects the server to return a valid
// SynthesizeResponse with base64 encoded float32 little endian binary data
type TTSHttpBackend struct {
	url string
}

func New(url string) (*TTSHttpBackend, error) {
	if url == "" {
		return nil, errors.New(fmt.Sprintf("invalid url for TTSHttpBackend %s", url))
	}

	return &TTSHttpBackend{
		url: url,
	}, nil
}

func (t *TTSHttpBackend) Synthesize(text string) (engine.AudioChunk, error) {
	chunk := engine.AudioChunk{}

	payload, err := json.Marshal(SynthesisRequest{Text: text})
	if err != nil {
		return chunk, err
	}

	resp, err := http.Post(t.url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return chunk, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return chunk, err
	}

	if resp.StatusCode == http.StatusOK {
		response := SynthesizeResponse{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return chunk, err
		}

		if response.SampleRate == 0 {
			return chunk, fmt.Errorf("invalid SampleRate from server %d", response.SampleRate)
		}

		buf := make([]byte, len(response.Data))

		n, err := b64.StdEncoding.Decode(buf, []byte(response.Data))
		if err != nil {
			Logger.Error(err, "error decoding b64")
			return chunk, err
		}

		chunk.Data = util.BinaryToFloat32(buf[:n])
		chunk.SampleRate = response.SampleRate

		// FIXME do validation since we only suppport 1 or 2 chan
		if response.ChannelCount == 0 {
			chunk.ChannelCount = 1
		} else {
			chunk.ChannelCount = response.ChannelCount
		}

		return chunk, nil

	} else {
		return engine.AudioChunk{}, fmt.Errorf("invalid status code %d: %s", resp.StatusCode, body)
	}
}
