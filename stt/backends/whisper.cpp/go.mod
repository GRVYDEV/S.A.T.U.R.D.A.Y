module github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/backends/whisper.cpp

go 1.20

replace (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log => ../../../log
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine => ../../engine
)

require (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log v0.0.0-00010101000000-000000000000
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine v0.0.0-00010101000000-000000000000
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-20230524181101-5e2b3407ef46
)

require golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
