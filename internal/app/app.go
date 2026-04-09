package app

import (
	socket "QunDev/GoRemoteDesktop_Server/internal/websocket"
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"context"
)

type App struct {
	Logger logger.Logger
	Config *config.Config
}

func NewApp(cfg *config.Config, logger logger.Logger, ctx context.Context, hub *socket.Hub) *App {
	go hub.Run()
	socket.RegisterHandlers(cfg, hub)
	return &App{
		Logger: logger,
		Config: cfg,
	}
}
