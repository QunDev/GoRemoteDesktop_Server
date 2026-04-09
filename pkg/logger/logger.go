package logger

import (
	config2 "QunDev/GoRemoteDesktop_Server/pkg/config"
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	With(fields ...zap.Field) Logger
	Sync() error

	DebugCtx(ctx context.Context, msg string, fields ...zap.Field)
	InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
	WarnCtx(ctx context.Context, msg string, fields ...zap.Field)
	ErrorCtx(ctx context.Context, msg string, fields ...zap.Field)

	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

type loggerImpl struct {
	zap   *zap.Logger
	sugar *zap.SugaredLogger
	cfg   *config2.Config
}

var (
	globalLogger Logger
	mu           sync.RWMutex
)

func NewLogger(cfg *config2.Config) (Logger, error) {
	return newLoggerImpl(cfg)
}

func newLoggerImpl(cfg *config2.Config) (*loggerImpl, error) {
	encoderConfig := cfg.BuildEncoderConfig()

	level, err := zapcore.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	if cfg.Logger.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	outputSyncer, err := BuildWriteSyncers(&cfg.Logger, cfg.Logger.OutputPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create output syncers: %w", err)
	}

	errorSyncer, err := BuildWriteSyncers(&cfg.Logger, cfg.Logger.ErrorOutputPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create error syncers: %w", err)
	}

	warnSyncer, err := BuildWriteSyncers(&cfg.Logger, cfg.Logger.WarnOutputPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create warn syncers: %w", err)
	}

	var primaryCore zapcore.Core
	if cfg.Logger.Development {
		primaryCore = zapcore.NewCore(encoder, outputSyncer, level)
	} else {
		primaryCore = zapcore.NewSamplerWithOptions(
			zapcore.NewCore(encoder, outputSyncer, level),
			time.Second,
			cfg.Logger.Sampling.Initial,
			cfg.Logger.Sampling.Thereafter,
		)
	}

	errorCore := zapcore.NewCore(
		encoder,
		errorSyncer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.ErrorLevel && level <= lvl
		}),
	)

	warnCore := zapcore.NewCore(
		encoder,
		warnSyncer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.WarnLevel && level <= lvl
		}),
	)

	core := zapcore.NewTee(primaryCore, errorCore, warnCore)

	opts := []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(1),
	}

	if !cfg.Logger.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	if cfg.Logger.DisableCallstack {
		opts = append(opts, zap.AddStacktrace(zap.DPanicLevel))
	}

	if cfg.Logger.Development {
		opts = append(opts, zap.Development())
	}

	zapLogger := zap.New(core, opts...)

	return &loggerImpl{
		zap:   zapLogger,
		sugar: zapLogger.Sugar(),
		cfg:   cfg,
	}, nil
}

func Get() Logger {
	mu.RLock()
	defer mu.RUnlock()

	if globalLogger == nil {
		l, _ := newLoggerImpl(config2.Get())
		return l
	}
	return globalLogger
}

func SetGlobalLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = l
}

func Sync() error {
	mu.RLock()
	defer mu.RUnlock()

	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

func (l *loggerImpl) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *loggerImpl) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *loggerImpl) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *loggerImpl) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *loggerImpl) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)
}

func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}

func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}

func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}

func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	l.sugar.Fatalf(format, args...)
}

func (l *loggerImpl) With(fields ...zap.Field) Logger {
	newZap := l.zap.With(fields...)
	return &loggerImpl{
		zap:   newZap,
		sugar: newZap.Sugar(),
		cfg:   l.cfg,
	}
}

func (l *loggerImpl) Sync() error {
	return l.zap.Sync()
}

func (l *loggerImpl) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.Debug(msg, append(fields, extractContextFields(ctx)...)...)
}

func (l *loggerImpl) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.Info(msg, append(fields, extractContextFields(ctx)...)...)
}

func (l *loggerImpl) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.Warn(msg, append(fields, extractContextFields(ctx)...)...)
}

func (l *loggerImpl) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.Error(msg, append(fields, extractContextFields(ctx)...)...)
}

func (l *loggerImpl) WithError(err error) Logger {
	return l.With(zap.Error(err))
}

func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	return l.With(zap.Any(key, value))
}

func (l *loggerImpl) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return l.With(zapFields...)
}

func (l *loggerImpl) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

func (l *loggerImpl) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

func (l *loggerImpl) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

func (l *loggerImpl) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

func (l *loggerImpl) Fatalw(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

func (l *loggerImpl) Zap() *zap.Logger {
	return l.zap
}

func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if ctx == nil {
		return fields
	}

	if reqID := getStringFromContext(ctx, "request_id"); reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}

	if userID := getStringFromContext(ctx, "user_id"); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if traceID := getStringFromContext(ctx, "trace_id"); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if spanID := getStringFromContext(ctx, "span_id"); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	if sessionID := getStringFromContext(ctx, "session_id"); sessionID != "" {
		fields = append(fields, zap.String("session_id", sessionID))
	}

	return fields
}

func getStringFromContext(ctx context.Context, key string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	return ""
}

func GetCaller(skip int) (function, file string, line int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0
	}

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		return fn.Name(), file, line
	}

	return "", file, line
}

func NewContextWithLogFields(ctx context.Context, fields map[string]string) context.Context {
	for k, v := range fields {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}

var (
	String    = zap.String
	Int       = zap.Int
	Int64     = zap.Int64
	Float64   = zap.Float64
	Bool      = zap.Bool
	Any       = zap.Any
	Error     = zap.Error
	Time      = zap.Time
	Duration  = zap.Duration
	Namespace = zap.Namespace
)

type LoggerFields map[string]interface{}

func (lf LoggerFields) ToZapFields() []zap.Field {
	fields := make([]zap.Field, 0, len(lf))
	for k, v := range lf {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
