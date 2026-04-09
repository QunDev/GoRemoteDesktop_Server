package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	configMutex  sync.RWMutex

	onChangeCallbacks []func(*Config)
)

type LoaderOptions struct {
	ConfigPaths     []string
	ConfigName      string
	ConfigType      string
	EnableHotReload bool
}

func DefaultLoaderOptions() *LoaderOptions {
	return &LoaderOptions{
		ConfigPaths: []string{
			"./configs",
		},
		ConfigName:      ConfigFileName,
		ConfigType:      ConfigFileExtension,
		EnableHotReload: true,
	}
}

func NewConfig() (*Config, error) {
	options := DefaultLoaderOptions()

	v := viper.New()

	v.SetConfigName(options.ConfigName)
	v.SetConfigType(options.ConfigType)

	for _, path := range expandPaths(options.ConfigPaths) {
		v.AddConfigPath(path)
	}

	if err := readConfig(v, options); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := parseDurations(v, &cfg); err != nil {
		return nil, err
	}

	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	configMutex.Lock()
	globalConfig = &cfg
	configMutex.Unlock()

	if options.EnableHotReload {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Config file changed: %s", e.Name)
			if err := reloadConfig(v, &cfg); err != nil {
				log.Printf("Failed to reload config: %v", err)
			}
		})
	}

	log.Printf("Config loaded successfully from: %s", v.ConfigFileUsed())
	PrintConfig()
	return globalConfig, nil
}

func readConfig(v *viper.Viper, options *LoaderOptions) error {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found")
		}
		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}

func expandPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if expandedPath := expandPath(path); expandedPath != "" {
			expanded = append(expanded, expandedPath)
		}
	}
	return expanded
}

func expandPath(path string) string {
	if strings.Contains(path, "$") {
		path = os.ExpandEnv(path)
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func reloadConfig(v *viper.Viper, cfg *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	v.ReadInConfig()

	var newConfig Config
	if err := v.Unmarshal(&newConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config on reload: %w", err)
	}

	if err := parseDurations(v, cfg); err != nil {
		return err
	}

	if err := Validate(&newConfig); err != nil {
		return fmt.Errorf("validation failed on reload: %w", err)
	}

	globalConfig = &newConfig

	for _, callback := range onChangeCallbacks {
		callback(&newConfig)
	}

	log.Println("Config reloaded successfully")
	PrintConfig()
	return nil
}

func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

func parseDurations(v *viper.Viper, cfg *Config) error {
	if str := v.GetString("websocket.pong_wait"); str != "" {
		if dur, err := time.ParseDuration(str); err == nil {
			cfg.WebSocketConfig.PongWait = dur
		}
	}
	if str := v.GetString("websocket.write_wait"); str != "" {
		if dur, err := time.ParseDuration(str); err == nil {
			cfg.WebSocketConfig.WriteWait = dur
		}
	}
	return nil
}
