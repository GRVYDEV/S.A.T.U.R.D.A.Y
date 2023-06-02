WHISPER_DIR := $(abspath ./whisper.cpp)
MODEL_NAME := base.en
MODEL_FILE_NAME := ggml-$(MODEL_NAME).bin
MODELS_DIR := models

fetch-whisper:
	@git submodule init
	@git submodule update

fetch-model:
	@${MAKE} -C ../whisper.cpp base.en
	@cp $(WHISPER_DIR)/models/ggml-$(MODEL_NAME).bin $(MODELS_DIR)

build-whisper-lib:
	@${MAKE} -C ./whisper.cpp libwhisper.a

run-rtc:
	@${MAKE} -C ./rtc run

run-client:
	@${MAKE} -C ./client run

run-tester:
	@${MAKE} -C ./tester run
