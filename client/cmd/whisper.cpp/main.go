package main

import (
	"flag"
	"net/url"
	"os"

	"S.A.T.U.R.D.A.Y/client"
	logr "S.A.T.U.R.D.A.Y/log"
	whisper "S.A.T.U.R.D.A.Y/stt/backends/whisper.cpp"
	"S.A.T.U.R.D.A.Y/stt/engine"

	"github.com/rs/zerolog"
)

var debug = flag.Bool("debug", false, "print debug logs")

var (
	logger = logr.New()
)

func main() {
	flag.Parse()
	if !*debug {
		logr.SetGlobalOptions(logr.GlobalConfig{V: int(zerolog.DebugLevel)})
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

	transcriptionStream := make(chan engine.TranscriptionSegment, 100)

	onTranscriptionSegment := func(segment engine.TranscriptionSegment) {
		// FIXME this is horrible. We need to figure out how to fix the whisper segmenting logic
		// maybe look into seeding the context
		if segment.Text[0] != '(' && segment.Text[0] != '[' && segment.Text[0] != '.' {
			transcriptionStream <- segment
		}
	}

	engine, err := engine.New(engine.EngineParams{
		Transcriber:            whisperCpp,
		OnTranscriptionSegment: onTranscriptionSegment,
	})

	sc, err := client.NewSaturdayClient(client.SaturdayConfig{Room: room, Url: url, SttEngine: engine})
	if err != nil {
		logger.Fatal(err, "error creating saturday client")
	}

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}
