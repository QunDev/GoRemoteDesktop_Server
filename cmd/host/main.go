package main

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	rtc "QunDev/GoRemoteDesktop_Server/internal/webrtc"
	socket "QunDev/GoRemoteDesktop_Server/internal/websocket"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	conn, err := socket.SetupHostConnection("localhost:8080")
	if err != nil {
		log.Fatal("Error connect to server: ", err)
	}
	defer conn.Close()

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
			_, desc, err := rtc.CreateHostOffer()
			if err != nil {
				log.Println("Error create host offer:", err)
				break
			}
			data, err := json.Marshal(protocol.OfferPayload{
				SDP:  desc.SDP,
				Type: desc.Type.String(),
			})
			if err != nil {
				fmt.Println("Error encode TypeOffer payload:", err)
				break
			}
			msg := protocol.Message{
				Type:    protocol.TypeOffer,
				Payload: data,
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Error encode TypeOffer payload:", err)
			}
			conn.WriteMessage(websocket.TextMessage, msgBytes)
		case protocol.TypeAnswer:
			fmt.Println("Answer received")
		}
	}
}
