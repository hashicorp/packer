// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/pathing"
)

// PluginFolder returns the known plugin folder based on system.
func PluginFolder() (string, error) {
	var res string

	if packerPluginPath := os.Getenv("PACKER_PLUGIN_PATH"); packerPluginPath != "" {
		return packerPluginPath, nil
	}

	cd, err := pathing.ConfigDir()
	if err != nil {
		log.Printf("[ERR] Error loading config directory: %v", err)
		return res, err
	}

	return filepath.Join(cd, "plugins"), nil
}
