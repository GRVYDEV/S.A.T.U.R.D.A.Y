fetch-whisper:
	@git submodule init
	@git submodule update

build-whisper-lib:
	@${MAKE} -C ./whisper.cpp libwhisper.a
