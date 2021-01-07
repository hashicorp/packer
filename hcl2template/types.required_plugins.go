package hcl2template

import (
	"fmt"

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

	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case sourceLabel:
			// TODO
		}
	}
	return diags
}

// RequiredPlugin represents a declaration of a dependency on a particular
// Plugin version or source.
type RequiredPlugin struct {
	Name        string
	Source      string
	Type        *addrs.Plugin
	Requirement VersionConstraint
	DeclRange   hcl.Range
}

type RequiredPlugins struct {
	RequiredPlugins map[string]*RequiredPlugin
	DeclRange       hcl.Range
}

func decodeRequiredPluginsBlock(block *hcl.Block) (*RequiredPlugins, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	ret := &RequiredPlugins{
		RequiredPlugins: make(map[string]*RequiredPlugin),
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
			vc, reqDiags := decodeVersionConstraint(attr)
			diags = append(diags, reqDiags...)
			rp.Requirement = vc
			rp.Type, err = addrs.ParsePluginSourceString(name)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid plugin type",
					Detail:   fmt.Sprintf(`Invalid plugin type %q: %s"`, name, err),
				})
			}

		case expr.Type().IsObjectType():
			if expr.Type().HasAttribute("version") {
				vc := VersionConstraint{
					DeclRange: attr.Range,
				}
				constraint := expr.GetAttr("version")
				if !constraint.Type().Equals(cty.String) || constraint.IsNull() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid version constraint",
						Detail:   "Version must be specified as a string.",
						Subject:  attr.Expr.Range().Ptr(),
					})
				} else {
					constraintStr := constraint.AsString()
					constraints, err := version.NewConstraint(constraintStr)
					if err != nil {
						// NewConstraint doesn't return user-friendly errors, so we'll just
						// ignore the provided error and produce our own generic one.
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Invalid version constraint",
							Detail:   "This string does not use correct version constraint syntax.",
							Subject:  attr.Expr.Range().Ptr(),
						})
					} else {
						vc.Required = constraints
						rp.Requirement = vc
					}
				}
			}
			if expr.Type().HasAttribute("source") {
				source := expr.GetAttr("source")
				if !source.Type().Equals(cty.String) || source.IsNull() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid source",
						Detail:   "Source must be specified as a string.",
						Subject:  attr.Expr.Range().Ptr(),
					})
				} else {
					rp.Source = source.AsString()

					p, sourceDiags := addrs.ParsePluginSourceString(rp.Source)

					if sourceDiags.HasErrors() {
						for _, diag := range sourceDiags {
							if diag.Subject == nil {
								diag.Subject = attr.Expr.Range().Ptr()
							}
						}
						diags = append(diags, sourceDiags...)
					} else {
						rp.Type = p
					}
				}
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
