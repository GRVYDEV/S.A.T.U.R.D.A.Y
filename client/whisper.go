package main

import (
	"errors"
	"math"
	"runtime"
	"time"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

type WhisperModel struct {
	ctx    *whisper.Context
	params whisper.Params
}

type Transcription struct {
	from           uint32
	transcriptions []TranscriptionSegment
}

type TranscriptionSegment struct {
	startTimestamp uint32 `json:"start_timestamp"`
	endTimestamp   uint32 `json:"end_timestamp"`
	text           string `json:"text"`
}

func NewWhisperModel() (*WhisperModel, error) {
	ctx := whisper.Whisper_init("./models/ggml-base.en.bin")
	if ctx == nil {
		return nil, errors.New("failed to initialize whisper")
	}

	params := ctx.Whisper_full_default_params(whisper.SAMPLING_GREEDY)
	params.SetPrintProgress(false)
	params.SetPrintSpecial(false)
	params.SetPrintRealtime(false)
	params.SetPrintTimestamps(false)
	params.SetSingleSegment(false)
	params.SetMaxTokensPerSegment(32)
	params.SetThreads(int(math.Min(float64(4), float64(runtime.NumCPU()))))
	params.SetSpeedup(false)
	params.SetLanguage(ctx.Whisper_lang_id("en"))

	logger.Infof("Initialized whisper model with params:\n %s", params.String())

	return &WhisperModel{ctx: ctx, params: params}, nil
}

func (w *WhisperModel) Process(samples []float32, recordingStartTime uint32) (error, Transcription) {
	start := time.Now()
	transcription := Transcription{}
	transcription.from = recordingStartTime
	if err := w.ctx.Whisper_full(w.params, samples, nil, nil); err != nil {
		return err, transcription
	} else {
		segments := w.ctx.Whisper_full_n_segments()
		for i := 0; i < segments; i++ {
			trasncriptionSegment := TranscriptionSegment{}

			trasncriptionSegment.startTimestamp = uint32(w.ctx.Whisper_full_get_segment_t0(i) * 10)
			trasncriptionSegment.endTimestamp = uint32(w.ctx.Whisper_full_get_segment_t1(i) * 10)

			trasncriptionSegment.text = w.ctx.Whisper_full_get_segment_text(i)

			transcription.transcriptions = append(transcription.transcriptions, trasncriptionSegment)
		}
	}
	elapsed := time.Since(start)
	logger.Debugf("Process took %s", elapsed)
	return nil, transcription
}
