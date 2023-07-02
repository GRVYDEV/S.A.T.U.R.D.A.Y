package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/ttt/engine"
)

var _ engine.Generator = (*TTTHttpBackend)(nil)
var Logger = logr.New()

// Data format that the http endpoint must return
type GenerateResponse struct {
	Text string `json:"text"`
}

// Data format that the http endpoint must accept
type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

// TTTHttpBackend conforms to the engine.Generator interface and posts
// text to a remote server for TTT inference. It expects the server to return a valid
// GenerateResponse.
type TTTHttpBackend struct {
	url string
}

func New(url string) (*TTTHttpBackend, error) {
	if url == "" {
		return nil, errors.New(fmt.Sprintf("invalid url for TTTHttpBackend %s", url))
	}

	return &TTTHttpBackend{
		url: url,
	}, nil
}

func (t *TTTHttpBackend) Generate(prompt string) (engine.TextChunk, error) {
	var (
		response engine.TextChunk
		err      error
	)
	Logger.Infof("CALLING LLM WITH PROMPT %s", prompt)

	payload, err := json.Marshal(GenerateRequest{Prompt: prompt})
	if err != nil {
		Logger.Error(err, "error marshaling to json")
		return response, err
	}

	resp, err := http.Post(t.url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		Logger.Error(err, "error making llm request")
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger.Error(err, "error reading body")
		return response, err
	}

	if resp.StatusCode == http.StatusOK {
		tttResp := GenerateResponse{}
		err = json.Unmarshal(body, &tttResp)
		if err != nil {
			Logger.Error(err, "error unmarshaling body")
			return response, err
		}

		return engine.TextChunk{Text: tttResp.Text}, err
	} else {
		return response, fmt.Errorf("got bad response %d", resp.StatusCode)
	}

}
