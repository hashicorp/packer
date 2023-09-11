// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"crypto/sha256"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/hashicorp/hcl/v2"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

// PluginRequirements returns a sorted list of plugin requirements.
func (cfg *PackerConfig) PluginRequirements() (plugingetter.Requirements, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var reqs plugingetter.Requirements
	reqPluginsBlocks := cfg.Packer.RequiredPlugins

	// Take all required plugins, make sure there are no conflicting blocks
	// and append them to the list.
	uniq := map[string]*RequiredPlugin{}
	for _, requiredPluginsBlock := range reqPluginsBlocks {
		for name, block := range requiredPluginsBlock.RequiredPlugins {

			if previouslySeenBlock, found := uniq[name]; found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Duplicate required_plugin.%q block", name),
					Detail: fmt.Sprintf("Block previously seen at %s is already named %q.\n", previouslySeenBlock.DeclRange, name) +
						"Names at the left hand side of required_plugins are made available to use in your HCL2 configurations.\n" +
						"To allow to calling to their features correctly two plugins have to have different accessors.",
					Context: &block.DeclRange,
				})
				continue
			}

			reqs = append(reqs, &plugingetter.Requirement{
				Accessor:           name,
				Identifier:         block.Type,
				VersionConstraints: block.Requirement.Required,
			})
			uniq[name] = block
		}

	}

	return reqs, diags
}

func (cfg *PackerConfig) DetectPluginBinaries() hcl.Diagnostics {
	opts := plugingetter.ListInstallationsOptions{
		FromFolders: cfg.parser.PluginConfig.KnownPluginFolders,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	pluginReqs, diags := cfg.PluginRequirements()
	if diags.HasErrors() {
		return diags
	}

	uninstalledPlugins := map[string]string{}

	for _, pluginRequirement := range pluginReqs {
		sortedInstalls, err := pluginRequirement.ListInstallations(opts)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to list installation for %s", pluginRequirement.Identifier),
				Detail:   err.Error(),
			})
			continue
		}
		if len(sortedInstalls) == 0 {
			uninstalledPlugins[pluginRequirement.Identifier.String()] = pluginRequirement.VersionConstraints.String()
			continue
		}
		log.Printf("[TRACE] Found the following %q installations: %v", pluginRequirement.Identifier, sortedInstalls)
		install := sortedInstalls[len(sortedInstalls)-1]
		err = cfg.parser.PluginConfig.DiscoverMultiPlugin(pluginRequirement.Accessor, install.BinaryPath)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Error discovering plugin %s", pluginRequirement.Identifier),
				Detail:   err.Error(),
			})
			continue
		}
	}

	if len(uninstalledPlugins) > 0 {
		detailMessage := &strings.Builder{}
		detailMessage.WriteString("The following plugins are required, but not installed:\n\n")
		for pluginName, pluginVersion := range uninstalledPlugins {
			fmt.Fprintf(detailMessage, "* %s %s\n", pluginName, pluginVersion)
		}
		detailMessage.WriteString("\nDid you run packer init for this project ?")
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing plugins",
			Detail:   detailMessage.String(),
		})
	}

	return diags
}

func (cfg *PackerConfig) initializeBlocks() hcl.Diagnostics {
	// verify that all used plugins do exist
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		diags = diags.Extend(build.Initialize(cfg))
	}

	return diags
}
