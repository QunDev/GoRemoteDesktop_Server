package webrtc

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type WsWriter struct {
	mu   sync.Mutex
	Conn *websocket.Conn
}

func (w *WsWriter) send(msgType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := protocol.Message{Type: msgType, Payload: data}
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Conn.WriteMessage(websocket.TextMessage, raw)
}

func HandleCreateOffer(ws *WsWriter) (*webrtc.PeerConnection, error) {
	onCandidate := func(c *webrtc.ICECandidate) {
		sdpMid := ""
		if c.ToJSON().SDPMid != nil {
			sdpMid = *c.ToJSON().SDPMid
		}
		sdpMLineIndex := uint16(0)
		if c.ToJSON().SDPMLineIndex != nil {
			sdpMLineIndex = *c.ToJSON().SDPMLineIndex
		}

		if err := ws.send(protocol.TypeICECandidate, protocol.ICECandidatePayload{
			Candidate:     c.ToJSON().Candidate,
			SDPMid:        sdpMid,
			SDPMLineIndex: sdpMLineIndex,
		}); err != nil {
			log.Println("[ICE] Send candidate error:", err)
		}
	}
	newPC, desc, err := CreateHostOffer(onCandidate)
	if err != nil {
		return nil, fmt.Errorf("CreateHostOffer: %w", err)
	}

	if err := ws.send(protocol.TypeOffer, protocol.OfferPayload{
		SDP:  desc.SDP,
		Type: desc.Type.String(),
	}); err != nil {
		return nil, fmt.Errorf("send offer: %w", err)
	}

	watchConnectionState(newPC)
	log.Println("[Host] Offer sent ✓")
	return newPC, nil
}

func HandleAnswer(raw json.RawMessage, pc *webrtc.PeerConnection) error {
	if pc == nil {
		return fmt.Errorf("PeerConnection chưa được khởi tạo")
	}
	payload, err := protocol.DecodeOfferPayload(raw)
	if err != nil {
		return fmt.Errorf("decode answer: %w", err)
	}

	answer := webrtc.SessionDescription{
		Type: webrtc.NewSDPType(payload.Type),
		SDP:  payload.SDP,
	}
	if err := pc.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("SetRemoteDescription: %w", err)
	}

	log.Println("[Host] Remote description set ✓")
	return nil
}

func HandleRemoteICECandidate(raw json.RawMessage, pc *webrtc.PeerConnection) error {
	if pc == nil {
		return fmt.Errorf("PeerConnection chưa được khởi tạo")
	}
	payload, err := protocol.DecodeICECandidatePayload(raw)
	if err != nil {
		return fmt.Errorf("decode ICE candidate: %w", err)
	}

	sdpMid := payload.SDPMid
	sdpMLineIndex := payload.SDPMLineIndex

	if err := pc.AddICECandidate(webrtc.ICECandidateInit{
		Candidate:     payload.Candidate,
		SDPMid:        &sdpMid,
		SDPMLineIndex: &sdpMLineIndex,
	}); err != nil {
		return fmt.Errorf("AddICECandidate: %w", err)
	}

	log.Println("[ICE] Remote candidate added ✓")
	return nil
}
