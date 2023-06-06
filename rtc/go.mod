module github.com/GRVYDEV/S.A.T.U.R.D.A.Y/rtc

go 1.20

require (
	github.com/gorilla/websocket v1.5.0
	github.com/pion/ion-sfu v1.11.0
	github.com/sourcegraph/jsonrpc2 v0.1.0
	github.com/spf13/viper v1.9.0
	internal/ws v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bep/debounce v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/zerologr v1.2.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/lucsky/cuid v1.2.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pion/datachannel v1.5.2 // indirect
	github.com/pion/dtls/v2 v2.1.3 // indirect
	github.com/pion/ice/v2 v2.2.2 // indirect
	github.com/pion/interceptor v0.1.10 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.9 // indirect
	github.com/pion/rtp v1.7.7 // indirect
	github.com/pion/sctp v1.8.2 // indirect
	github.com/pion/sdp/v3 v3.0.4 // indirect
	github.com/pion/srtp/v2 v2.0.5 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.13.0 // indirect
	github.com/pion/turn/v2 v2.0.8 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pion/webrtc/v3 v3.1.25 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rs/zerolog v1.29.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace internal/ws => ./internal/ws

// ion hax
replace github.com/pion/ion-sfu => github.com/GRVYDEV/ion-sfu v0.0.0-20230418020340-dbea2d314f93

//replace github.com/pion/ion-sfu => /Users/garrettgraves/dev/ion-sfu
