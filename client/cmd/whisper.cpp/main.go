package main

import (
	"flag"
	"net/url"
	"os"

	"S.A.T.U.R.D.A.Y/client"
	logr "S.A.T.U.R.D.A.Y/log"
	whisper "S.A.T.U.R.D.A.Y/stt/backends/whisper.cpp"
	"S.A.T.U.R.D.A.Y/stt/engine"
	stt "S.A.T.U.R.D.A.Y/stt/engine"
	shttp "S.A.T.U.R.D.A.Y/tts/backends/http"
	tts "S.A.T.U.R.D.A.Y/tts/engine"

	"golang.org/x/exp/slog"
)

var debug = flag.Bool("debug", false, "print debug logs")

var (
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

	onDocumentUpdate := func(document engine.Document) {
		// TODO move this to document composer
		// FIXME this is horrible. We need to figure out how to fix the whisper segmenting logic
		// maybe look into seeding the context
		// if segment.Text[0] != '(' && segment.Text[0] != '[' && segment.Text[0] != '.' {
		// 	transcriptionStream <- segment
		// }
		transcriptionStream <- document
	}

	synthesizer, err := shttp.New("http://localhost:8000/synthesize")
	if err != nil {
		logger.Fatal(err, "error creating http")
	}

	ttsEngine, err := tts.New(tts.EngineParams{
		Synthesizer: synthesizer,
	})

	sttEngine, err := stt.New(stt.EngineParams{
		Transcriber:      whisperCpp,
		OnDocumentUpdate: onDocumentUpdate,
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
