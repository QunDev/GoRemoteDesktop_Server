package config

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

func PrintConfig() {
	cfg := Get()
	if cfg == nil {
		fmt.Println("Config not loaded")
		return
	}

	fmt.Println("=== Current Configuration ===")
	fmt.Printf("Server: %+v\n", cfg.Server)
	fmt.Printf("Logger: %+v\n", cfg.Logger)
	fmt.Printf("WebSocket: %+v\n", cfg.WebSocketConfig)
	fmt.Println("=============================")
}

func getLevelEncoder(encoder string) zapcore.LevelEncoder {
	switch encoder {
	case "lowercaseColor":
		return zapcore.LowercaseColorLevelEncoder
	case "capital":
		return zapcore.CapitalLevelEncoder
	case "capitalColor":
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

func getTimeEncoder(encoder string) zapcore.TimeEncoder {
	switch encoder {
	case "epoch":
		return zapcore.EpochTimeEncoder
	case "millis":
		return zapcore.EpochMillisTimeEncoder
	case "nanos":
		return zapcore.EpochNanosTimeEncoder
	case "iso8601":
		return zapcore.ISO8601TimeEncoder
	default:
		return zapcore.ISO8601TimeEncoder
	}
}

func getDurationEncoder(encoder string) zapcore.DurationEncoder {
	switch encoder {
	case "nanos":
		return zapcore.NanosDurationEncoder
	case "ms":
		return zapcore.MillisDurationEncoder
	default:
		return zapcore.StringDurationEncoder
	}
}

func getCallerEncoder(encoder string) zapcore.CallerEncoder {
	switch encoder {
	case "full":
		return zapcore.FullCallerEncoder
	default:
		return zapcore.ShortCallerEncoder
	}
}

func (c *Config) BuildEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     c.Logger.EncoderConfig.MessageKey,
		LevelKey:       c.Logger.EncoderConfig.LevelKey,
		TimeKey:        c.Logger.EncoderConfig.TimeKey,
		NameKey:        c.Logger.EncoderConfig.NameKey,
		CallerKey:      c.Logger.EncoderConfig.CallerKey,
		StacktraceKey:  c.Logger.EncoderConfig.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    getLevelEncoder(c.Logger.EncoderConfig.LevelEncoder),
		EncodeTime:     getTimeEncoder(c.Logger.EncoderConfig.TimeEncoder),
		EncodeDuration: getDurationEncoder(c.Logger.EncoderConfig.DurationEncoder),
		EncodeCaller:   getCallerEncoder(c.Logger.EncoderConfig.CallerEncoder),
	}

	if c.Logger.Development {
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}
