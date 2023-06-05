package internal

import "gopkg.in/hraban/opus.v2"

type OpusDecoder struct {
	dec *opus.Decoder
}

func NewOpusDecoder(sampleRate, channels int) (*OpusDecoder, error) {
	dec, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	return &OpusDecoder{
		dec: dec,
	}, nil
}

func (o *OpusDecoder) Decode(data []byte, buf []float32) (int, error) {
	return o.dec.DecodeFloat32(data, buf)
}
