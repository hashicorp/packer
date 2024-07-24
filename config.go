// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
)

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
func (c *config) LoadExternalComponentsFromConfig() error {
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

	if len(pluginPaths) == 0 {
		return nil
	}

	componentList := &strings.Builder{}
	for _, path := range pluginPaths {
		fmt.Fprintf(componentList, "- %s\n", path)
	}

	return fmt.Errorf("Your configuration file describes some legacy components: \n%s"+
		"Packer does not support these mono-component plugins anymore.\n"+
		"Please refer to our Installing Plugins docs for an overview of how to manage installation of local plugins:\n"+
		"https://developer.hashicorp.com/packer/docs/plugins/install-plugins",
		componentList.String())
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
			c.Plugins.Builders.Set(builder, func() (packersdk.Builder, error) {
				args := []string{"execute"}

				if packer.PackerUseProto {
					args = append(args, "--protobuf")
				}

				args = append(args, fmt.Sprintf("packer-builder-%s", builder))

				return c.Plugins.Client(packerPath, args...).Builder()
			})
		}
	}

	for provisioner := range command.Provisioners {
		provisioner := provisioner
		if !c.Plugins.Provisioners.Has(provisioner) {
			c.Plugins.Provisioners.Set(provisioner, func() (packersdk.Provisioner, error) {
				args := []string{"execute"}

				if packer.PackerUseProto {
					args = append(args, "--protobuf")
				}

				args = append(args, fmt.Sprintf("packer-provisioner-%s", provisioner))

				return c.Plugins.Client(packerPath, args...).Provisioner()
			})
		}
	}

	for postProcessor := range command.PostProcessors {
		postProcessor := postProcessor
		if !c.Plugins.PostProcessors.Has(postProcessor) {
			c.Plugins.PostProcessors.Set(postProcessor, func() (packersdk.PostProcessor, error) {
				args := []string{"execute"}

				if packer.PackerUseProto {
					args = append(args, "--protobuf")
				}

				args = append(args, fmt.Sprintf("packer-post-processor-%s", postProcessor))

				return c.Plugins.Client(packerPath, args...).PostProcessor()
			})
		}
	}

	for dataSource := range command.Datasources {
		dataSource := dataSource
		if !c.Plugins.DataSources.Has(dataSource) {
			c.Plugins.DataSources.Set(dataSource, func() (packersdk.Datasource, error) {
				args := []string{"execute"}

				if packer.PackerUseProto {
					args = append(args, "--protobuf")
				}

				args = append(args, fmt.Sprintf("packer-datasource-%s", dataSource))

				return c.Plugins.Client(packerPath, args...).Datasource()
			})
		}
	}

	return nil
}
