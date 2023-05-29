package main

import (
	"flag"
	"net/url"

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
	logger.Debug("hello")
	logger.Info("hello info")

	u := url.URL{Scheme: "ws", Host: "localhost:8088", Path: "/ws"}

	sc := NewSaturdayClient(SaturdayConfig{Room: "test", Url: u})

	logger.Info("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		logger.Fatal(err, "error starting Saturday Client")
	}
}
