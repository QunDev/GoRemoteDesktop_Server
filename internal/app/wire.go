//go:build wireinject
// +build wireinject

package app

import (
	"QunDev/GoRemoteDesktop_Server/internal/websocket"
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"QunDev/GoRemoteDesktop_Server/pkg/logger"
	"context"

	"github.com/google/wire"
)

var ContextSet = wire.NewSet(
	context.Background,
)

var ConfigSet = wire.NewSet(
	config.NewConfig,
)

var LoggerSet = wire.NewSet(
	logger.NewLogger,
)

var SocketServerSet = wire.NewSet(
	websocket.NewHub)

var AppSet = wire.NewSet(
	NewApp,
)

func InitializeApp() (*App, error) {
	wire.Build(
		ContextSet,
		ConfigSet,
		LoggerSet,
		SocketServerSet,
		AppSet,
	)

	return &App{}, nil
}
