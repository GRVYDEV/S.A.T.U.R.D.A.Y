package main

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// JoinConfig allow adding more control to the peers joining a SessionLocal.
type JoinConfig struct {
	// If true the peer will not be allowed to publish tracks to SessionLocal.
	NoPublish bool
	// If true the peer will not be allowed to subscribe to other peers in SessionLocal.
	NoSubscribe bool
	// If true the peer will not automatically subscribe all tracks,
	// and then the peer can use peer.Subscriber().AddDownTrack/RemoveDownTrack
	// to customize the subscrbe stream combination as needed.
	// this parameter depends on NoSubscribe=false.
	NoAutoSubscribe bool
}

// TODO move these to core
// Join message sent when initializing a peer connection
type Join struct {
	SID    string                     `json:"sid"`
	UID    string                     `json:"uid"`
	Offer  *webrtc.SessionDescription `json:"offer,omitempty"`
	Config JoinConfig                 `json:"config,omitempty"`
}

// Negotiation message sent when renegotiating the peer connection
type Negotiation struct {
	Desc webrtc.SessionDescription `json:"desc"`
}

// Trickle message sent when renegotiating the peer connection
type Trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

type Message[T Join | Negotiation | Trickle] struct {
	Method string `json:"method"`
	Params T      `json:"params"`
}

type SocketConnection struct {
	url  url.URL
	ws   *websocket.Conn
	done chan int

	// called when we get a remote offer
	onOffer func(offer webrtc.SessionDescription) error
	// called when we get a remote candidate
	onTrickle func(candidate webrtc.ICECandidateInit, target int) error
}

func NewSocketConnection(url url.URL) *SocketConnection {
	return &SocketConnection{
		url:  url,
		done: make(chan int),
	}
}

func (s *SocketConnection) WaitForDone() {
	<-s.done
}

func (s *SocketConnection) SetOnOffer(onOffer func(offer webrtc.SessionDescription) error) {
	s.onOffer = onOffer
}

func (s *SocketConnection) SetOnTrickle(onTrickle func(candidate webrtc.ICECandidateInit, target int) error) {
	s.onTrickle = onTrickle
}

func (s *SocketConnection) Connect(room string) error {
	c, _, err := websocket.DefaultDialer.Dial(s.url.String(), nil)
	if err != nil {
		logger.Error(err, "dial err")
		return err
	}

	s.ws = c

	msg := Message[Join]{
		Method: "join",
		Params: Join{
			SID: room,
			UID: "SaturdayClient",
			Config: JoinConfig{
				NoPublish: true,
			},
		},
	}

	if err = s.sendMessage(msg); err != nil {
		logger.Errorf(err, "Error sending join message %+v", msg)
		return err
	}

	go s.readMessages()
	return err
}

func (s *SocketConnection) readMessages() error {
	for {
		_, message, err := s.ws.ReadMessage()
		if err != nil {
			logger.Error(err, "err reading message")
			s.ws.Close()
			close(s.done)
			return err
		}

		var msg map[string]interface{}

		json.Unmarshal(message, &msg)

		// FIXME handle errors better
		switch msg["method"] {
		case "offer":
			params, ok := msg["params"].(map[string]interface{})
			if !ok {
				logger.Infof("invalid params for offer %+v", msg["params"])
				continue
			}
			ty, ok := params["type"].(string)
			if !ok {
				logger.Infof("invalid type for offer %+v", params["type"])
				continue
			}
			sdp, ok := params["sdp"].(string)
			if !ok {
				logger.Infof("invalid sdp for offer %+v", params["sdp"])
				continue
			}

			offer := webrtc.SessionDescription{Type: webrtc.NewSDPType(ty), SDP: sdp}

			if s.onOffer != nil {
				if err := s.onOffer(offer); err != nil {
					logger.Errorf(err, "error calling onOffer with offer %+v", offer)
				}
			}
		case "trickle":
			params, ok := msg["params"].(map[string]interface{})
			if !ok {
				logger.Infof("invalid params for trickle %+v", msg["params"])
				continue
			}

			paramsJson, err := json.Marshal(params)
			if err != nil {
				logger.Error(err, "error marshalling trickle params")
				continue
			}

			var trickle Trickle

			if err = json.Unmarshal(paramsJson, &trickle); err != nil {
				logger.Error(err, "error unmarshalling trickle params")
				continue
			}

			if s.onTrickle != nil {
				if err := s.onTrickle(trickle.Candidate, 1); err != nil {
					logger.Errorf(err, "error calling onTrickle with candidate %+v", trickle)
				}
			}

		default:
			logger.Infof("got unhandled message: %+v", msg)
		}

	}
}

func (s *SocketConnection) SendTrickle(candidate *webrtc.ICECandidate, target int) error {
	if candidate == nil {
		return nil
	}

	msg := Message[Trickle]{
		Method: "trickle",
		Params: Trickle{
			Target:    target,
			Candidate: candidate.ToJSON(),
		},
	}

	log.Print("Sending trickle")

	return s.sendMessage(msg)
}

func (s *SocketConnection) SendAnswer(answer webrtc.SessionDescription) error {
	msg := Message[Negotiation]{
		Method: "answer",
		Params: Negotiation{
			Desc: answer,
		},
	}

	log.Print("Sending answer")

	return s.sendMessage(msg)
}

func (s *SocketConnection) sendMessage(msg any) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message to json %+v, %+v", msg, err)
		return err
	}
	log.Printf("Sending message %s", payload)
	if err := s.ws.WriteMessage(websocket.TextMessage, payload); err != nil {
		log.Printf("Error sending websocket message %+v err %+v", msg, err)
		return err
	}
	return nil
}
