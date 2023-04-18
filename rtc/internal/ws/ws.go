package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
)

var defaultConfig = sfu.JoinConfig{
	NoPublish:       false,
	NoSubscribe:     false,
	NoAutoSubscribe: false,
}

// Join message sent when initializing a peer connection
type Join struct {
	SID    string                    `json:"sid"`
	UID    string                    `json:"uid"`
	Offer  webrtc.SessionDescription `json:"offer"`
	Config sfu.JoinConfig            `json:"config"`
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

type Connection struct {
	*sfu.PeerLocal
	logr.Logger
}

func NewConnection(p *sfu.PeerLocal, l logr.Logger) *Connection {
	return &Connection{p, l}
}

func (c *Connection) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	replyError := func(err error) {
		_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    500,
			Message: fmt.Sprintf("%s", err),
		})
	}

	switch req.Method {
	case "join":
		var join Join
		err := json.Unmarshal(*req.Params, &join)
		if err != nil {
			c.Logger.Error(err, "connect: error parsing offer")
			replyError(err)
			break
		}

		c.Logger.Info(fmt.Sprintf("Join msg %+v", join))

		c.OnOffer = func(offer *webrtc.SessionDescription) {
			if err := conn.Notify(ctx, "offer", offer); err != nil {
				c.Logger.Error(err, "error sending offer")
			}

		}
		c.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
			if err := conn.Notify(ctx, "trickle", Trickle{
				Candidate: *candidate,
				Target:    target,
			}); err != nil {
				c.Logger.Error(err, "error sending ice candidate")
			}
		}

		err = c.Join(join.SID, join.UID, join.Config)
		if err != nil {
			replyError(err)
			break
		}

		if !join.Config.NoPublish {
			answer, err := c.Answer(join.Offer)
			if err != nil {
				replyError(err)
				break
			}
			_ = conn.Reply(ctx, req.ID, answer)
		}

	case "offer":
		var negotiation Negotiation
		err := json.Unmarshal(*req.Params, &negotiation)
		if err != nil {
			c.Logger.Error(err, "connect: error parsing offer")
			replyError(err)
			break
		}

		answer, err := c.Answer(negotiation.Desc)
		if err != nil {
			replyError(err)
			break
		}
		_ = conn.Reply(ctx, req.ID, answer)

	case "answer":
		var negotiation Negotiation
		err := json.Unmarshal(*req.Params, &negotiation)
		if err != nil {
			c.Logger.Error(err, "connect: error parsing answer")
			replyError(err)
			break
		}

		err = c.SetRemoteDescription(negotiation.Desc)
		if err != nil {
			replyError(err)
		}

	case "trickle":
		var trickle Trickle
		err := json.Unmarshal(*req.Params, &trickle)
		if err != nil {
			c.Logger.Error(err, "connect: error parsing candidate")
			replyError(err)
			break
		}

		err = c.Trickle(trickle.Candidate, trickle.Target)
		if err != nil {
			c.Logger.Error(err, "error setting trickle")
			replyError(err)
		}
	}
}
