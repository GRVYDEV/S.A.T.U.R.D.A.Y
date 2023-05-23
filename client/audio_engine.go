package main

import (
	"log"
	"os"

	"github.com/pion/rtp"
	nopus "gopkg.in/hraban/opus.v2"
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
	_, err := os.Create("audio.pcm")
	if err != nil {
		log.Printf("err creating file %+v", err)
		return
	}

	channels := 2
	frameSizeMs := 20
	sampleRate := 48000 // negotiated in SDP?

	frameSize := channels * frameSizeMs * sampleRate / 1000
	out := make([]int16, frameSize)
	//	decoder := opus.NewDecoder()

	dec, err := nopus.NewDecoder(sampleRate, channels)
	if err != nil {
		log.Fatalf("error creating opus decoder %+v", err)
	}

	for {
		pkt, ok := <-a.rtpIn
		if !ok {
			log.Print("rtpIn channel closed...")
			return
		}
		log.Printf("got pkt of size %d", len(pkt.Payload))
		if n, err := dec.Decode(pkt.Payload, out); err != nil {
			log.Fatalf("error decoding opus packet %+v", err)
		} else {
			log.Printf("decoded %d samples", n)
		}

		// if _, err = f.Write(out); err != nil {
		// 	log.Fatalf("error writing to file %+v", err)
		// }
	}
}
