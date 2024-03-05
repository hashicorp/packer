// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
)

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

type config struct {
	DisableCheckpoint          bool              `json:"disable_checkpoint"`
	DisableCheckpointSignature bool              `json:"disable_checkpoint_signature"`
	RawBuilders                map[string]string `json:"builders"`
	RawProvisioners            map[string]string `json:"provisioners"`
	RawPostProcessors          map[string]string `json:"post-processors"`

	Plugins *packer.PluginConfig
}

// decodeConfig decodes configuration in JSON format from the given io.Reader into
// the config object pointed to.
func decodeConfig(r io.Reader, c *config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

// LoadExternalComponentsFromConfig loads plugins defined in RawBuilders, RawProvisioners, and RawPostProcessors.
func (c *config) LoadExternalComponentsFromConfig() {
	// helper to build up list of plugin paths
	extractPaths := func(m map[string]string) []string {
		paths := make([]string, 0, len(m))
		for _, v := range m {
			paths = append(paths, v)
		}

		return paths
	}

	var pluginPaths []string
	pluginPaths = append(pluginPaths, extractPaths(c.RawProvisioners)...)
	pluginPaths = append(pluginPaths, extractPaths(c.RawBuilders)...)
	pluginPaths = append(pluginPaths, extractPaths(c.RawPostProcessors)...)

	var externallyUsed = make([]string, 0, len(pluginPaths))
	for _, pluginPath := range pluginPaths {
		name, err := c.loadSingleComponent(pluginPath)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("loaded plugin: %s = %s", name, pluginPath)
		externallyUsed = append(externallyUsed, name)
	}

	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external plugins %v", externallyUsed)
	}
}

func (c *config) loadSingleComponent(path string) (string, error) {
	pluginName := filepath.Base(path)

	// On Windows, ignore any plugins that don't end in .exe.
	// We could do a full PATHEXT parse, but this is probably good enough.
	if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(pluginName)) != ".exe" {
		return "", fmt.Errorf("error loading plugin %q, no exe extension", path)
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("error loading plugin %q: %s", path, err)
	}

	// If the filename has a ".", trim up to there
	if idx := strings.Index(pluginName, "."); idx >= 0 {
		pluginName = pluginName[:idx]
	}

	switch {
	case strings.HasPrefix(pluginName, "packer-builder-"):
		pluginName = pluginName[len("packer-builder-"):]
		c.Plugins.Builders.Set(pluginName, func() (packersdk.Builder, error) {
			return c.Plugins.Client(path).Builder()
		})
	case strings.HasPrefix(pluginName, "packer-post-processor-"):
		pluginName = pluginName[len("packer-post-processor-"):]
		c.Plugins.PostProcessors.Set(pluginName, func() (packersdk.PostProcessor, error) {
			return c.Plugins.Client(path).PostProcessor()
		})
	case strings.HasPrefix(pluginName, "packer-provisioner-"):
		pluginName = pluginName[len("packer-provisioner-"):]
		c.Plugins.Provisioners.Set(pluginName, func() (packersdk.Provisioner, error) {
			return c.Plugins.Client(path).Provisioner()
		})
	}

	return pluginName, nil
}

// This is a proper packer.BuilderFunc that can be used to load packersdk.Builder
// implementations from the defined plugins.
func (c *config) StartBuilder(name string) (packersdk.Builder, error) {
	log.Printf("Loading builder: %s\n", name)
	return c.Plugins.Builders.Start(name)
}

// This is a proper implementation of packer.HookFunc that can be used
// to load packersdk.Hook implementations from the defined plugins.
func (c *config) StarHook(name string) (packersdk.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return c.Plugins.Client(name).Hook()
}

// This is a proper packersdk.PostProcessorFunc that can be used to load
// packersdk.PostProcessor implementations from defined plugins.
func (c *config) StartPostProcessor(name string) (packersdk.PostProcessor, error) {
	log.Printf("Loading post-processor: %s", name)
	return c.Plugins.PostProcessors.Start(name)
}

// This is a proper packer.ProvisionerFunc that can be used to load
// packer.Provisioner implementations from defined plugins.
func (c *config) StartProvisioner(name string) (packersdk.Provisioner, error) {
	log.Printf("Loading provisioner: %s\n", name)
	return c.Plugins.Provisioners.Start(name)
}

func (c *config) discoverInternalComponents() error {
	// Get the packer binary path
	packerPath, err := os.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
		return err
	}

	for builder := range command.Builders {
		builder := builder
		if !c.Plugins.Builders.Has(builder) {
			bin := fmt.Sprintf("%s%sexecute%spacker-builder-%s",
				packerPath, PACKERSPACE, PACKERSPACE, builder)
			c.Plugins.Builders.Set(builder, func() (packersdk.Builder, error) {
				return c.Plugins.Client(bin).Builder()
			})
		}
	}

	for provisioner := range command.Provisioners {
		provisioner := provisioner
		if !c.Plugins.Provisioners.Has(provisioner) {
			bin := fmt.Sprintf("%s%sexecute%spacker-provisioner-%s",
				packerPath, PACKERSPACE, PACKERSPACE, provisioner)
			c.Plugins.Provisioners.Set(provisioner, func() (packersdk.Provisioner, error) {
				return c.Plugins.Client(bin).Provisioner()
			})
		}
	}

	for postProcessor := range command.PostProcessors {
		postProcessor := postProcessor
		if !c.Plugins.PostProcessors.Has(postProcessor) {
			bin := fmt.Sprintf("%s%sexecute%spacker-post-processor-%s",
				packerPath, PACKERSPACE, PACKERSPACE, postProcessor)
			c.Plugins.PostProcessors.Set(postProcessor, func() (packersdk.PostProcessor, error) {
				return c.Plugins.Client(bin).PostProcessor()
			})
		}
	}

	for dataSource := range command.Datasources {
		dataSource := dataSource
		if !c.Plugins.DataSources.Has(dataSource) {
			bin := fmt.Sprintf("%s%sexecute%spacker-datasource-%s",
				packerPath, PACKERSPACE, PACKERSPACE, dataSource)
			c.Plugins.DataSources.Set(dataSource, func() (packersdk.Datasource, error) {
				return c.Plugins.Client(bin).Datasource()
			})
		}
	}

	return nil
}
