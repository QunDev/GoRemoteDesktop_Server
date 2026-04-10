package main

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	"QunDev/GoRemoteDesktop_Server/internal/websocket"
	"fmt"
	"log"
)

func main() {
	conn, err := websocket.SetupHostConnection("localhost:8080")
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
		}
	}
}
