package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/go-checkpoint"
	"github.com/mitchellh/packer/packer"
)

func runCheckpoint(c *config) {
	configDir, err := ConfigDir()
	if err != nil {
		log.Printf("[ERR] Checkpoint setup error: %s", err)
		return
	}

	version := packer.Version
	if packer.VersionPrerelease != "" {
		version += fmt.Sprintf(".%s", packer.VersionPrerelease)
	}

	_, err = checkpoint.Check(&checkpoint.CheckParams{
		Product:       "packer",
		Version:       version,
		SignatureFile: filepath.Join(configDir, "checkpoint_signature"),
		CacheFile:     filepath.Join(configDir, "checkpoint_cache"),
	})
	if err != nil {
		log.Printf("[ERR] Checkpoint error: %s", err)
	}
}
