WHISPER_DIR := $(abspath ../whisper.cpp)
MODELS_DIR := $(abspath ../models)
MODEL_NAME := base.en
MODEL_FILE_NAME := ggml-$(MODEL_NAME).bin
# TODO make this configurable
INCLUDE_PATH := $(WHISPER_DIR)
LIBRARY_PATH := $(WHISPER_DIR)

backend := whisper.cpp

whisper-cpp-pre-reqs:
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
	@${MAKE} -C ../ fetch-model
endif

run-whisper-cpp: whisper-cpp-pre-reqs
	@echo "running client with whisper.cpp backend..."
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} go run cmd/whisper.cpp/main.go

run-faster-whisper:
	@echo "running client with faster-whisper backend..."
	@go run cmd/faster-whisper/main.go

ifeq ($(backend), wcpp)
run: run-whisper-cpp
else ifeq ($(backend), fwhisper)
run: run-faster-whisper
else
run: run-whisper-cpp
endif

debug: pre-reqs
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} PKG_CONFIG_PATH= go run . --debug=true
