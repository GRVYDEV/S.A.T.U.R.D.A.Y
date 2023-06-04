package main

import (
	"flag"
	"net/url"
	"os"

	"S.A.T.U.R.D.A.Y/client"
	shttp "S.A.T.U.R.D.A.Y/stt/backends/http"
	"S.A.T.U.R.D.A.Y/stt/engine"

	logr "S.A.T.U.R.D.A.Y/log"
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

	url_env := os.Getenv("URL")
	if url_env == "" {
		url_env = "localhost:8088"
	}

	room := os.Getenv("ROOM")
	if room == "" {
		room = "test"
	}

	url_scheme := url.URL{Scheme: "ws", Host: url_env, Path: "/ws"}

	transcriptionService := os.Getenv("TRASCRIPTION_SERVICE")
	if transcriptionService == "" {
		transcriptionService = "http://localhost:8000/"
	}
	transcriptionUrl := transcriptionService + room + "/transcribe" // Replace with the appropriate API URL

	httpApi, err := shttp.New(transcriptionUrl)
	if err != nil {
		logger.Fatal(err, "error creating http api")
	}

	transcriptionStream := make(chan *engine.Document, 100)

	onTranscriptionSegment := func(segment *engine.Document) {
		logger.Debug(segment.NewText)
		transcriptionStream <- segment
	}

	engine, err := engine.New(engine.EngineParams{
		Transcriber:            httpApi,
		OnTranscriptionSegment: onTranscriptionSegment,
	})

	sc, err := client.NewSaturdayClient(client.SaturdayConfig{
		Room:                room,
		Url:                 url_scheme,
		SttEngine:           engine,
		TranscriptionStream: transcriptionStream})
	if err != nil {
		logger.Fatal(err, "error creating saturday client")
	}

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}
