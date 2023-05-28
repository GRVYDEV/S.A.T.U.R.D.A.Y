package main

import (
	"errors"
	"log"
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
	transcriptions []TrasncriptionSegment
}

type TrasncriptionSegment struct {
	from uint32
	to   uint32
	text string
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

	log.Printf("Initialized whisper model with params:\n %s", params.String())

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
		log.Printf("Got %d segments start %d", segments, recordingStartTime)
		for i := 0; i < segments; i++ {
			trasncriptionSegment := TrasncriptionSegment{}

			trasncriptionSegment.from = uint32(w.ctx.Whisper_full_get_segment_t0(i) * 10)
			trasncriptionSegment.to = uint32(w.ctx.Whisper_full_get_segment_t1(i) * 10)

			trasncriptionSegment.text = w.ctx.Whisper_full_get_segment_text(i)

			log.Printf("Segment %d %d- %d: %s", i, recordingStartTime+trasncriptionSegment.from, recordingStartTime+trasncriptionSegment.to, trasncriptionSegment.text)
			transcription.transcriptions = append(transcription.transcriptions, trasncriptionSegment)
		}
	}
	elapsed := time.Since(start)
	log.Printf("Process took %s", elapsed)
	return nil, transcription
}
