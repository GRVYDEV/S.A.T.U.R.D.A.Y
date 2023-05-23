package main

import (
	"log"
	"net/url"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8088", Path: "/ws"}

	sc := NewSaturdayClient(SaturdayConfig{Room: "test", Url: u})

	log.Print("Starting Saturday Client...")

	if err := sc.Start(); err != nil {
		log.Fatalf("error with Saturday Client %+v", err)
	}
}
