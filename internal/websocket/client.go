package websocket

import (
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"bytes"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send   chan []byte
	logger logger.Logger
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
}

func NewClient(hub *Hub, conn *websocket.Conn, logger logger.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		logger: logger,
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request, cfg *config.Config, logger logger.Logger) {
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
	client := NewClient(hub, conn, logger)
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("websocket read error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
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

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.logger.Errorf("websocket write error: %v", err)
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
