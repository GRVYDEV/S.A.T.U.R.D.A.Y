package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	// This is determined by the hyperparameter configuration that whisper was trained on.
	// See more here: https://github.com/ggerganov/whisper.cpp/issues/909
	whisperSampleRate   = 16000 // 16kHz
	whisperSampleRateMs = whisperSampleRate / 1000
	// This determines how much audio we will be passing to whisper inference.
	// We will buffer up to (whisperSampleWindowMs - pcmSampleRateMs) of old audio and then add
	// audioSampleRateMs of new audio onto the end of the buffer for inference
	whisperSampleWindowMs = 24000 // 24 second sample window
	whisperWindowSize     = whisperSampleWindowMs * whisperSampleRateMs
	// This is the minimum ammount of audio we want to buffer before running inference
	// 2 seconds of audio samples
	whisperWindowMinSize = 2000 * whisperSampleRateMs
	// This determines how often we will try to run inference.
	// We will buffer (pcmSampleRateMs * whisperSampleRate / 1000) samples and then run inference
	pcmSampleRateMs = 500
	pcmWindowSize   = pcmSampleRateMs * whisperSampleRateMs
)

type WhisperEngine struct {
	sync.Mutex
	// Buffer to store new audio. When this fills up we will try to run inference
	pcmWindow []float32
	// Buffer to store old and new audio to run inference on.
	// By inferring on old and new audio we can help smooth out cross word boundaries
	whisperWindow        []float32
	model                *WhisperModel
	lastHandledTimestamp uint32
	transcriptionStream  chan TranscriptionSegment
}

func NewWhisperEngine(transcriptionStream chan TranscriptionSegment) (*WhisperEngine, error) {
	model, err := NewWhisperModel()
	if err != nil {
		return nil, err
	}

	return &WhisperEngine{
		whisperWindow:        make([]float32, 0, whisperWindowSize),
		pcmWindow:            make([]float32, 0, pcmWindowSize),
		model:                model,
		lastHandledTimestamp: 0,
		transcriptionStream:  transcriptionStream,
	}, nil
}

// TODO: We need to de-noise this audio before running inference
func (we *WhisperEngine) Write(pcm []float32, Timestamp uint32) {
	we.Lock()
	defer we.Unlock()
	if len(we.pcmWindow)+len(pcm) > pcmWindowSize {
		// This shouldn't happen hopefully...
		logger.Infof("GOING TO OVERFLOW PCM WINDOW BY %d", len(we.pcmWindow)+len(pcm)-pcmWindowSize)
	}
	we.pcmWindow = append(we.pcmWindow, pcm...)
	// We have filled up our window so lets run inference
	if len(we.pcmWindow) >= pcmWindowSize {
		// TODO make this run in a go routine

		transcription, err := we.runInference(Timestamp)

		if err == nil {
			logger.Infof("Got %d segments start %d", len(transcription.transcriptions), transcription.from)
			// if there is more than one segment then send all except the last, slide the whisper window and update the timestamp
			if len(transcription.transcriptions) > 1 {
				for i, segment := range transcription.transcriptions {
					// if this is the last one do nothing
					if i == len(transcription.transcriptions)-1 {
						break
					}
					// FIXME this is horrible. We need to figure out how to fix the whisper segmenting logic
					// maybe look into seeding the context
					if segment.Text[0] != '(' && segment.Text[0] != '[' && segment.Text[0] != '.' {
						// send on the chan
						we.transcriptionStream <- segment
					}
					// if this is the second to last one then update last handled timestamp and chop the window
					if i == len(transcription.transcriptions)-2 {
						transcriptEnd := segment.EndTimestamp
						transcriptLen := transcriptEnd - we.lastHandledTimestamp
						windowDelta := transcriptLen * whisperSampleRateMs

						we.whisperWindow = we.whisperWindow[windowDelta:]
						we.lastHandledTimestamp = transcriptEnd

						logger.Infof("new endTimestamp: %d", we.lastHandledTimestamp)
					}

				}
			}

		} else {
			logger.Error(err, "error running inference")
		}
	}
}

// endTimestamp is the latest packet timestamp + len of the audio in the packet
func (we *WhisperEngine) runInference(endTimestamp uint32) (Transcription, error) {
	var (
		whisperWinLen = len(we.whisperWindow)
		pcmWinLen     = len(we.pcmWindow)
	)

	if whisperWinLen == whisperWindowSize || whisperWinLen+pcmWinLen > whisperWindowSize {
		// we have a full window or we might overflow
		// we need to drop the oldest samples and append the newest ones
		we.whisperWindow = append(we.whisperWindow[pcmWinLen:], we.pcmWindow...)
		// we also need to increment the last handled timestamp by the number of samples we slid the window
		we.lastHandledTimestamp += uint32(pcmWinLen) * whisperSampleRateMs
		// empty the pcm window so we can add new samples
		we.pcmWindow = we.pcmWindow[:0]
	} else if whisperWinLen+pcmWinLen < whisperWindowMinSize {
		// we dont have enough audio to run inference so add the pcmWindow and return
		message := fmt.Sprintf("not enough audio we only have %d samples continuing...", whisperWinLen)
		we.whisperWindow = append(we.whisperWindow, we.pcmWindow...)
		we.pcmWindow = we.pcmWindow[:0]
		return Transcription{}, errors.New(message)
	} else {
		// we have enough audio to run inference
		we.whisperWindow = append(we.whisperWindow, we.pcmWindow...)
		we.pcmWindow = we.pcmWindow[:0]
	}

	logger.Debugf("running whisper inference with %d window length", len(we.whisperWindow))
	return we.model.Process(we.whisperWindow, we.lastHandledTimestamp)

}
