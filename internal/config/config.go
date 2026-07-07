package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server Server `yaml:"server" mapstructure:"server"`
	Logger Logger `yaml:"logger" mapstructure:"logger"`
}

type Server struct {
	Port       uint16 `yaml:"port" mapstructure:"port"`
	OnlineMode bool   `yaml:"online_mode" mapstructure:"online_mode"`
	MaxPlayers uint64 `yaml:"max_players" mapstructure:"max_players"`
	MOTD       string `yaml:"motd" mapstructure:"motd"`
}

type Logger struct {
	Type  string `yaml:"type" mapstructure:"type"`
	Level string `yaml:"level" mapstructure:"level"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Server: Server{
			Port:       25565,
			OnlineMode: true,
			MaxPlayers: 10,
			MOTD:       "A Minecraft Server",
		},
		Logger: Logger{
			Type:  "text",
			Level: "info",
		},
	}
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			data, err := yaml.Marshal(NewDefaultConfig())

			if err != nil {
				return nil, fmt.Errorf("default configuration marshalling error: %w", err)
			}

			if err := os.WriteFile(path, data, 0644); err != nil {
				return nil, fmt.Errorf("error writing the configuration file: %w", err)
			}

			return NewDefaultConfig(), nil
		}
		return nil, fmt.Errorf("configuration read error: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &cfg, nil
}
