package engine

type Transcriber interface {
	Transcribe(audioData []float32) (Transcription, error)
}

type TranscriptionSegment struct {
	StartTimestamp uint32 `json:"startTimestamp"`
	EndTimestamp   uint32 `json:"endTimestamp"`
	Text           string `json:"text"`
}

type Transcription struct {
	From           uint32
	Transcriptions []TranscriptionSegment
}
