package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

//ServerConfig defines config options for running the server
type ServerConfig struct {
	Scheme     string
	Hostname   string
	PublicKey  string `toml:"public_key"`
	PrivateKey string `toml:"private_key"`
}

//Config is the config object
type Config struct {
	Server ServerConfig
}

// LoadConfig loads a config at configPath
func LoadConfig(configPath string) (*Config, error) {
	var conf Config
	md, err := toml.DecodeFile(configPath, &conf)
	if err != nil {
		return nil, err
	}

	undecoded := md.Undecoded()
	if len(undecoded) != 0 {
		return nil, fmt.Errorf("these config fields are unused: %q", undecoded)
	}

	err = ValidateConfig(conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// ValidateConfig validates a Config
func ValidateConfig(conf Config) error {
	if conf.Server.Hostname == "" {
		return fmt.Errorf("no hostname given")
	}

	if conf.Server.PublicKey == "" {
		return fmt.Errorf("no public key path given")
	}

	if conf.Server.PrivateKey == "" {
		return fmt.Errorf("no private key path given")
	}

	if conf.Server.Scheme == "" {
		return fmt.Errorf("no scheme given")
	}

	return nil
}
