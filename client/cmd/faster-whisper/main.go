package main

import (
	"flag"
	"net/url"
	"os"

	"S.A.T.U.R.D.A.Y/client"
	fwhisper "S.A.T.U.R.D.A.Y/stt/backends/faster-whisper"
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
		transcriptionService = "http://localhost:8000"
	}
	transcriptionUrl := room + "/transcribe" // Replace with the appropriate API URL

	fasterWhisper, err := fwhisper.New(transcriptionUrl)
	if err != nil {
		logger.Fatal(err, "error creating whisper api")
	}

	transcriptionStream := make(chan engine.TranscriptionSegment, 100)

	onTranscriptionSegment := func(segment engine.TranscriptionSegment) {
		transcriptionStream <- segment
	}

	engine, err := engine.New(engine.EngineParams{
		Transcriber:            fasterWhisper,
		OnTranscriptionSegment: onTranscriptionSegment,
	})

	sc, err := client.NewSaturdayClient(client.SaturdayConfig{Room: room, Url: url_scheme, SttEngine: engine})
	if err != nil {
		logger.Fatal(err, "error creating saturday client")
	}

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}
