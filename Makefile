fetch-whisper:
	@git submodule init
	@git submodule update

build-whisper-lib:
	@${MAKE} -C ./whisper.cpp libwhisper.a

run-rtc:
	@${MAKE} -C ./rtc run

run-client:
	@${MAKE} -C ./client run

run-tester:
	@${MAKE} -C ./tester run
