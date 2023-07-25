package whisper

import (
	"errors"
	"math"
	"runtime"
	"strings"
	"time"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

var Logger = logr.New()

// ensure this satisfies the interface
var _ engine.Transcriber = (*WhisperModel)(nil)

type WhisperModel struct {
	ctx    *whisper.Context
	params whisper.Params
}

// NewWhisperModel creates a new WhisperModel with the model spicified by modelPath
func New(modelPath string) (*WhisperModel, error) {
	ctx := whisper.Whisper_init(modelPath)
	if ctx == nil {
		return nil, errors.New("failed to initialize whisper")
	}

	// FIXME make a way to configure these
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

	Logger.Infof("Initialized whisper model with params:\n %s", params.String())

	return &WhisperModel{ctx: ctx, params: params}, nil
}

func (w *WhisperModel) Transcribe(samples []float32) (engine.Transcription, error) {
	start := time.Now()
	transcription := engine.Transcription{}
	if err := w.ctx.Whisper_full(w.params, samples, nil, nil, nil); err != nil {
		return transcription, err
	} else {
		segments := w.ctx.Whisper_full_n_segments()
		for i := 0; i < segments; i++ {
			trasncriptionSegment := engine.TranscriptionSegment{}

			trasncriptionSegment.StartTimestamp = uint32(w.ctx.Whisper_full_get_segment_t0(i) * 10)
			trasncriptionSegment.EndTimestamp = uint32(w.ctx.Whisper_full_get_segment_t1(i) * 10)

			trasncriptionSegment.Text = strings.TrimLeft(w.ctx.Whisper_full_get_segment_text(i), " ")

			transcription.Transcriptions = append(transcription.Transcriptions, trasncriptionSegment)
		}
	}
	elapsed := time.Since(start)
	Logger.Debugf("Process took %s", elapsed)
	return transcription, nil
}
