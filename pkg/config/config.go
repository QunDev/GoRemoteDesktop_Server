package config

import "time"

type Config struct {
	Server          ServerConfig    `mapstructure:"server"`
	Logger          LoggerConfig    `mapstructure:"logger"`
	WebSocketConfig WebSocketConfig `mapstructure:"websocket"`
}

type ServerConfig struct {
	Development bool `mapstructure:"development"`
	Port        int  `mapstructure:"port"`
}

type LoggerConfig struct {
	Development      bool     `mapstructure:"development"`
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	OutputPaths      []string `mapstructure:"output_paths"`
	WarnOutputPaths  []string `mapstructure:"warn_output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	DisableCallstack bool     `mapstructure:"disable_callstack"`
	DisableCaller    bool     `mapstructure:"disable_caller"`

	EncoderConfig struct {
		MessageKey      string `mapstructure:"message_key"`
		LevelKey        string `mapstructure:"level_key"`
		TimeKey         string `mapstructure:"time_key"`
		NameKey         string `mapstructure:"name_key"`
		CallerKey       string `mapstructure:"caller_key"`
		StacktraceKey   string `mapstructure:"stacktrace_key"`
		LevelEncoder    string `mapstructure:"level_encoder"`
		TimeEncoder     string `mapstructure:"time_encoder"`
		DurationEncoder string `mapstructure:"duration_encoder"`
		CallerEncoder   string `mapstructure:"caller_encoder"`
	} `mapstructure:"encoder_config"`

	Sampling struct {
		Initial    int `mapstructure:"initial"`
		Thereafter int `mapstructure:"thereafter"`
	} `mapstructure:"sampling"`

	Rotation struct {
		Enabled    bool `mapstructure:"enabled"`
		MaxSizeMB  int  `mapstructure:"max_size_mb"`
		MaxBackups int  `mapstructure:"max_backups"`
		MaxAgeDays int  `mapstructure:"max_age_days"`
		Compress   bool `mapstructure:"compress"`
		LocalTime  bool `mapstructure:"local_time"`
	} `mapstructure:"rotation"`
}

type WebSocketConfig struct {
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	MaxMessageSize  int64         `mapstructure:"max_message_size"`
	PongWait        time.Duration `mapstructure:"pong_wait"`
	WriteWait       time.Duration `mapstructure:"write_wait"`
}
