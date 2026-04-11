package protocol

import (
	"encoding/json"

	"github.com/pion/webrtc/v4"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SignalPayload struct {
	ID string `json:"id"`
}

type OfferPayload struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ICECandidatePayload struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdp_mid"`
	SDPMLineIndex uint16 `json:"sdp_mline_index"`
}

type OnICECandidateFunc func(candidate *webrtc.ICECandidate)

func DecodeSignalPayload(d json.RawMessage) (*SignalPayload, error) {
	var p SignalPayload
	if err := json.Unmarshal(d, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func DecodeOfferPayload(d json.RawMessage) (*OfferPayload, error) {
	var p OfferPayload
	if err := json.Unmarshal(d, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func DecodeICECandidatePayload(raw json.RawMessage) (*ICECandidatePayload, error) {
	var p ICECandidatePayload
	return &p, json.Unmarshal(raw, &p)
}
