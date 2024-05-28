// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeConfig(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true,
		"provisioners": {
		    "super-shell": "packer-provisioner-super-shell"
		}
	}`

	var cfg config
	err := decodeConfig(strings.NewReader(packerConfig), &cfg)
	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	var expectedCfg config
	json.NewDecoder(strings.NewReader(packerConfig)).Decode(&expectedCfg)
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("failed to load custom configuration data; expected %v got %v", expectedCfg, cfg)
	}

}
