WHISPER_DIR := $(abspath ./whisper.cpp)
MODEL_NAME := base.en
MODEL_FILE_NAME := ggml-$(MODEL_NAME).bin
MODELS_DIR := models

.PHONY: tts rtc client

clean:
	@rm $(WHISPER_DIR)/whisper.o
	@rm $(WHISPER_DIR)/ggml.o
	@rm $(WHISPER_DIR)/libwhisper.a
	@go clean --cache

fetch-whisper:
	@git submodule init
	@git submodule update

fetch-model:
	@${MAKE} -C ./whisper.cpp base.en
	@cp $(WHISPER_DIR)/models/ggml-$(MODEL_NAME).bin $(MODELS_DIR)

build-whisper-lib:
	@${MAKE} -C ./whisper.cpp libwhisper.a

rtc:
	@${MAKE} -C ./rtc run

client:
	@${MAKE} -C ./client run

tts:
	@${MAKE} -C ./tts/servers/coqui-tts run
