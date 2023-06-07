module github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/backends/http

go 1.20

replace (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log => ../../../log
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine => ../../engine
)

require github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine v0.0.0-00010101000000-000000000000

require (
	github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log v0.0.0-20230607012320-a2f6bc254fdf // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
)
