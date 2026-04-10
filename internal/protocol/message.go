package protocol

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SignalPayload struct {
	ID string `json:"id"`
}

func DecodeSignalPayload(d json.RawMessage) (*SignalPayload, error) {
	var p SignalPayload
	if err := json.Unmarshal(d, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
