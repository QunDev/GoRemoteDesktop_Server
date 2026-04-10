package websocket

import (
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"net/http"
)

func RegisterHandlers(cfg *config.Config, hub *Hub, logger logger.Logger) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r, cfg, logger)
	})
}
