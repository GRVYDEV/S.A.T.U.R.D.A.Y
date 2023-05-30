package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type TranscriptionSegment struct {
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Text           string `json:"text"`
}

type Transcription struct {
	Transcriptions []TranscriptionSegment `json:"transcriptions"`
}

func main() {
	router := gin.Default()
	whisperEngine, err := NewWhisperModel()
	if err != nil {
		log.Fatalf("failed to create whisper engine %+v", err)
	}
	router.POST("/transcribe", func(c *gin.Context) {
		var transcriptionRequest []float32

		if err := c.ShouldBindJSON(&transcriptionRequest); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		start := time.Now()
		err, transcription := whisperEngine.Process(transcriptionRequest)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		end := time.Now()

		elapsed := end.Sub(start)
		log.Printf("Took: %v", elapsed)
		log.Printf("%v", transcription)

		c.JSON(200, transcription)
	})

	router.Run(":8000")
}
