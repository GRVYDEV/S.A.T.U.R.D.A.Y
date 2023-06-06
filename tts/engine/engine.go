package engine

type Engine struct {
	onAudioChunk func(AudioChunk)
	synthesizer  Synthesizer
}

// Generate will chunk the text into sections and call the synthesize function on the Synthesizer for each
// segment. It will call onAudioChunk for each returned chunk from the synthesizer
func (e *Engine) Generate(text string) {
	chunks := chunkText(text)
	for i, txt := range chunks {

	}
}

func chunkText(text string) []string {
	return []string{}
}
