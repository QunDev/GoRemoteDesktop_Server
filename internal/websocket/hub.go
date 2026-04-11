package websocket

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	"QunDev/GoRemoteDesktop_Server/internal/utils"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Hub struct {
	clients       map[string]*Client
	hosts         map[string]*Client
	IDs           map[string]struct{}
	signalSuccess chan map[*Client]*Client // host -> client

	broadcast chan map[*Client]*protocol.Message

	register chan *Client

	unregister chan *Client
	logger     logger.Logger
}

func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		broadcast:     make(chan map[*Client]*protocol.Message),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[string]*Client),
		hosts:         make(map[string]*Client),
		signalSuccess: make(chan map[*Client]*Client, 256),
		logger:        logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			for {
				id := uuid.New().String()
				if _, ok := h.IDs[id]; !ok {
					client.ID = id
					if client.role == "host" {
						h.hosts[id] = client
						p, err := json.Marshal(&protocol.SignalPayload{ID: id})
						if err != nil {
							h.logger.Error("json marshal err: ", zap.Error(err))
							continue
						}
						client.send <- &protocol.Message{
							Type:    protocol.TypeRegisterHost,
							Payload: p,
						}
					} else if client.role == "client" {
						h.clients[id] = client
						if client.HostID != "" && utils.IsValidUUID(client.HostID) {
							for {
								p, err := json.Marshal(&protocol.SignalPayload{ID: id})
								if err != nil {
									h.logger.Error("json marshal err: ", zap.Error(err))
									continue
								}
								if host, ok := h.hosts[client.HostID]; ok {
									host.ClientID = id
									host.send <- &protocol.Message{
										Type:    protocol.TypeSignal,
										Payload: p,
									}
									h.signalSuccess <- map[*Client]*Client{
										host: client,
									}
									break
								}
							}
						}
					}
					break
				}
			}
		case client := <-h.unregister:
			if client.role == "host" {
				if _, ok := h.hosts[client.ID]; ok {
					delete(h.hosts, client.ID)
					close(client.send)
				}
			} else if client.role == "client" {
				if _, ok := h.clients[client.ID]; ok {
					delete(h.clients, client.ID)
					close(client.send)
				}
			}
		case data := <-h.broadcast:
			for client, message := range data {
				switch message.Type {
				case protocol.TypeSignal:
					if payload, err := protocol.DecodeSignalPayload(message.Payload); err == nil {
						for {
							p, err := json.Marshal(&protocol.SignalPayload{ID: client.ID})
							if err != nil {
								h.logger.Error("json marshal err: ", zap.Error(err))
								continue
							}
							if host, ok := h.hosts[payload.ID]; ok {
								host.ClientID = client.ID
								host.send <- &protocol.Message{
									Type:    protocol.TypeSignal,
									Payload: p,
								}
								h.signalSuccess <- map[*Client]*Client{
									host: client,
								}
								break
							}
						}
					}
				case protocol.TypeOffer:
					if _, err := protocol.DecodeOfferPayload(message.Payload); err == nil {
						if c, ok := h.clients[client.ClientID]; ok {
							c.send <- &protocol.Message{
								Type:    protocol.TypeOffer,
								Payload: message.Payload,
							}
						}
					}
				case protocol.TypeAnswer:
					if _, err := protocol.DecodeOfferPayload(message.Payload); err == nil {
						if h, ok := h.hosts[client.ClientID]; ok {
							h.send <- &protocol.Message{
								Type:    protocol.TypeAnswer,
								Payload: message.Payload,
							}
						}
					}
				}
			}
		case signal := <-h.signalSuccess:
			for server := range signal {
				server.send <- &protocol.Message{
					Type: protocol.TypeOffer,
				}
			}
		}
	}
}
