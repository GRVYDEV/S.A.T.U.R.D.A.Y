WHISPER_DIR := $(abspath ../whisper.cpp)
MODELS_DIR := models
INCLUDE_PATH := $(abspath ./lib)
LIBRARY_PATH := $(abspath ./lib)
HAS_WHISPER := "$(shell find $(WHISPER_DIR) -maxdepth 0 -empty)"
HAS_WHISPER := "foo"
HAS_MODEL := "foo"



test: pre-reqs

pre-reqs:
ifneq ($(HAS_WHISPER), "")
	@echo "fetching whisper repo"
	#@${MAKE} -C ../ fetch-whisper
endif

ifneq ($(HAS_MODEL), "")
	@echo "fetching model"
endif

run:
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} go run .
