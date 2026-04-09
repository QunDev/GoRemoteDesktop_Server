package logger

import (
	"QunDev/GoRemoteDesktop_Server/pkg/config"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type lockedWriter struct {
	w  *lumberjack.Logger
	mu sync.Mutex
}

func (lw *lockedWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

func (lw *lockedWriter) Sync() error {
	return nil
}

func BuildWriteSyncer(cfg *config.LoggerConfig, path string) (zapcore.WriteSyncer, error) {
	if path == "stdout" {
		return zapcore.Lock(os.Stdout), nil
	}

	if path == "stderr" {
		return zapcore.Lock(os.Stderr), nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if !cfg.Rotation.Enabled {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		return zapcore.Lock(file), nil
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.Rotation.MaxSizeMB,
		MaxBackups: cfg.Rotation.MaxBackups,
		MaxAge:     cfg.Rotation.MaxAgeDays,
		Compress:   cfg.Rotation.Compress,
		LocalTime:  cfg.Rotation.LocalTime,
	}

	return zapcore.AddSync(&lockedWriter{w: lumberjackLogger}), nil
}

func BuildWriteSyncers(cfg *config.LoggerConfig, paths []string) (zapcore.WriteSyncer, error) {
	var syncers []zapcore.WriteSyncer

	for _, path := range paths {
		syncer, err := BuildWriteSyncer(cfg, path)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, syncer)
	}

	return zapcore.NewMultiWriteSyncer(syncers...), nil
}
