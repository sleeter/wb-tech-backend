package core

import (
	"github.com/spf13/viper"
	"wb-tech-backend/internal/pkg/web"
)

type StorageConfig struct {
	URL string `yaml:"url" env-required:"true"`
}

type Config struct {
	Storage StorageConfig    `yaml:"storage"`
	Server  web.ServerConfig `yaml:"server"`
}

func ParseConfig(loader *viper.Viper) (*Config, error) {
	cfg := &Config{}
	if err := loader.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := loader.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
