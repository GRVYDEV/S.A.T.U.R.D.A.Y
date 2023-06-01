WHISPER_DIR := $(abspath ../whisper.cpp)
MODELS_DIR := models
# TODO make this configurable
MODEL_NAME := base.en
MODEL_FILE_NAME := ggml-$(MODEL_NAME).bin
INCLUDE_PATH := $(WHISPER_DIR)
LIBRARY_PATH := $(WHISPER_DIR)

pre-reqs:
ifeq ($(wildcard $(WHISPER_DIR)/*),)
	@echo "fetching whisper repo"
	@${MAKE} -C ../ fetch-whisper
endif

ifeq ($(wildcard $(WHISPER_DIR)/libwhisper.a),)
	@echo "building whisper lib"
	@${MAKE} -C ../ build-whisper-lib
endif

ifeq ($(wildcard $(MODELS_DIR)/$(MODEL_FILE_NAME)),)
	@echo "fetching model"
	@${MAKE} -C ../whisper.cpp base.en
	@cp $(WHISPER_DIR)/models/ggml-$(MODEL_NAME).bin $(MODELS_DIR)
endif

run: pre-reqs
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} go run cmd/whisper.cpp/main.go

debug: pre-reqs
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} PKG_CONFIG_PATH= go run . --debug=true
