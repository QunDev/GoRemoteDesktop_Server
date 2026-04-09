package websocket

import (
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"net/http"
)

func RegisterHandlers(cfg *config.Config, hub *Hub) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r, cfg)
	})
}
