// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	"github.com/hashicorp/packer/command"
	packerVersion "github.com/hashicorp/packer/version"
)

func init() {
	checkpointResult = make(chan *checkpoint.CheckResponse, 1)
}

var checkpointResult chan *checkpoint.CheckResponse

// runCheckpoint runs a HashiCorp Checkpoint request. You can read about
// Checkpoint here: https://github.com/hashicorp/go-checkpoint.
func runCheckpoint(c *config) {
	// If the user doesn't want checkpoint at all, then return.
	if c.DisableCheckpoint {
		log.Printf("[INFO] Checkpoint disabled. Not running.")
		checkpointResult <- nil
		return
	}

	configDir, err := pathing.ConfigDir()
	if err != nil {
		log.Printf("[ERR] Checkpoint setup error: %s", err)
		checkpointResult <- nil
		return
	}

	version := packerVersion.Version
	if packerVersion.VersionPrerelease != "" {
		version += fmt.Sprintf("-%s", packerVersion.VersionPrerelease)
	}

	signaturePath := filepath.Join(configDir, "checkpoint_signature")
	if c.DisableCheckpointSignature {
		log.Printf("[INFO] Checkpoint signature disabled")
		signaturePath = ""
	}

	resp, err := checkpoint.Check(&checkpoint.CheckParams{
		Product:       "packer",
		Version:       version,
		SignatureFile: signaturePath,
		CacheFile:     filepath.Join(configDir, "checkpoint_cache"),
	})
	if err != nil {
		log.Printf("[ERR] Checkpoint error: %s", err)
		resp = nil
	}

	checkpointResult <- resp
}

// commandVersionCheck implements command.VersionCheckFunc and is used
// as the version checker.
func commandVersionCheck() (command.VersionCheckInfo, error) {
	// Wait for the result to come through
	info := <-checkpointResult
	if info == nil {
		var zero command.VersionCheckInfo
		return zero, nil
	}

	// Build the alerts that we may have received about our version
	alerts := make([]string, len(info.Alerts))
	for i, a := range info.Alerts {
		alerts[i] = a.Message
	}

	return command.VersionCheckInfo{
		Outdated: info.Outdated,
		Latest:   info.CurrentVersion,
		Alerts:   alerts,
	}, nil
}
