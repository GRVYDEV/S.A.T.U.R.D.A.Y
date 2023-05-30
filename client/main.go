package main

import (
	"flag"
	"net/url"
	"os"

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

	sc := NewSaturdayClient(SaturdayConfig{Room: room, Url: url_scheme})

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}
