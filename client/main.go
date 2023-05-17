package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"sync"

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

	conn := newConn(c)

	go func() {
		for {
			_, message, err := conn.ws.ReadMessage()
			if err != nil {
				log.Printf("err reading message %+v", err)
			}

			log.Printf("recv: %s", message)
		}
	}()

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected

}

// FIXME add to new file

type Conn struct {
	ws                *websocket.Conn
	conn              *webrtc.PeerConnection
	pendingCandidates []*webrtc.ICECandidate
	mu                sync.Mutex
}

func newConn(ws *websocket.Conn) *Conn {
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
		log.Fatal("pc err", err)
	}

	// When an ICE candidate is available send to the other Pion instance
	// the other Pion instance will add this candidate by calling AddICECandidate
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
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
		if err := ws.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("Error sending websocket message %+v err %+v", msg, err)
		}
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			log.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		kind := "unknown kind"
		if t.Kind() == webrtc.RTPCodecTypeVideo {
			kind = "video"
		} else if t.Kind() == webrtc.RTPCodecTypeAudio {
			kind = "audio"
		}
		log.Printf("got track %s", kind)
	})

	return &Conn{
		ws:                ws,
		conn:              peerConnection,
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	// defer func() {
	// 	if err := peerConnection.Close(); err != nil {
	// 		fmt.Printf("cannot close peerConnection: %v\n", err)
	// 	}
	// }()
}

func (c *Conn) Offer(offer webrtc.SessionDescription) error {
	return c.conn.SetRemoteDescription(offer)
}

func (c *Conn) Answer() (*webrtc.SessionDescription, error) {
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
		if err = c.conn.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.String()}); err != nil {
			log.Printf("error adding ice candidate %s %+v", candidate.String(), err)
		}
	}

	return &answer, nil
}

func (c *Conn) AddIceCandidate(candidate *webrtc.ICECandidate) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn.RemoteDescription() == nil {
		c.pendingCandidates = append(c.pendingCandidates, candidate)
		return nil
	} else {
		return c.conn.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.String()})
	}
}
