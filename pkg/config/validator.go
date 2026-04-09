package config

import (
	"fmt"
)

func Validate(cfg *Config) error {
	if err := validateServer(cfg.Server); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := ValidateWebSocketConfig(cfg.WebSocketConfig); err != nil {
		return fmt.Errorf("web socket config: %w", err)
	}
	return nil
}

func validateServer(cfg ServerConfig) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("server config: port must be between 1 and 65535")
	}
	return nil
}

func ValidateWebSocketConfig(cfg WebSocketConfig) error {
	if cfg.ReadBufferSize < 1 || cfg.WriteBufferSize < 1 {
		return fmt.Errorf("web socket config: read buffer size must be between 1")
	}
	if cfg.MaxMessageSize < 1 {
		return fmt.Errorf("web socket config: max message size must be greater or equal to 1")
	}
	return nil
}
