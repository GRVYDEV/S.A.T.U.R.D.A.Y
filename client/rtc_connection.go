package main

import (
	"errors"
	"fmt"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type RTCConnection struct {
	sub   PeerConn
	pub   PeerConn
	rtpIn chan<- *rtp.Packet
}

func NewRTCConnection(trickleFn func(*webrtc.ICECandidate, int) error, rtpIn chan<- *rtp.Packet) *RTCConnection {
	rtc := &RTCConnection{
		rtpIn: rtpIn,
	}

	sub := NewPeerConn(func(candidate *webrtc.ICECandidate) {
		trickleFn(candidate, 1)
	})
	sub.conn.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
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
					rtc.rtpIn <- pkt
				}
			}()
		}
		logger.Debugf("got track %s", kind)
	})

	rtc.sub = sub

	pub := NewPeerConn(func(candidate *webrtc.ICECandidate) {
		trickleFn(candidate, 0)
	})
	rtc.pub = pub

	return rtc
}

func (r *RTCConnection) OnTrickle(candidate webrtc.ICECandidateInit, target int) error {
	switch target {
	case 0:
		return r.pub.AddIceCandidate(candidate)
	case 1:
		return r.sub.AddIceCandidate(candidate)
	default:
		err := errors.New(fmt.Sprintf("unknown target %d for candidate", target))
		logger.Error(err, "error OnTrickle")
		return err
	}
}

func (r *RTCConnection) OnOffer(offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	var answer = webrtc.SessionDescription{}
	if err := r.sub.Offer(offer); err != nil {
		logger.Error(err, "error setting offer")
		return answer, err
	}

	answer, err := r.sub.Answer()
	if err != nil {
		logger.Error(err, "error getting answer")
		return answer, err
	}
	return answer, nil
}
