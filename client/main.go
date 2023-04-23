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
	SID    string                    `json:"sid"`
	UID    string                    `json:"uid"`
	Offer  webrtc.SessionDescription `json:"offer"`
	Config JoinConfig                `json:"config"`
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
	method string
	params T
}

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	onIce := func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		msg := Message[Trickle]{
			method: "trickle",
			params: Trickle{
				Target:    1,
				Candidate: webrtc.ICECandidateInit{Candidate: candidate.String()},
			},
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message to json %+v", msg, err)
			return
		}
		if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("Error sending websocket message %+v err %+v", msg, err)
		}
	}

	_ = NewConnection(onIce)

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected

}
