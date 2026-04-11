package main

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	rtc "QunDev/GoRemoteDesktop_Server/internal/webrtc"
	socket "QunDev/GoRemoteDesktop_Server/internal/websocket"
	"fmt"
	"log"

	"github.com/pion/webrtc/v4"
)

var (
	pc *webrtc.PeerConnection
	ws *rtc.WsWriter
)

func main() {
	conn, err := socket.SetupHostConnection("localhost:8080")
	if err != nil {
		log.Fatal("Error connect to server: ", err)
	}
	defer conn.Close()

	ws = &rtc.WsWriter{Conn: conn}
	log.Println("Connect successfully")

	for {
		var msg *protocol.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error read msg:", err)
			break
		}

		switch msg.Type {
		case protocol.TypeRegisterHost:
			signalPayload, err := protocol.DecodeSignalPayload(msg.Payload)
			if err != nil {
				log.Println("Error decode TypeRegisterHost payload:", err)
				break
			}
			fmt.Println(signalPayload.ID)
		case protocol.TypeSignal:
			signalPayload, err := protocol.DecodeSignalPayload(msg.Payload)
			if err != nil {
				log.Println("Error decode TypeSignal payload:", err)
			}
			fmt.Println("Signal received:", signalPayload.ID)
		case protocol.TypeOffer:
			pc, err = rtc.HandleCreateOffer(ws)
			if err != nil {
				log.Println("Error handleCreateOffer:", err)
			}
		case protocol.TypeAnswer:
			if err := rtc.HandleAnswer(msg.Payload, pc); err != nil {
				log.Println("Error handleAnswer:", err)
			}
		case protocol.TypeICECandidate:
			if err := rtc.HandleRemoteICECandidate(msg.Payload, pc); err != nil {
				log.Println("Error handleRemoteICECandidate:", err)
			}
		default:
			fmt.Printf("[WS] Unknown msg.Type=%q, payload=%s\n", msg.Type, string(msg.Payload))
		}
	}
}
