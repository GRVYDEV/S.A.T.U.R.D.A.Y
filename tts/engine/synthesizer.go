package engine

type Synthesizer interface {
	Synthesize(text string) ([]float32, error)
}

type AudioChunk struct {
	data  []float32
	index int
}
