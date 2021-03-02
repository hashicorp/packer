package hcl2template

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/zclconf/go-cty/cty"
)

func (cfg *PackerConfig) decodeRequiredPluginsBlock(f *hcl.File) hcl.Diagnostics {
	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case packerLabel:
			content, contentDiags := block.Body.Content(packerBlockSchema)
			diags = append(diags, contentDiags...)

			// We ignore "packer_version"" here because
			// sniffCoreVersionRequirements already dealt with that

			for _, innerBlock := range content.Blocks {
				switch innerBlock.Type {
				case "required_plugins":
					reqs, reqsDiags := decodeRequiredPluginsBlock(innerBlock)
					diags = append(diags, reqsDiags...)
					cfg.Packer.RequiredPlugins = append(cfg.Packer.RequiredPlugins, reqs)
				default:
					continue
				}

			}
		}
	}
	return diags
}

func (cfg *PackerConfig) decodeImplicitRequiredPluginsBlocks(f *hcl.File) hcl.Diagnostics {
	// when a plugin is used but not defined in the required plugin blocks, it
	// is 'implicitly used'. Here we read common configuration blocks to try to
	// guess plugins.

	// decodeRequiredPluginsBlock needs to be called first; otherwise all
	// required plugins will be implicitly required too.

	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case sourceLabel, dataSourceLabel:
			moreDiags := cfg.inferImplicitRequiredPluginFromBlocks(block)
			diags = append(diags, moreDiags...)
		case buildLabel:
			content, moreDiags := block.Body.Content(buildSchema)
			diags = append(diags, moreDiags...)
			for _, block := range content.Blocks {
				switch block.Type {
				case buildProvisionerLabel, buildPostProcessorLabel:
					moreDiags := cfg.inferImplicitRequiredPluginFromBlocks(block)
					diags = append(diags, moreDiags...)
				case buildPostProcessorsLabel:
					content, moreDiags := block.Body.Content(postProcessorsSchema)
					diags = append(diags, moreDiags...)
					for _, block := range content.Blocks {
						moreDiags := cfg.inferImplicitRequiredPluginFromBlocks(block)
						diags = append(diags, moreDiags...)
					}
				}
			}

		}
	}
	return diags
}

// This function will infer an implicitly required plugin block. For plugins
// that are not present. Usually the plugin name is the first part of a plugin
// call, before the first dash.
//
// Exampes:
//  * data           "amazon-ami"      "..." adds implictly requried "github.com/hashicorp/amazon"
//  * source         "amazon-ebs"      "..." adds implictly requried "github.com/hashicorp/amazon"
//  * source         "google"          "..." adds implictly requried "github.com/hashicorp/google"
//  * provisioner    "windos-restart"  "..." adds implictly requried "github.com/hashicorp/windows"
//  * post-processor "exoscale-import" "..." adds implictly requried "github.com/hashicorp/exoscale"
//  * source         "amazon-v2-ebs"   "..." adds implictly requried "github.com/hashicorp/exoscale"
//
// For now this function will only work with hashicorp plugins, and therefore
// pretend the user meant using an official hashicorp plugin if the plugin could
// not be found.
//
// Plugin name will stop at first dash found, so that means that if a users uses
// an "amazon-v2" plugin, this won't work.
func (cfg *PackerConfig) inferImplicitRequiredPluginFromBlocks(block *hcl.Block) hcl.Diagnostics {
	labels := block.Labels
	if len(labels) == 0 {
		return nil
	}
	var diags hcl.Diagnostics
	probablePluginNames := strings.Split(labels[0], "-")
	if len(probablePluginNames) == 0 || probablePluginNames[0] == "" {
		// probably a WIP config, return now to avoid panics.
		return nil
	}
	probablePluginName := probablePluginNames[0]

	for _, requiredPluginBlock := range cfg.Packer.RequiredPlugins {
		for _, requiredPlugin := range requiredPluginBlock.RequiredPlugins {
			if requiredPlugin.Name == probablePluginName {
				// Found a plugin that matches this name. No implicitly required
				// plugin needed.
				return nil
			}
		}
	}
	fullImportPath := "github.com/hashicorp/" + probablePluginName
	pType, diags := addrs.ParsePluginSourceString(fullImportPath)
	if diags.HasErrors() {
		return diags
	}

	cfg.Packer.RequiredPlugins = append(cfg.Packer.RequiredPlugins, &RequiredPlugins{
		RequiredPlugins: map[string]*RequiredPlugin{
			probablePluginName: {
				Name:   probablePluginName,
				Source: fullImportPath,
				Type:   pType,
				Requirement: VersionConstraint{
					Required: nil, // means latest
				},
				PluginDependencyReason: PluginDependencyImplicit,
			},
		},
		DeclRange: block.DefRange,
	})

	return diags
}

// RequiredPlugin represents a declaration of a dependency on a particular
// Plugin version or source.
type RequiredPlugin struct {
	Name string
	// Source used to be able to tell how the template referenced this source,
	// for example, "awesomecloud" instead of github.com/awesome/awesomecloud.
	// This one is left here in case we want to go back to allowing inexplicit
	// source url definitions.
	Source      string
	Type        *addrs.Plugin
	Requirement VersionConstraint
	DeclRange   hcl.Range
	PluginDependencyReason
}

// PluginDependencyReason is an enumeration of reasons why a dependency might be
// present.
type PluginDependencyReason int

