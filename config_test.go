package main

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestLoadConfig(t *testing.T) {
	configData := `
        [server]
        hostname = "example.com"
        public_key = "example.key"
        private_key = "example.pem"
        `

	var config Config
	r := strings.NewReader(configData)
	_, err := toml.DecodeReader(r, &config)
	if err != nil {
		t.Errorf("could not parse example config properly")
	}

	err = ValidateConfig(config)

	if err != nil {
		t.Errorf("could not validate config: %v", err)
	}

	if config.Server.Hostname != "example.com" {
		t.Errorf(
			"config hostname expected example.com got: %s", config.Server.Hostname,
		)
	}

	if config.Server.PublicKey != "example.key" {
		t.Errorf(
			"config public_key expected example.key got: %s", config.Server.PublicKey,
		)
	}

	if config.Server.PrivateKey != "example.pem" {
		t.Errorf(
			"config private_key expected example.pem got: %s", config.Server.PrivateKey,
		)
	}
}
