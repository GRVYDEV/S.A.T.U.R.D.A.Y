package main

import (
	"time"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	whisper "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/backends/whisper.cpp"

	"github.com/gin-gonic/gin"
)

var (
	logger = logr.New()
)

func main() {
	whisperEngine, err := whisper.New("../models/ggml-base.en.bin")
	if err != nil {
		logger.Fatal(err, "error creating whisper model")
	}

	router := gin.Default()
	router.POST("/transcribe", func(c *gin.Context) {
		var transcriptionRequest []float32

		if err := c.ShouldBindJSON(&transcriptionRequest); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		start := time.Now()
		transcription, err := whisperEngine.Transcribe(transcriptionRequest)
		logger.Error(err, "error running inference")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		end := time.Now()

		elapsed := end.Sub(start)
		logger.Infof("Took: %v", elapsed)
		logger.Infof("%v", transcription)

		c.JSON(200, transcription)
	})

	router.Run(":8000")
}