const (
	// PluginDependencyExplicit means that there is an explicit
	// "required_plugin" block in the configuration.
	PluginDependencyExplicit PluginDependencyReason = iota

	// PluginDependencyImplicit means that there is no explicit
	// "required_plugin" block but there is at least one resource that uses this
	// plugin.
	PluginDependencyImplicit
)

type RequiredPlugins struct {
	RequiredPlugins map[string]*RequiredPlugin
	DeclRange       hcl.Range
}

func decodeRequiredPluginsBlock(block *hcl.Block) (*RequiredPlugins, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	ret := &RequiredPlugins{
		RequiredPlugins: nil,
		DeclRange:       block.DefRange,
	}
	for name, attr := range attrs {
		expr, err := attr.Expr.Value(nil)
		if err != nil {
			diags = append(diags, err...)
		}

		nameDiags := checkPluginNameNormalized(name, attr.Expr.Range())
		diags = append(diags, nameDiags...)

		rp := &RequiredPlugin{
			Name:      name,
			DeclRange: attr.Expr.Range(),
		}

		switch {
		case expr.Type().IsPrimitiveType():
			c := "version"
			if cs, _ := decodeVersionConstraint(attr); len(cs.Required) > 0 {
				c = cs.Required.String()
			}

			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid plugin requirement",
				Detail: fmt.Sprintf(`'%s = "%s"' plugin requirement calls are not possible.`+
					` You must define a whole block. For example:`+"\n"+
					`%[1]s = {`+"\n"+
					`  source  = "github.com/hashicorp/%[1]s"`+"\n"+
					`  version = "%[2]s"`+"\n"+`}`,
					name, c),
				Subject: attr.Range.Ptr(),
			})
			continue

		case expr.Type().IsObjectType():
			if !expr.Type().HasAttribute("version") {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No version constraint was set",
					Detail:   "The version field must be specified as a string. Ex: `version = \">= 1.2.0, < 2.0.0\". See https://www.packer.io/docs/templates/hcl_templates/blocks/packer#version-constraints for docs",
					Subject:  attr.Expr.Range().Ptr(),
				})
				continue
			}

			vc := VersionConstraint{
				DeclRange: attr.Range,
			}
			constraint := expr.GetAttr("version")
			if !constraint.Type().Equals(cty.String) || constraint.IsNull() {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid version constraint",
					Detail:   "Version must be specified as a string. See https://www.packer.io/docs/templates/hcl_templates/blocks/packer#version-constraint-syntax for docs.",
					Subject:  attr.Expr.Range().Ptr(),
				})
				continue
			}
			constraintStr := constraint.AsString()
			constraints, err := version.NewConstraint(constraintStr)
			if err != nil {
				// NewConstraint doesn't return user-friendly errors, so we'll just
				// ignore the provided error and produce our own generic one.
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid version constraint",
					Detail: "This string does not use correct version constraint syntax. " +
						"See https://www.packer.io/docs/templates/hcl_templates/blocks/packer#version-constraint-syntax for docs.\n" +
						err.Error(),
					Subject: attr.Expr.Range().Ptr(),
				})
				continue
			}
			vc.Required = constraints
			rp.Requirement = vc

			if !expr.Type().HasAttribute("source") {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No source was set",
					Detail:   "The source field must be specified as a string. Ex: `source = \"coolcloud\". See https://www.packer.io/docs/templates/hcl_templates/blocks/packer#specifying-plugin-requirements for docs",
					Subject:  attr.Expr.Range().Ptr(),
				})
				continue
			}
			source := expr.GetAttr("source")

			if !source.Type().Equals(cty.String) || source.IsNull() {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid source",
					Detail:   "Source must be specified as a string. For example: " + `source = "coolcloud"`,
					Subject:  attr.Expr.Range().Ptr(),
				})
				continue
			}

			rp.Source = source.AsString()
			p, sourceDiags := addrs.ParsePluginSourceString(rp.Source)

			if sourceDiags.HasErrors() {
				for _, diag := range sourceDiags {
					if diag.Subject == nil {
						diag.Subject = attr.Expr.Range().Ptr()
					}
				}
				diags = append(diags, sourceDiags...)
				continue
			} else {
				rp.Type = p
			}

			attrTypes := expr.Type().AttributeTypes()
			for name := range attrTypes {
				if name == "version" || name == "source" {
					continue
				}
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid required_plugins object",
					Detail:   `required_plugins objects can only contain "version" and "source" attributes.`,
					Subject:  attr.Expr.Range().Ptr(),
				})
				break
			}

		default:
			// should not happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid required_plugins syntax",
				Detail:   "required_plugins entries must be objects.",
				Subject:  attr.Expr.Range().Ptr(),
			})
		}

		if ret.RequiredPlugins == nil {
			ret.RequiredPlugins = make(map[string]*RequiredPlugin)
		}
		ret.RequiredPlugins[rp.Name] = rp
	}

	return ret, diags
}

// checkPluginNameNormalized verifies that the given string is already
// normalized and returns an error if not.
func checkPluginNameNormalized(name string, declrange hcl.Range) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// verify that the plugin local name is normalized
	normalized, err := addrs.IsPluginPartNormalized(name)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin local name",
			Detail:   fmt.Sprintf("%s is an invalid plugin local name: %s", name, err),
			Subject:  &declrange,
		})
		return diags
	}
	if !normalized {
		// we would have returned this error already
		normalizedPlugin, _ := addrs.ParsePluginPart(name)
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid plugin local name",
			Detail:   fmt.Sprintf("Plugin names must be normalized. Replace %q with %q to fix this error.", name, normalizedPlugin),
			Subject:  &declrange,
		})
	}
	return diags
}
