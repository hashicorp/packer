package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/go-checkpoint"
	"github.com/mitchellh/packer/packer"
)

// runCheckpoint runs a HashiCorp Checkpoint request. You can read about
// Checkpoint here: https://github.com/hashicorp/go-checkpoint.
func runCheckpoint(c *config) {
	// If the user doesn't want checkpoint at all, then return.
	if c.DisableCheckpoint {
		log.Printf("[INFO] Checkpoint disabled. Not running.")
		return
	}

	configDir, err := ConfigDir()
	if err != nil {
		log.Printf("[ERR] Checkpoint setup error: %s", err)
		return
	}

	version := packer.Version
	if packer.VersionPrerelease != "" {
		version += fmt.Sprintf(".%s", packer.VersionPrerelease)
	}

	signaturePath := filepath.Join(configDir, "checkpoint_signature")
	if c.DisableCheckpointSignature {
		log.Printf("[INFO] Checkpoint signature disabled")
		signaturePath = ""
	}

	_, err = checkpoint.Check(&checkpoint.CheckParams{
		Product:       "packer",
		Version:       version,
		SignatureFile: signaturePath,
		CacheFile:     filepath.Join(configDir, "checkpoint_cache"),
	})
	if err != nil {
		log.Printf("[ERR] Checkpoint error: %s", err)
	}
}
