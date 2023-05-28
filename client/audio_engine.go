package main

import (
	"log"
	"math"

	"github.com/pion/rtp"
	"gopkg.in/hraban/opus.v2"
)

const (
	sampleRate  = whisperSampleRate // found in whisper_engine.go (16000)
	channels    = 1                 // decode into 1 channel since that is what whisper.cpp wants
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
	pcm []float32
	// slice to hold binary encoded pcm data
	buf []byte

	firstTimeStamp *uint32

	we *WhisperEngine
}

func NewAudioEngine() (*AudioEngine, error) {
	dec, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	we, err := NewWhisperEngine()
	if err != nil {
		return nil, err
	}

	return &AudioEngine{
		rtpIn:  make(chan *rtp.Packet),
		rtpOut: make(chan *rtp.Packet),
		pcm:    make([]float32, frameSize),
		buf:    make([]byte, frameSize*2),
		dec:    dec,
		we:     we,
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
	// _, err := os.Create("audio.pcm")
	// if err != nil {
	// 	log.Printf("err creating file %+v", err)
	// 	return
	// }

	for {
		pkt, ok := <-a.rtpIn
		if !ok {
			log.Print("rtpIn channel closed...")
			return
		}
		// log.Printf("got pkt of size %d", len(pkt.Payload))
		if pkt.SequenceNumber == 0 {
			log.Print("Resetting timestamp bc sequencenumber 0...")
			a.firstTimeStamp = &pkt.Timestamp
		}
		if a.firstTimeStamp == nil {
			log.Print("Resetting timestamp bc firstTimeStamp is nil...  ", pkt.Timestamp)
			a.firstTimeStamp = &pkt.Timestamp
		}

		if _, err := a.decodePacket(pkt); err != nil {
			log.Fatalf("error decoding opus packet %+v", err)
		} else {

			// log.Printf("decoded %d bytes", n)
			// if _, err = f.Write(a.buf[:n]); err != nil {
			// 	log.Fatalf("error writing to file %+v", err)
			// }
		}
	}
}

func (a *AudioEngine) decodePacket(pkt *rtp.Packet) (int, error) {
	_, err := a.dec.DecodeFloat32(pkt.Payload, a.pcm)
	// we decode to float32 here since that is what whisper.cpp takes
	if err != nil {
		log.Printf("error decoding fb packet %+v", err)
		return 0, err
	} else {
		// log.Printf("decoded %d FB samples", n)
		// log.Printf("decoded %d samples", len(a.pcm))
		// log.Printf("timestamps %d %d", pkt.Timestamp, *a.firstTimeStamp)
		// log.Printf("Calc %d", (pkt.Timestamp-(*a.firstTimeStamp))/(sampleRate*3))
		timestampMS := (pkt.Timestamp - (*a.firstTimeStamp)) / ((sampleRate / 1000) * 3)
		lengthOfRecording := uint32(len(a.pcm)) * 3
		timestampRecordingEnds := timestampMS + lengthOfRecording
		a.we.Write(a.pcm, timestampRecordingEnds)
		return convertToBytes(a.pcm, a.buf), nil
	}
}

// This function converts f32le to s16le bytes for writing to a file
func convertToBytes(in []float32, out []byte) int {
	currIndex := 0
	for i := range in {
		res := int16(math.Floor(float64(in[i] * 32767)))

		out[currIndex] = byte(res & 0b11111111)
		currIndex++

		out[currIndex] = (byte(res >> 8))
		currIndex++
	}
	return currIndex
}
