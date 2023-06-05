package internal

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
)

func binaryToFloat32(data []byte) []float32 {
	floats := make([]float32, 0, len(data)/4)
	for {
		if len(data) == 0 {
			break
		}

		u := binary.LittleEndian.Uint32(data[:4])
		f := math.Float32frombits(u)
		floats = append(floats, f)
		data = data[4:]
	}
	return floats
}

func TestEncDec(t *testing.T) {
	data, err := os.ReadFile("./audio.pcm")
	if err != nil {
		t.Fatalf("error reading file %+v", err)
	}

	floats := binaryToFloat32(data)

	// 1 channel 20ms framesize
	enc, err := NewOpusEncoder(1, 20)
	if err != nil {
		t.Fatalf("error creating opus enc %+v", err)
	}

	// encode to 48khz from 22.05khz
	opusFrames, err := enc.Encode(floats, 22050)
	if err != nil {
		t.Fatalf("error encoding opus frames %+v", err)
	}

	// decode back to 48khz
	dec, err := NewOpusDecoder(48000, 1)
	if err != nil {
		t.Fatalf("error creating opus decoder %+v", err)
	}

	frameSize := 1 * 20 * 48000 / 1000

	rawOut := make([]float32, 0, len(opusFrames)*frameSize)

	for _, frame := range opusFrames {
		buf := make([]float32, frameSize)
		n, err := dec.Decode(frame.data, buf)
		if err != nil {
			t.Fatalf("error decoding opus frame %+v", err)
		}

		if n != frameSize {
			t.Fatalf("WEIRD FRAME SIZE got: %d expected: %d", n, frameSize)
		}

		rawOut = append(rawOut, buf...)
	}

	f, err := os.OpenFile("audio_test.pcm", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = binary.Write(f, binary.LittleEndian, rawOut)
	if err != nil {
		panic(err)
	}

}
