package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeConfig_basic(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10001,
		"PluginMaxPort": 26000,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true
	}`

	var cfg config
	err := DecodeConfig(strings.NewReader(packerConfig), &cfg)
	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	var expectedCfg config
	json.NewDecoder(strings.NewReader(packerConfig)).Decode(&expectedCfg)
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("failed to load custom configuration data; expected %v got %v", expectedCfg, cfg)
	}

}

func TestDecodeConfig_provisioners(t *testing.T) {

	packerConfig := `
	{
			"provisioners": {
					"comment": "/tmp/packer-provisioner-comment"
			}
	}`

	var cfg config
	err := DecodeConfig(strings.NewReader(packerConfig), &cfg)

	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	if _, ok := cfg.Provisioners["comment"]; !ok {
		t.Errorf("provisioner by the name of comment was not found")
	}

}
