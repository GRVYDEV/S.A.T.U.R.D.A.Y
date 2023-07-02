package main

import (
	"flag"
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
	thttp "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/ttt/backends/http"
	ttt "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/ttt/engine"

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

	documentComposer := stt.NewDocumentComposer()
	documentComposer.FilterSegment(func(ts stt.TranscriptionSegment) bool {
		return ts.Text[0] == '.' || strings.ContainsAny(ts.Text, "[]()")
	})

	sttEngine, err := stt.New(stt.EngineParams{
		Transcriber:      whisperCpp,
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

	generator, err := thttp.New("http://localhost:9090/eva")
	if err != nil {
		logger.Fatal(err, "error creating http")
	}

	onTextChunk := func(chunk ttt.TextChunk) {
		err = ttsEngine.Generate(chunk.Text)
		if err != nil {
			logger.Error(err, "error generating speech")
		}
	}

	tttEngine, err := ttt.New(ttt.EngineParams{
		Generator:   generator,
		OnTextChunk: onTextChunk,
	})
	if err != nil {
		logger.Fatal(err, "error creating tttEngine")
	}

	pauseFunc := func() {
		sc.PauseTTS()
	}

	unpauseFunc := func() {
		sc.UnpauseTTS()
	}

	promptBuilder := NewPromptBuilder(llmTime, tttEngine, pauseFunc, unpauseFunc)

	onDocumentUpdate := func(document engine.Document) {
		transcriptionStream <- document
		promptBuilder.UpdatePrompt(document.NewText)
	}

	sttEngine.OnDocumentUpdate(onDocumentUpdate)

	go promptBuilder.Start()
	defer promptBuilder.Stop()

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}

// TODO eventually make this a general tool
// LLMInterface will call an llm with the specified prompt every 3 seconds and
// turn the response into audio
type PromptBuilder struct {
	tttEngine *ttt.Engine
	timer     *time.Timer
	prompt    string
	cancel    chan int

	// callback to pause tts inference
	pauseFunc func()
	// callback to unpause tts inference
	unpauseFunc func()

	sync.Mutex
}

func NewPromptBuilder(interval time.Duration, engine *ttt.Engine, pauseFunc func(), unpauseFunc func()) *PromptBuilder {
	return &PromptBuilder{
		tttEngine:   engine,
		timer:       time.NewTimer(interval),
		prompt:      "",
		cancel:      make(chan int),
		pauseFunc:   pauseFunc,
		unpauseFunc: unpauseFunc,
	}
}

// update the prompt and reset the timer
func (p *PromptBuilder) UpdatePrompt(prompt string) {
	logger.Infof("UPDATING LLM PROMPT %s", prompt)
	p.Lock()
	defer p.Unlock()

	if p.prompt != "" {
		p.prompt += " "
	}

	p.prompt += prompt
	p.timer.Stop()
	p.timer.Reset(llmTime)
}

func (p *PromptBuilder) Stop() {
	p.cancel <- 1
}

func (p *PromptBuilder) Start() {
	for {
		// wait for the timer to fire for stop to be called
		select {
		case <-p.timer.C:
			p.tryCallEngine()
		case <-p.cancel:
			logger.Info("shutting down llm interface")
			return
		}
	}
}

func (p *PromptBuilder) tryCallEngine() {
	p.Lock()

	// no prompt so wait again
	if p.prompt == "" {
		p.Unlock()
		return
	}

	currentPrompt := p.prompt
	p.prompt = ""

	p.Unlock()
	// pause TTS inference so we dont interrupt the tts
	p.pauseFunc()
	// run inference
	err := p.tttEngine.Generate(currentPrompt)
	if err != nil {
		logger.Error(err, "error calling tttEngine")
		// unpause TTS inference so we dont interrupt the tts
		p.unpauseFunc()
		return
	}

}
