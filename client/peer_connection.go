package main

import (
	"os"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type PeerConn struct {
	conn              *webrtc.PeerConnection
	pendingCandidates []webrtc.ICECandidateInit
	mu                sync.Mutex
	rtpIn             chan<- *rtp.Packet
}

func NewPeerConn(onICECandidate func(candidate *webrtc.ICECandidate), rtpIn chan<- *rtp.Packet) *PeerConn {
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		logger.Fatal(err, "pc err")
	}

	pc := &PeerConn{
		conn:              peerConnection,
		pendingCandidates: make([]webrtc.ICECandidateInit, 0),
		rtpIn:             rtpIn,
	}

	// When an ICE candidate is available send to the other Pion instance
	// the other Pion instance will add this candidate by calling AddICECandidate
	peerConnection.OnICECandidate(onICECandidate)

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		logger.Infof("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			logger.Info("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		kind := "unknown kind"
		if t.Kind() == webrtc.RTPCodecTypeVideo {
			kind = "video"
		} else if t.Kind() == webrtc.RTPCodecTypeAudio {
			kind = "audio"
			go func() {
				for {
					pkt, _, err := t.ReadRTP()
					if err != nil {
						logger.Error(err, "err reading rtp")
						return
					}
					pc.rtpIn <- pkt
				}
			}()
		}
		logger.Debugf("got track %s", kind)
	})

	return pc
	// defer func() {
	// 	if err := peerConnection.Close(); err != nil {
	// 		fmt.Printf("cannot close peerConnection: %v\n", err)
	// 	}
	// }()
}

func (c *PeerConn) Offer(offer webrtc.SessionDescription) error {
	return c.conn.SetRemoteDescription(offer)
}

func (c *PeerConn) Answer() (*webrtc.SessionDescription, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	answer, err := c.conn.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	if err = c.conn.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	for _, candidate := range c.pendingCandidates {
		if err = c.conn.AddICECandidate(candidate); err != nil {
			logger.Errorf(err, "error adding ice candidate %+v", candidate)
		}
	}

	c.pendingCandidates = make([]webrtc.ICECandidateInit, 0)

	return &answer, nil
}

func (c *PeerConn) AddIceCandidate(candidate webrtc.ICECandidateInit) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// we got a candiate before the offer here so buffer
	if c.conn.RemoteDescription() == nil {
		c.pendingCandidates = append(c.pendingCandidates, candidate)
		return nil
	} else {
		return c.conn.AddICECandidate(candidate)
	}
}
