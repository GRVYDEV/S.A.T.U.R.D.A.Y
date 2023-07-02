package engine

type Generator interface {
	Generate(text string) (TextChunk, error)
}

type TextChunk struct {
	Text string
}
