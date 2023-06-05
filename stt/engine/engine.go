package engine

import (
	"errors"
	"fmt"
	"sync"

	logr "S.A.T.U.R.D.A.Y/log"
)

// FIXME make these configurable
const (
	// This is determined by the hyperparameter configuration that whisper was trained on.
	// See more here: https://github.com/ggerganov/whisper.cpp/issues/909
	SampleRate   = 16000 // 16kHz
	sampleRateMs = SampleRate / 1000
	// This determines how much audio we will be passing to whisper inference.
	// We will buffer up to (whisperSampleWindowMs - pcmSampleRateMs) of old audio and then add
	// audioSampleRateMs of new audio onto the end of the buffer for inference
	sampleWindowMs = 24000 // 24 second sample window
	windowSize     = sampleWindowMs * sampleRateMs
	// This is the minimum ammount of audio we want to buffer before running inference
	// 2 seconds of audio samples
	windowMinSize = 2000 * sampleRateMs
	// This determines how often we will try to run inference.
	// We will buffer (pcmSampleRateMs * whisperSampleRate / 1000) samples and then run inference
	pcmSampleRateMs = 500
	pcmWindowSize   = pcmSampleRateMs * sampleRateMs
)

var Logger = logr.New()

type EngineParams struct {
	OnDocumentUpdate func(Document)
	Transcriber      Transcriber
	DocumentComposer *DocumentComposer
}

type Engine struct {
	sync.Mutex
	// Buffer to store new audio. When this fills up we will try to run inference
	pcmWindow []float32
	// Buffer to store old and new audio to run inference on.
	// By inferring on old and new audio we can help smooth out cross word boundaries
	window               []float32
	lastHandledTimestamp uint32

	// document composer to handle incoming transcriptions
	documentComposer *DocumentComposer

	// callback when we have a document update
	onDocumentUpdate func(Document)

	transcriber Transcriber
}

func New(params EngineParams) (*Engine, error) {
	if params.Transcriber == nil {
		return nil, errors.New("you must supply a Transciber to create an engine")
	}

	if params.DocumentComposer == nil {
		params.DocumentComposer = NewDocumentComposer()
	}

	return &Engine{
		window:               make([]float32, 0, windowSize),
		pcmWindow:            make([]float32, 0, pcmWindowSize),
		lastHandledTimestamp: 0,
		onDocumentUpdate:     params.OnDocumentUpdate,
		transcriber:          params.Transcriber,
		documentComposer:     NewDocumentComposer(),
	}, nil
}

func (e *Engine) OnDocumentUpdate(fn func(Document)) {
	e.onDocumentUpdate = fn
}

// endTimestamp is the latest packet timestamp + len of the audio in the packet
func (e *Engine) Write(pcm []float32, Timestamp uint32) {
	e.Lock()
	defer e.Unlock()
	if len(e.pcmWindow)+len(pcm) > pcmWindowSize {
		// This shouldn't happen hopefully...
		Logger.Infof("GOING TO OVERFLOW PCM WINDOW BY %d", len(e.pcmWindow)+len(pcm)-pcmWindowSize)
	}
	e.pcmWindow = append(e.pcmWindow, pcm...)
	if len(e.pcmWindow) >= pcmWindowSize {
		// TODO make this run in a go routine
		currentTime := Timestamp + sampleWindowMs
		transcription, err := e.runInference(currentTime)

		if err == nil {
			document, timestamp := e.documentComposer.NewTranscript(transcription)

			if e.onDocumentUpdate != nil {
				e.onDocumentUpdate(document)
			}

			transcriptLen := timestamp - e.lastHandledTimestamp
			if timestamp > e.lastHandledTimestamp {
				windowDelta := transcriptLen * sampleRateMs
				if windowDelta > uint32(len(e.window)) {
					windowDelta = uint32(len(e.window))
				}
				e.window = e.window[windowDelta:]
			}
			e.lastHandledTimestamp = timestamp
		} else {
			Logger.Error(err, "error running inference")
		}
	}
}

// endTimestamp is the latest packet timestamp + len of the audio in the packet
func (e *Engine) runInference(endTimestamp uint32) (Transcription, error) {
	var (
		whisperWinLen = len(e.window)
		pcmWinLen     = len(e.pcmWindow)
	)

	if whisperWinLen == windowSize || whisperWinLen+pcmWinLen > windowSize {
		// we have a full window or we might overflow
		// we need to drop the oldest samples and append the newest ones
		e.window = append(e.window[pcmWinLen:], e.pcmWindow...)
		// we also need to increment the last handled timestamp by the number of samples we slid the window
		e.lastHandledTimestamp += uint32(pcmWinLen) / sampleRateMs
		// empty the pcm window so we can add new samples
		e.pcmWindow = e.pcmWindow[:0]
	} else if whisperWinLen+pcmWinLen < windowMinSize {
		// we dont have enough audio to run inference so add the pcmWindow and return
		message := fmt.Sprintf("not enough audio we only have %d samples continuing...", whisperWinLen)
		e.window = append(e.window, e.pcmWindow...)
		e.pcmWindow = e.pcmWindow[:0]
		return Transcription{}, errors.New(message)
	} else {
		// we have enough audio to run inference
		e.window = append(e.window, e.pcmWindow...)
		e.pcmWindow = e.pcmWindow[:0]
	}

	Logger.Debugf("running whisper inference with %d window length", len(e.window))

	transcript, err := e.transcriber.Transcribe(e.window)
	transcript.From = e.lastHandledTimestamp
	return transcript, err

}
