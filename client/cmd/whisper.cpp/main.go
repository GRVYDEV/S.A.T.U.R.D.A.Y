package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/client"
	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	whisper "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/backends/whisper.cpp"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"
	stt "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"
	shttp "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/backends/http"
	tts "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/engine"

	"golang.org/x/exp/slog"
)

const llmTime = time.Second * 2

var (
	debug  = flag.Bool("debug", false, "print debug logs")
	logger = logr.New()
)

func main() {
	flag.Parse()
	if *debug {
		logr.SetLevel(slog.LevelDebug)
	}

	urlEnv := os.Getenv("URL")
	if urlEnv == "" {
		urlEnv = "localhost:8088"
	}

	room := os.Getenv("ROOM")
	if room == "" {
		room = "test"
	}

	url := url.URL{Scheme: "ws", Host: urlEnv, Path: "/ws"}

	// FIXME read from env
	whisperCpp, err := whisper.New("../models/ggml-base.en.bin")
	if err != nil {
		logger.Fatal(err, "error creating whisper model")
	}

	transcriptionStream := make(chan engine.Document, 100)

	synthesizer, err := shttp.New("http://localhost:8000/synthesize")
	if err != nil {
		logger.Fatal(err, "error creating http")
	}

	ttsEngine, err := tts.New(tts.EngineParams{
		Synthesizer: synthesizer,
	})

	llm := NewLLM("http://localhost:9090/eva", ttsEngine)
	go llm.Start()
	defer llm.Stop()

	onDocumentUpdate := func(document engine.Document) {
		transcriptionStream <- document
		llm.UpdatePrompt(document.NewText)
	}

	documentComposer := stt.NewDocumentComposer()
	documentComposer.FilterSegment(func(ts stt.TranscriptionSegment) bool {
		return ts.Text[0] == '.' || strings.ContainsAny(ts.Text, "[]()")
	})

	sttEngine, err := stt.New(stt.EngineParams{
		Transcriber:      whisperCpp,
		OnDocumentUpdate: onDocumentUpdate,
		DocumentComposer: documentComposer,
		UseVad:           true,
	})

	sc, err := client.NewSaturdayClient(client.SaturdayConfig{
		Room:                room,
		Url:                 url,
		SttEngine:           sttEngine,
		TtsEngine:           ttsEngine,
		TranscriptionStream: transcriptionStream,
	})
	if err != nil {
		logger.Fatal(err, "error creating saturday client")
	}

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}

// LLMInterface will call an llm with the specified prompt every 3 seconds and
// turn the response into audio
type LLMInterface struct {
	ttsEngine *tts.Engine
	timer     *time.Timer
	prompt    string
	cancel    chan int

	llmUrl string
	sync.Mutex
}

type LLMRequest struct {
	Prompt string `json:"prompt"`
}

type LLMResponse struct {
	Text string `json:"text"`
}

func NewLLM(url string, ttsEngine *tts.Engine) *LLMInterface {
	return &LLMInterface{
		llmUrl:    url,
		timer:     time.NewTimer(llmTime),
		prompt:    "",
		ttsEngine: ttsEngine,
		cancel:    make(chan int),
	}
}

// update the prompt and reset the timer
func (l *LLMInterface) UpdatePrompt(prompt string) {
	logger.Infof("UPDATING LLM PROMPT %s", prompt)
	l.Lock()
	defer l.Unlock()

	if l.prompt != "" {
		l.prompt += " "
	}

	l.prompt += prompt
	l.timer.Stop()
	l.timer.Reset(llmTime)
}

func (l *LLMInterface) Stop() {
	l.cancel <- 1
}

func (l *LLMInterface) Start() {
	for {
		// wait for the timer to fire for stop to be called
		select {
		case <-l.timer.C:
			l.tryCallLLM()
		case <-l.cancel:
			logger.Info("shutting down llm interface")
			return
		}
	}
}

func (l *LLMInterface) tryCallLLM() {
	l.Lock()

	// no prompt so wait again
	if l.prompt == "" {
		l.Unlock()
		return
	}

	currentPrompt := l.prompt
	l.prompt = ""

	l.Unlock()
	// run inference
	response, err := l.callLLM(currentPrompt)
	if err != nil {
		logger.Error(err, "error calling llm")
		return
	}

	err = l.ttsEngine.Generate(response)
	if err != nil {
		logger.Error(err, "error generating speech")
	}

}

func (l *LLMInterface) callLLM(prompt string) (string, error) {
	var (
		response string
		err      error
	)
	logger.Infof("CALLING LLM WITH PROMPT %s", prompt)

	payload, err := json.Marshal(LLMRequest{Prompt: prompt})
	if err != nil {
		logger.Error(err, "error marshaling to json")
		return response, err
	}

	resp, err := http.Post(l.llmUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(err, "error making llm request")
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "error reading body")
		return response, err
	}

	if resp.StatusCode == http.StatusOK {
		llmResp := LLMResponse{}
		err = json.Unmarshal(body, &llmResp)
		if err != nil {
			logger.Error(err, "error unmarshaling body")
			return response, err
		}

		return llmResp.Text, err
	} else {
		return response, fmt.Errorf("got bad response %d", resp.StatusCode)
	}

}
