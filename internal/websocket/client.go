package websocket

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	HostID   string
	ClientID string
	hub      *Hub

	conn *websocket.Conn

	send   chan *protocol.Message
	logger logger.Logger
	role   string
}

var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
}

func NewClient(hub *Hub, conn *websocket.Conn, logger logger.Logger, role string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan *protocol.Message, 256),
		logger: logger,
		role:   role,
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request, cfg *config.Config, logger logger.Logger, role string) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  cfg.WebSocketConfig.ReadBufferSize,
		WriteBufferSize: cfg.WebSocketConfig.ReadBufferSize, CheckOrigin: func(r *http.Request) bool {
			if cfg.Server.Development {
				return true
			}
			return allowedOrigins[r.Header.Get("Origin")]
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("websocket upgrade error: %v", err)
		return
	}
	client := NewClient(hub, conn, logger, role)
	client.HostID = r.URL.Query().Get("hostid")
	client.hub.register <- client

	go client.writePump(cfg.WebSocketConfig.PongWait, cfg.WebSocketConfig.WriteWait)
	go client.readPump(cfg.WebSocketConfig.MaxMessageSize, cfg.WebSocketConfig.PongWait)
}

func (c *Client) readPump(maxMessageSize int64, pongWait time.Duration) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg *protocol.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error read msg:", err)
			break
		}
		c.hub.broadcast <- map[*Client]*protocol.Message{
			c: msg,
		}
	}
}

func (c *Client) writePump(pongWait, writeWait time.Duration) {
	ticker := time.NewTicker((pongWait * 9) / 10)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				c.logger.Errorf("write json error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func SetupHostConnection(serverAddr string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws", RawQuery: "role=host"}

	log.Println("Connecting to WebSocket Server...")

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
