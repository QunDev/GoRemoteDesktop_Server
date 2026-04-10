package websocket

import (
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"net/http"
)

func RegisterHandlers(cfg *config.Config, hub *Hub, logger logger.Logger) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		if role == "" || (role != "client" && role != "host") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		serveWs(hub, w, r, cfg, logger, role)
	})
}
