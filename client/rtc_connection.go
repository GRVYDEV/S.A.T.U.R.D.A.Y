package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"S.A.T.U.R.D.A.Y/stt/engine"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type RTCConnection struct {
	sub PeerConn
	pub PeerConn
	// channel that incoming audio rtp packets will be relayed on
	rtpIn chan<- *rtp.Packet
	// channel to send outgoing audio samples to
	mediaIn    <-chan media.Sample
	audioTrack *webrtc.TrackLocalStaticSample
}

type RTCConnectionParams struct {
	trickleFn           func(*webrtc.ICECandidate, int) error
	rtpChan             chan<- *rtp.Packet
	transcriptionStream <-chan engine.TranscriptionSegment
	mediaIn             <-chan media.Sample
}

// FIXME if transcriptionStream AND mediaIn are not provided this will blow up
func NewRTCConnection(params RTCConnectionParams) (*RTCConnection, error) {
	rtc := &RTCConnection{
		rtpIn:   params.rtpChan,
		mediaIn: params.mediaIn,
	}

	sub := NewPeerConn(func(candidate *webrtc.ICECandidate) {
		params.trickleFn(candidate, 1)
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
						Logger.Error(err, "err reading rtp")
						return
					}
					rtc.rtpIn <- pkt
				}
			}()
		}
		Logger.Debugf("got track %s", kind)
	})

	rtc.sub = sub

	pub := NewPeerConn(func(candidate *webrtc.ICECandidate) {
		params.trickleFn(candidate, 0)
	})

	if params.mediaIn != nil {
		audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "saturday_audio")
		if err != nil {
			Logger.Error(err, "error creating local audio track")
			return nil, err
		}

		_, err = pub.conn.AddTransceiverFromTrack(audioTrack, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly})
		if err != nil {
			Logger.Error(err, "error adding local audio transceiver")
			return nil, err
		}

		rtc.audioTrack = audioTrack

		go rtc.processOutgoingMedia()
	} else {
		Logger.Info("mediaIn not provided... audio relay is disabled")
	}

	if params.transcriptionStream != nil {
		ordered := true
		maxRetransmits := uint16(0)

		dc, err := pub.conn.CreateDataChannel(
			"transcriptions",
			&webrtc.DataChannelInit{
				Ordered:        &ordered,
				MaxRetransmits: &maxRetransmits,
			})
		if err != nil {
			return nil, err
		}

		dc.OnOpen(func() {
			Logger.Info("data channel opened...")

			for transcription := range params.transcriptionStream {
				Logger.Infof("got transcript %s", transcription.Text)
				data, err := json.Marshal(transcription)
				if err != nil {
					Logger.Error(err, "error marshalling transcript")
					continue
				}
				Logger.Debugf("sending transcript %+v on data channel", transcription)
				dc.Send(data)
			}
		})

	} else {
		Logger.Info("transcriptionStream not provided... transcription relay is disabled")
	}

	rtc.pub = pub

	return rtc, nil
}

// processIncomingMedia sends the provided samples on the audioTrack
func (r *RTCConnection) processOutgoingMedia() {
	if r.mediaIn == nil {
		Logger.Info("MediaIn not provided... skipping relay")
		return
	}
	for sample := range r.mediaIn {
		if err := r.audioTrack.WriteSample(sample); err != nil {
			Logger.Error(err, "error writing sample")
		}
	}
}

func (r *RTCConnection) OnTrickle(candidate webrtc.ICECandidateInit, target int) error {
	switch target {
	case 0:
		return r.pub.AddIceCandidate(candidate)
	case 1:
		return r.sub.AddIceCandidate(candidate)
	default:
		err := errors.New(fmt.Sprintf("unknown target %d for candidate", target))
		Logger.Error(err, "error OnTrickle")
		return err
	}
}

func (r *RTCConnection) GetOffer() (webrtc.SessionDescription, error) {
	return r.pub.GetOffer()
}

func (r *RTCConnection) SetAnswer(answer webrtc.SessionDescription) error {
	return r.pub.SetAnswer(answer)
}

func (r *RTCConnection) OnOffer(offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	var answer = webrtc.SessionDescription{}
	if err := r.sub.Offer(offer); err != nil {
		Logger.Error(err, "error setting offer")
		return answer, err
	}

	answer, err := r.sub.Answer()
	if err != nil {
		Logger.Error(err, "error getting answer")
		return answer, err
	}
	return answer, nil
}
