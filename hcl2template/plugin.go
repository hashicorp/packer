package hcl2template

import (
	"crypto/sha256"
	"fmt"
	"log"
	"runtime"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
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

func (cfg *PackerConfig) detectPluginBinaries() hcl.Diagnostics {
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
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("no plugin installed for %s %v", pluginRequirement.Identifier, pluginRequirement.VersionConstraints.String()),
				Detail:   "Did you run packer init for this project ?",
			})
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

	return diags
}

func (cfg *PackerConfig) initializeBlocks() hcl.Diagnostics {
	// verify that all used plugins do exist and expand dynamic bodies
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for i := range build.Sources {
			// here we grab a pointer to the source usage because we will set
			// its body.
			srcUsage := &(build.Sources[i])
			if !cfg.parser.PluginConfig.Builders.Has(srcUsage.Type) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + buildSourceLabel + " type " + srcUsage.Type,
					Subject:  &build.HCL2Ref.DefRange,
					Detail:   fmt.Sprintf("known builders: %v", cfg.parser.PluginConfig.Builders.List()),
					Severity: hcl.DiagError,
				})
				continue
			}

			sourceDefinition, found := cfg.Sources[srcUsage.SourceRef]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + sourceLabel + " " + srcUsage.String(),
					Subject:  build.HCL2Ref.DefRange.Ptr(),
					Severity: hcl.DiagError,
					Detail:   fmt.Sprintf("Known: %v", cfg.Sources),
					// TODO: show known sources as a string slice here ^.
				})
				continue
			}

			body := sourceDefinition.block.Body
			if srcUsage.Body != nil {
				// merge additions into source definition to get a new body.
				body = hcl.MergeBodies([]hcl.Body{body, srcUsage.Body})
			}
			// expand any dynamic block.
			body = dynblock.Expand(body, cfg.EvalContext(BuildContext, nil))

			srcUsage.Body = body
		}

		for _, provBlock := range build.ProvisionerBlocks {
			if !cfg.parser.PluginConfig.Provisioners.Has(provBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+buildProvisionerLabel+" type %q", provBlock.PType),
					Subject:  provBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+buildProvisionerLabel+"s: %v", cfg.parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
			// Allow rest of the body to have dynamic blocks
			provBlock.HCL2Ref.Rest = dynblock.Expand(provBlock.HCL2Ref.Rest, cfg.EvalContext(BuildContext, nil))
		}

		if build.ErrorCleanupProvisionerBlock != nil {
			if !cfg.parser.PluginConfig.Provisioners.Has(build.ErrorCleanupProvisionerBlock.PType) {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  fmt.Sprintf("Unknown "+buildErrorCleanupProvisionerLabel+" type %q", build.ErrorCleanupProvisionerBlock.PType),
					Subject:  build.ErrorCleanupProvisionerBlock.HCL2Ref.TypeRange.Ptr(),
					Detail:   fmt.Sprintf("known "+buildErrorCleanupProvisionerLabel+"s: %v", cfg.parser.PluginConfig.Provisioners.List()),
					Severity: hcl.DiagError,
				})
			}
			// Allow rest of the body to have dynamic blocks
			build.ErrorCleanupProvisionerBlock.HCL2Ref.Rest = dynblock.Expand(build.ErrorCleanupProvisionerBlock.HCL2Ref.Rest, cfg.EvalContext(BuildContext, nil))
		}

		for _, ppList := range build.PostProcessorsLists {
			for _, ppBlock := range ppList {
				if !cfg.parser.PluginConfig.PostProcessors.Has(ppBlock.PType) {
					diags = append(diags, &hcl.Diagnostic{
						Summary:  fmt.Sprintf("Unknown "+buildPostProcessorLabel+" type %q", ppBlock.PType),
						Subject:  ppBlock.HCL2Ref.TypeRange.Ptr(),
						Detail:   fmt.Sprintf("known "+buildPostProcessorLabel+"s: %v", cfg.parser.PluginConfig.PostProcessors.List()),
						Severity: hcl.DiagError,
					})
				}
				// Allow the rest of the body to have dynamic blocks
				ppBlock.HCL2Ref.Rest = dynblock.Expand(ppBlock.HCL2Ref.Rest, cfg.EvalContext(BuildContext, nil))
			}
		}

	}

	return diags
}
