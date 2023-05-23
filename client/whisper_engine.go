package main

import "sync"

type WhisperEngine struct {
	sync.Mutex
	buf   []float32
	count int
}

func NewWhisperEngine() *WhisperEngine {
	return &WhisperEngine{
		buf:   make([]float32, 32000),
		count: 0,
	}
}

func (we *WhisperEngine) Write(pcm []float32) {
	we.Lock()
	defer we.Unlock()
	we.buf = append(we.buf, pcm...)
}
