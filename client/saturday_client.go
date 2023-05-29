package main

import (
	"net/url"

	"github.com/pion/webrtc/v3"
)

type SaturdayClient struct {
	ws     *SocketConnection
	pc     *PeerConn
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
		logger.Fatalf(err, "failed to create audio engine")
	}
	ws := NewSocketConnection(config.Url)
	pc := NewPeerConn(func(candidate *webrtc.ICECandidate) {
		// TODO make this support both sub and pub
		ws.SendTrickle(candidate, 1)
	}, ae.In())

	s := &SaturdayClient{
		ws:     ws,
		pc:     pc,
		config: config,
		ae:     ae,
	}

	s.ws.SetOnOffer(s.OnOffer)
	s.ws.SetOnTrickle(func(candidate webrtc.ICECandidateInit, target int) error {
		return s.pc.AddIceCandidate(candidate)
	})

	// Starting a new goroutine to read from the channel
	go func() {
		for transcription := range transcriptionStream {
			// Process the received transcription here
			// For now, we will just log it
			logger.Infof("Received transcription: %s", transcription.text)
		}
	}()
	return s
}

func (s *SaturdayClient) OnOffer(offer webrtc.SessionDescription) error {
	if err := s.pc.Offer(offer); err != nil {
		logger.Error(err, "error setting offer")
		return err
	}

	ans, err := s.pc.Answer()
	if err != nil {
		logger.Error(err, "error getting answer")
		return err
	}

	return s.ws.SendAnswer(*ans)
}

func (s *SaturdayClient) Start() error {
	if err := s.ws.Connect(s.config.Room); err != nil {
		logger.Error(err, "error connecting to websocket")
		return err
	}

	s.ae.Start()

	s.ws.WaitForDone()
	logger.Info("Socket done goodbye")
	return nil
}
