package hcl2template

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

// PluginRequirements returns a sorted list of plugin requirements.
func (cfg *PackerConfig) PluginRequirements() (plugingetter.List, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var reqs plugingetter.List
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

			reqs = append(reqs, &plugingetter.Plugin{
				Identifier:         block.Type,
				VersionConstraints: block.Requirement.Required,
			})
			uniq[name] = block
		}

	}

	sort.Sort(reqs)

	return reqs, diags
}
