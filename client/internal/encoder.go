package internal

import (
	logr "S.A.T.U.R.D.A.Y/log"

	"gopkg.in/hraban/opus.v2"
)

const opusSampleRate = 48000

// PcmFrame is used for chunking raw pcm input into frames for the opus encoder
type PcmFrame struct {
	data  []float32
	index int
}

// OpusFrame contains and encoded opus frame
type OpusFrame struct {
	data  []byte
	index int
}

type OpusEncoder struct {
	enc         *opus.Encoder
	channels    int
	sampleRate  int
	frameSizeMs int
}

var Logger = logr.New()

func NewOpusEncoder(channels, frameSizeMs int) (*OpusEncoder, error) {
	enc, err := opus.NewEncoder(opusSampleRate, channels, opus.AppRestrictedLowdelay)
	if err != nil {
		return nil, err
	}

	return &OpusEncoder{
		enc:         enc,
		channels:    channels,
		frameSizeMs: frameSizeMs,
	}, nil
}

// Encode will resample and encode the provided pcm audio to 48khz Opus
func (o *OpusEncoder) Encode(pcm []float32, inputSampleRate int) ([]OpusFrame, error) {
	if inputSampleRate != opusSampleRate {
		pcm = Resample(pcm, inputSampleRate, opusSampleRate)
	}
	frames := o.chunkPcm(pcm, opusSampleRate)

	opusFrames := make([]OpusFrame, 0, len(frames))

	for _, frame := range frames {
		opusFrame, err := o.encodeToOpus(frame)
		if err != nil {
			Logger.Error(err, "error encoding opus frame")
			return opusFrames, err
		}

		opusFrames = append(opusFrames, opusFrame)
	}

	Logger.Infof("encoded %d opus frames", len(opusFrames))

	return opusFrames, nil

}

func (o *OpusEncoder) encodeToOpus(frame PcmFrame) (OpusFrame, error) {
	opusFrame := OpusFrame{index: frame.index}
	data := make([]byte, 1000)

	Logger.Infof("encoding pcm size %d", len(frame.data))

	n, err := o.enc.EncodeFloat32(frame.data, data)
	if err != nil {
		Logger.Errorf(err, "error encoding frame %+v", err)
		return opusFrame, err
	}
	opusFrame.data = data[:n]

	return opusFrame, nil
}

// chunkPcm will split the provided pcm audio into properly sized frames
func (o *OpusEncoder) chunkPcm(pcm []float32, sampleRate int) []PcmFrame {
	// the amount of samples that fit into a frame
	outputFrameSize := o.channels * o.frameSizeMs * sampleRate / 1000
	// TODO make sure this rounds up
	totalFrames := len(pcm) / outputFrameSize

	frames := make([]PcmFrame, 0, totalFrames)

	idx := 0
	for idx < totalFrames {
		pcmLen := len(pcm)
		// we have at least a full frame left
		if pcmLen > outputFrameSize {
			Logger.Debug("Got a full frame")
			frames = append(frames, PcmFrame{index: idx, data: pcm[:outputFrameSize]})
			// chop frame off of input
			pcm = pcm[outputFrameSize:]
			idx++
		} else {
			// we have less than a full frame so lets pad with silence
			sampleDelta := outputFrameSize - pcmLen
			silence := make([]float32, sampleDelta)

			Logger.Debugf("Got a partial frame len %d padding with %d silence samples", pcmLen, len(silence))

			frames = append(frames, PcmFrame{index: idx, data: append(pcm, silence...)})
			break
		}
	}

	Logger.Debugf("got %d frames", len(frames))

	return frames
}
