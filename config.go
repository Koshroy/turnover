package main

import (
	"github.com/BurntSushi/toml"
)

//ServerConfig defines config options for running the server
type ServerConfig struct {
	Hostname string
}

//Config is the config object
type Config struct {
	Server ServerConfig
}

// LoadConfig loads a config at configPath
func LoadConfig(configPath string) (*Config, error) {
	var conf Config
	_, err := toml.DecodeFile(configPath, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
