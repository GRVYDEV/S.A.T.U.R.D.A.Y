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

		err, transcription := we.runInference(Timestamp + whisperSampleWindowMs)

		if err == nil {
			var buffer []TranscriptionSegment
			logger.Debugf("Got %d segments start %d", len(transcription.transcriptions), transcription.from)
			//foreach of these transcription.transcriptions if there is a segment that is not the last add it to a buffer
			for i, segment := range transcription.transcriptions {
				logger.Debugf("Segment %d %d- %d: %s", i, transcription.from+segment.startTimestamp, transcription.from+segment.endTimestamp, segment.text)
				// If the segment is not the last one, add it to the buffer
				if i != len(transcription.transcriptions)-1 {
					buffer = append(buffer, segment)
					we.transcriptionStream <- segment
					we.lastHandledTimestamp = transcription.from + segment.endTimestamp
				}
			}
			logger.Debugf("new endTimestamp: %d", we.lastHandledTimestamp)

		} else {
			logger.Error(err, "error running inference")
		}
	}
}

func (we *WhisperEngine) runInference(addedRecordingStartTime uint32) (error, Transcription) {
	var (
		whisperWinLen = len(we.whisperWindow)
		pcmWinLen     = len(we.pcmWindow)
	)

	// log.Printf("attempting to run inference:\n WHISPER WINDOW: %d\n PCM WINDOWN: %d", whisperWinLen, pcmWinLen)

	if whisperWinLen == whisperWindowSize {
		// we have a full window so we need to drop the oldest samples and append the newest ones
		we.whisperWindow = append(we.whisperWindow[pcmWinLen:], we.pcmWindow...)
		// empty the pcm window so we can add new samples
		we.pcmWindow = we.pcmWindow[:0]
	} else if whisperWinLen+pcmWinLen > whisperWindowSize {
		// this shouldn't happen hopefully...
		message := fmt.Sprintf("GOING TO OVERFLOW WIN BUF BY %d", whisperWinLen+pcmWinLen-whisperWindowSize)
		return errors.New(message), Transcription{}
	} else if whisperWinLen+pcmWinLen < whisperWindowMinSize {
		// we dont have enough audio to run inference so add the pcmWindow and return
		message := fmt.Sprintf("not enough audio we only have %d samples continuing...", whisperWinLen)
		we.whisperWindow = append(we.whisperWindow, we.pcmWindow...)
		we.pcmWindow = we.pcmWindow[:0]
		return errors.New(message), Transcription{}
	} else {
		// we have enough audio to run inference
		we.whisperWindow = append(we.whisperWindow, we.pcmWindow...)
		we.pcmWindow = we.pcmWindow[:0]
	}

	whisperWindowStartTimestamp := addedRecordingStartTime - uint32(len(we.whisperWindow)/whisperSampleRateMs)
	timestampTranscriptionStartsFrom := whisperWindowStartTimestamp

	if we.lastHandledTimestamp > 0 && we.lastHandledTimestamp > whisperWindowStartTimestamp {
		delteToCutFromTheStart := (we.lastHandledTimestamp - whisperWindowStartTimestamp) * whisperSampleRateMs
		we.whisperWindow = we.whisperWindow[delteToCutFromTheStart:]
		timestampTranscriptionStartsFrom = we.lastHandledTimestamp
	}

	logger.Debugf("running whisper inference with %d window length", len(we.whisperWindow))
	return we.model.Process(we.whisperWindow, timestampTranscriptionStartsFrom)

}
