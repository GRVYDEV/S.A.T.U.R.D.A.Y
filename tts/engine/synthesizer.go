package engine

type Synthesizer interface {
	Synthesize(text string) (AudioChunk, error)
}

type AudioChunk struct {
	Data  []float32
	Index int
	// SampleRate of the audio in Hz (ex: 48kHz = 48000)
	SampleRate int
	// ChannelCount of the audio (usually 1)
	ChannelCount int
}
