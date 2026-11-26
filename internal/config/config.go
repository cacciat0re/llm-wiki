package config

import (
	"fmt"

	"github.com/jinzhu/configor"
)

type Configs struct {
	Server ServerConfig   `yaml:"server"`
	DB     DatabaseConfig `yaml:"db"`
	CORS   CORSConfig     `yaml:"cors"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"database_name"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func Load(path string) (Configs, error) {
	var cfg Configs
	if err := configor.Load(&cfg, path); err != nil {
		return Configs{}, fmt.Errorf("load config %s: %w", path, err)
	}
	return cfg, nil
}
