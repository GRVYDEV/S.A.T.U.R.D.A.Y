package main

import (
	"net/url"

	"github.com/pion/webrtc/v3"
)

type SaturdayClient struct {
	ws     *SocketConnection
	rtc    *RTCConnection
	config SaturdayConfig
	ae     *AudioEngine
}

type SaturdayConfig struct {
	// ION room name to connect to
	Room string
	// URL for websocket server
	Url url.URL
}

func NewSaturdayClient(config SaturdayConfig) *SaturdayClient {
	transcriptionStream := make(chan TranscriptionSegment, 100)
	ae, err := NewAudioEngine(transcriptionStream)
	if err != nil {
		logger.Fatal(err, "failed to create audio engine")
	}
	ws := NewSocketConnection(config.Url)

	rtc, err := NewRTCConnection(RTCConnectionParams{
		trickleFn: func(candidate *webrtc.ICECandidate, target int) error {
			return ws.SendTrickle(candidate, target)
		},
		rtpChan:             ae.In(),
		transcriptionStream: transcriptionStream,
	})
	if err != nil {
		logger.Fatal(err, "failed to create RTCConnection")
	}

	s := &SaturdayClient{
		ws:     ws,
		rtc:    rtc,
		config: config,
		ae:     ae,
	}

	s.ws.SetOnOffer(s.OnOffer)
	s.ws.SetOnAnswer(s.OnAnswer)
	s.ws.SetOnTrickle(s.rtc.OnTrickle)

	return s
}

func (s *SaturdayClient) OnAnswer(answer webrtc.SessionDescription) error {
	return s.rtc.SetAnswer(answer)
}

func (s *SaturdayClient) OnOffer(offer webrtc.SessionDescription) error {
	ans, err := s.rtc.OnOffer(offer)
	if err != nil {
		logger.Error(err, "error getting answer")
		return err
	}

	return s.ws.SendAnswer(ans)
}

func (s *SaturdayClient) Start() error {
	if err := s.ws.Connect(); err != nil {
		logger.Error(err, "error connecting to websocket")
		return err
	}
	offer, err := s.rtc.GetOffer()
	if err != nil {
		logger.Error(err, "error getting intial offer")
	}
	if err := s.ws.Join(s.config.Room, offer); err != nil {
		logger.Error(err, "error joining room")
		return err
	}

	s.ae.Start()

	s.ws.WaitForDone()
	logger.Info("Socket done goodbye")
	return nil
}
