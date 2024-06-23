package core

import (
	"wb-tech-backend/internal/pkg/web"

	"github.com/spf13/viper"
)

type NatsConfig struct {
	SubUrl    string `yaml:"suburl"`
	PubUrl    string `yaml:"puburl"`
	ClusterId string `yaml:"cluster"`
	Sub       string `yaml:"sub"`
	Prod      string `yaml:"prod"`
	Subject   string `yaml:"subject"`
}

type StorageConfig struct {
	URL string `yaml:"url" env-required:"true"`
}

type Config struct {
	Storage StorageConfig    `yaml:"storage"`
	Server  web.ServerConfig `yaml:"server"`
	Nats    NatsConfig       `yaml:"nats"`
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
