package main

import (
	"log"
	"os"

	"github.com/pion/opus"
	"github.com/pion/rtp"
)

// AudioEngine is used to convert RTP Opus packets to raw PCM audio to be sent to Whisper
// and to convert raw PCM audio from Coqui back to RTP Opus packets to be sent back over WebRTC
type AudioEngine struct {
	// RTP Opus packets to be converted to PCM
	rtpIn chan *rtp.Packet
	// RTP Opus packets converted from PCM to be sent over WebRTC
	rtpOut chan *rtp.Packet
}

func NewAudioEngine() *AudioEngine {
	return &AudioEngine{
		rtpIn:  make(chan *rtp.Packet),
		rtpOut: make(chan *rtp.Packet),
	}
}

func (a *AudioEngine) In() chan<- *rtp.Packet {
	return a.rtpIn
}

func (a *AudioEngine) Out() <-chan *rtp.Packet {
	return a.rtpOut
}

func (a *AudioEngine) Start() {
	log.Print("Starting audio engine")
	go a.decode()
}

// decode reads over the in channel in a loop, decodes the RTP packets to raw PCM and sends the data on another channel
func (a *AudioEngine) decode() {
	f, err := os.Create("audio.pcm")
	if err != nil {
		log.Printf("err creating file %+v", err)
		return
	}

	out := make([]byte, 1920)
	decoder := opus.NewDecoder()

	for {
		pkt, ok := <-a.rtpIn
		if !ok {
			log.Print("rtpIn channel closed...")
			return
		}
		log.Printf("got pkt of size %d", len(pkt.Payload))
		if _, _, err := decoder.Decode(pkt.Payload, out); err != nil {
			log.Fatalf("error decoding opus packet %+v", err)
		}

		if _, err = f.Write(out); err != nil {
			log.Fatalf("error writing to file %+v", err)
		}
	}
}
