package engine

import "errors"

type EngineParams struct {
	OnTextChunk func(TextChunk)
	Generator   Generator
}

type Engine struct {
	onTextChunk func(TextChunk)
	generator   Generator
}

func New(params EngineParams) (*Engine, error) {
	if params.Generator == nil {
		return nil, errors.New("you must supply a Generator to create an engine")
	}

	return &Engine{
		generator:   params.Generator,
		onTextChunk: params.OnTextChunk,
	}, nil
}

func (e *Engine) OnTextChunk(fn func(TextChunk)) {
	e.onTextChunk = fn
}

// Generate will call the engine generator with the provided prompt
func (e *Engine) Generate(prompt string) error {
	chunk, err := e.generator.Generate(prompt)
	if err != nil {
		return err
	}

	if e.onTextChunk != nil {
		e.onTextChunk(chunk)
	}

	return nil
}
