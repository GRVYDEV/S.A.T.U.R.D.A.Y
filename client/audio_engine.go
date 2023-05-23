package main

import (
	"log"
	"os"

	"github.com/pion/rtp"
	"gopkg.in/hraban/opus.v2"
)

const (
	sampleRate  = 48000
	channels    = 2
	frameSizeMs = 20
)

var frameSize = channels * frameSizeMs * sampleRate / 1000

// AudioEngine is used to convert RTP Opus packets to raw PCM audio to be sent to Whisper
// and to convert raw PCM audio from Coqui back to RTP Opus packets to be sent back over WebRTC
type AudioEngine struct {
	// RTP Opus packets to be converted to PCM
	rtpIn chan *rtp.Packet
	// RTP Opus packets converted from PCM to be sent over WebRTC
	rtpOut chan *rtp.Packet

	dec *opus.Decoder
	// slice to hold raw pcm data during decoding
	pcm []int16
	// slice to hold binary encoded pcm data
	buf []byte
}

func NewAudioEngine() (*AudioEngine, error) {
	dec, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	return &AudioEngine{
		rtpIn:  make(chan *rtp.Packet),
		rtpOut: make(chan *rtp.Packet),
		pcm:    make([]int16, frameSize),
		buf:    make([]byte, frameSize*2),
		dec:    dec,
	}, nil
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

	for {
		pkt, ok := <-a.rtpIn
		if !ok {
			log.Print("rtpIn channel closed...")
			return
		}
		log.Printf("got pkt of size %d", len(pkt.Payload))
		if n, err := a.decodePacket(pkt); err != nil {
			log.Fatalf("error decoding opus packet %+v", err)
		} else {
			log.Printf("decoded %d bytes", n)
			if _, err = f.Write(a.buf[:n]); err != nil {
				log.Fatalf("error writing to file %+v", err)
			}
		}

	}
}

func (a *AudioEngine) decodePacket(pkt *rtp.Packet) (int, error) {
	if n, err := a.dec.Decode(pkt.Payload, a.pcm); err != nil {
		log.Printf("error decoding fb packet %+v", err)
		return 0, err
	} else {
		log.Printf("decoded %d FB samples", n)
		return convertToBytes(a.pcm, a.buf), nil
	}
}

func convertToBytes(in []int16, out []byte) int {
	currIndex := 0
	for _, i := range in {
		out[currIndex] = byte(i & 0b11111111)
		currIndex++

		out[currIndex] = (byte(i >> 8))
		currIndex++
	}
	return currIndex
}
