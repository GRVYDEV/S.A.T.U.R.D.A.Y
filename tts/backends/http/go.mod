module github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/backends/http

go 1.20

replace (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log => ../../../log
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/engine => ../../engine
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/util => ../../../util

)

require (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log v0.0.0-00010101000000-000000000000
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/tts/engine v0.0.0-00010101000000-000000000000
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/util v0.0.0-00010101000000-000000000000
)

require golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
