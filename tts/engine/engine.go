package engine

import (
	"errors"
	"strings"
)

type EngineParams struct {
	OnAudioChunk func(AudioChunk)
	Synthesizer  Synthesizer
}

type Engine struct {
	onAudioChunk func(AudioChunk)
	synthesizer  Synthesizer
}

func New(params EngineParams) (*Engine, error) {
	if params.Synthesizer == nil {
		return nil, errors.New("you must supply a Synthesizer to create an engine")
	}

	return &Engine{
		synthesizer:  params.Synthesizer,
		onAudioChunk: params.OnAudioChunk,
	}, nil
}

func (e *Engine) OnAudioChunk(fn func(AudioChunk)) {
	e.onAudioChunk = fn
}

// Generate will chunk the text into sections and call the synthesize function on the Synthesizer for each
// segment. It will call onAudioChunk for each returned chunk from the synthesizer
func (e *Engine) Generate(text string) error {
	// FIXME we want to chunk into small sentences to parallelize audio generation
	// chunks := chunkText(text)
	// for i, textChunk := range chunks {
	// 	// TODO parallelize
	// 	chunk, err := e.synthesizer.Synthesize(textChunk)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if e.onAudioChunk != nil {
	// 		chunk.Index = i
	// 		e.onAudioChunk(chunk)
	// 	}
	// }

	chunk, err := e.synthesizer.Synthesize(text)
	if err != nil {
		return err
	}

	if e.onAudioChunk != nil {
		chunk.Index = 0
		// FIXME right now the audio sounds really bad at the beginning if we dont pad with some silence
		// I need to dig into why this happens
		data := make([]float32, 4800)
		data = append(data, chunk.Data...)
		chunk.Data = data
		e.onAudioChunk(chunk)
	}
	return nil
}

func chunkText(text string) []string {
	// TODO make this smarter
	return strings.Split(text, ".")
}
