package internal

import (
	"encoding/binary"
	"os"
	"testing"

	"S.A.T.U.R.D.A.Y/util"
)

func TestEncDec(t *testing.T) {
	data, err := os.ReadFile("./audio.pcm")
	if err != nil {
		t.Fatalf("error reading file %+v", err)
	}

	floats := util.BinaryToFloat32(data)

	floats = util.ConvertToDualChannel(floats)

	// 2 channel 20ms framesize
	enc, err := NewOpusEncoder(2, 20)
	if err != nil {
		t.Fatalf("error creating opus enc %+v", err)
	}

	// encode to 48khz from 22.05khz
	opusFrames, err := enc.Encode(floats, 2, 22050)
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

	for i, frame := range opusFrames {
		buf := make([]float32, frameSize)
		n, err := dec.Decode(frame.Data, buf)
		if err != nil {
			t.Fatalf("error decoding opus frame %+v", err)
		}

		if n != frameSize {
			t.Fatalf("WEIRD FRAME SIZE at %d got: %d expected: %d", i, n, frameSize)
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
