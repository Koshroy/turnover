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
        `

	var config Config
	r := strings.NewReader(configData)
	_, err := toml.DecodeReader(r, &config)
	if err != nil {
		t.Errorf("could not parse example config properly")
	}

	if config.Server.Hostname != "example.com" {
		t.Errorf(
			"config hostname expected example.com got: %s", config.Server.Hostname,
		)
	}
}
