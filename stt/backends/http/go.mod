module github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/backends/http

go 1.20

replace (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log => ../../../log
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine => ../../engine
)

require github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine v0.0.0-20230607012320-a2f6bc254fdf

require (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log v0.0.0-20230607014313-65b30ebb4805 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
)
