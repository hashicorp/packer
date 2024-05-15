// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/pathing"
)

var pathSep = fmt.Sprintf("%c", os.PathListSeparator)

// PluginFolder returns the known plugin folder based on system.
func PluginFolder() (string, error) {
	if packerPluginPath := os.Getenv("PACKER_PLUGIN_PATH"); packerPluginPath != "" {
		if strings.Contains(packerPluginPath, pathSep) {
			return "", fmt.Errorf("Multiple paths are no longer supported for PACKER_PLUGIN_PATH.\n"+
				"This should be defined as one of the following options for your environment:"+
				"\n* PACKER_PLUGIN_PATH=%v", strings.Join(strings.Split(packerPluginPath, pathSep), "\n* PACKER_PLUGIN_PATH="))
		}

		return packerPluginPath, nil
	}

	cd, err := pathing.ConfigDir()
	if err != nil {
		log.Printf("[ERR] Error loading config directory: %v", err)
		return "", err
	}

	return filepath.Join(cd, "plugins"), nil
}
