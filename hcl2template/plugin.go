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
	"github.com/hashicorp/packer-plugin-sdk/didyoumean"
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
	// Then we can apply any constraint from the template, if any
	opts := plugingetter.ListInstallationsOptions{
		PluginDirectory: cfg.parser.PluginConfig.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
			ReleasesOnly: cfg.parser.PluginConfig.ReleasesOnly,
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

	// Do a second pass to discover the remaining installed plugins
	err := cfg.parser.PluginConfig.Discover()
	if err != nil {
		return (hcl.Diagnostics{}).Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to discover installed plugins",
			Detail:   err.Error(),
		})
	}

	return diags
}

func (cfg *PackerConfig) initializeBlocks() hcl.Diagnostics {
	// verify that all used plugins do exist
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for i := range build.Sources {
			// here we grab a pointer to the source usage because we will set
			// its body.
			srcUsage := &(build.Sources[i])
			if !cfg.parser.PluginConfig.Builders.Has(srcUsage.Type) {
				detail := fmt.Sprintf(
					"The %s %s is unknown by Packer, and is likely part of a plugin that is not installed.\n"+
						"You may find the needed plugin along with installation instructions documented on the Packer integrations page.\n\n"+
						"https://developer.hashicorp.com/packer/integrations?filter=%s",
					buildSourceLabel,
					srcUsage.Type,
					strings.Split(srcUsage.Type, "-")[0],
				)

				if sugg := didyoumean.NameSuggestion(srcUsage.Type, cfg.parser.PluginConfig.Builders.List()); sugg != "" {
					detail = fmt.Sprintf("Did you mean to use %q?", sugg)
				}
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + buildSourceLabel + " type " + srcUsage.Type,
					Subject:  &build.HCL2Ref.DefRange,
					Detail:   detail,
					Severity: hcl.DiagError,
				})
				continue
			}

			sourceDefinition, found := cfg.Sources[srcUsage.SourceRef]
			if !found {
				availableSrcs := listAvailableSourceNames(cfg.Sources)
				detail := fmt.Sprintf("Known: %v", availableSrcs)
				if sugg := didyoumean.NameSuggestion(srcUsage.SourceRef.String(), availableSrcs); sugg != "" {
					detail = fmt.Sprintf("Did you mean to use %q?", sugg)
				}
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + sourceLabel + " " + srcUsage.SourceRef.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   detail,
				})
				continue
			}

			body := sourceDefinition.block.Body
			if srcUsage.Body != nil {
				// merge additions into source definition to get a new body.
				body = hcl.MergeBodies([]hcl.Body{body, srcUsage.Body})
			}

			srcUsage.Body = body
		}

		for _, provBlock := range build.ProvisionerBlocks {
			if !cfg.parser.PluginConfig.Provisioners.Has(provBlock.PType) {
				detail := fmt.Sprintf(
					"The %s %s is unknown by Packer, and is likely part of a plugin that is not installed.\n"+
						"You may find the needed plugin along with installation instructions documented on the Packer integrations page.\n\n"+
						"https://developer.hashicorp.com/packer/integrations?filter=%s",
					buildProvisionerLabel,
					provBlock.PType,
					strings.Split(provBlock.PType, "-")[0],
				)

				if sugg := didyoumean.NameSuggestion(provBlock.PType, cfg.parser.PluginConfig.Provisioners.List()); sugg != "" {
					detail = fmt.Sprintf("Did you mean to use %q?", sugg)
				}

				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+buildProvisionerLabel+" type %q", provBlock.PType),
					Subject:  provBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   detail,
					Severity: hcl.DiagError,
				})
			}
		}

		if build.ErrorCleanupProvisionerBlock != nil {
			if !cfg.parser.PluginConfig.Provisioners.Has(build.ErrorCleanupProvisionerBlock.PType) {
				detail := fmt.Sprintf(
					"The %s %s is unknown by Packer, and is likely part of a plugin that is not installed.\n"+
						"You may find the needed plugin along with installation instructions documented on the Packer integrations page.\n\n"+
						"https://developer.hashicorp.com/packer/integrations?filter=%s",
					buildErrorCleanupProvisionerLabel,
					build.ErrorCleanupProvisionerBlock.PType,
					strings.Split(build.ErrorCleanupProvisionerBlock.PType, "-")[0],
				)

				if sugg := didyoumean.NameSuggestion(build.ErrorCleanupProvisionerBlock.PType, cfg.parser.PluginConfig.Provisioners.List()); sugg != "" {
					detail = fmt.Sprintf("Did you mean to use %q?", sugg)
				}

				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+buildErrorCleanupProvisionerLabel+" type %q", build.ErrorCleanupProvisionerBlock.PType),
					Subject:  build.ErrorCleanupProvisionerBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   detail,
					Severity: hcl.DiagError,
				})
			}
		}

		for _, ppList := range build.PostProcessorsLists {
			for _, ppBlock := range ppList {
				if !cfg.parser.PluginConfig.PostProcessors.Has(ppBlock.PType) {
					detail := fmt.Sprintf(
						"The %s %s is unknown by Packer, and is likely part of a plugin that is not installed.\n"+
							"You may find the needed plugin along with installation instructions documented on the Packer integrations page.\n\n"+
							"https://developer.hashicorp.com/packer/integrations?filter=%s",
						buildPostProcessorLabel,
						ppBlock.PType,
						strings.Split(ppBlock.PType, "-")[0],
					)

					if sugg := didyoumean.NameSuggestion(ppBlock.PType, cfg.parser.PluginConfig.PostProcessors.List()); sugg != "" {
						detail = fmt.Sprintf("Did you mean to use %q?", sugg)
					}

					diags = append(diags, &hcl.Diagnostic{
						Summary:  fmt.Sprintf("Unknown "+buildPostProcessorLabel+" type %q", ppBlock.PType),
						Subject:  ppBlock.HCL2Ref.TypeRange.Ptr(),
						Detail:   detail,
						Severity: hcl.DiagError,
					})
				}
			}
		}

	}

	return diags
}
