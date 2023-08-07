// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	multierror "github.com/hashicorp/go-multierror"
	version "github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	packerversion "github.com/hashicorp/packer/version"
)

// Core is the main executor of Packer. If Packer is being used as a
// library, this is the struct you'll want to instantiate to get anything done.
type Core struct {
	Template *template.Template

	Components ComponentFinder
	Variables  map[string]string
	Builds     map[string]*template.Builder
	Version    string
	Secrets    []string
}

// CoreConfig is the structure for initializing a new Core. Once a CoreConfig
// is used to initialize a Core, it shouldn't be re-used or modified again.
type CoreConfig struct {
	Components         ComponentFinder
	Template           *template.Template
	Variables          map[string]string
	SensitiveVariables []string
	Version            string
}

// The function type used to lookup Builder implementations.
type BuilderFunc func(name string) (packersdk.Builder, error)

// The function type used to lookup Hook implementations.
type HookFunc func(name string) (packersdk.Hook, error)

// The function type used to lookup PostProcessor implementations.
type PostProcessorFunc func(name string) (packersdk.PostProcessor, error)

// The function type used to lookup Provisioner implementations.
type ProvisionerFunc func(name string) (packersdk.Provisioner, error)

type BasicStore interface {
	Has(name string) bool
	List() (names []string)
}

type BuilderStore interface {
	BasicStore
	Start(name string) (packersdk.Builder, error)
}

type BuilderSet interface {
	BuilderStore
	Set(name string, starter func() (packersdk.Builder, error))
}

type ProvisionerStore interface {
	BasicStore
	Start(name string) (packersdk.Provisioner, error)
}

type ProvisionerSet interface {
	ProvisionerStore
	Set(name string, starter func() (packersdk.Provisioner, error))
}

type PostProcessorStore interface {
	BasicStore
	Start(name string) (packersdk.PostProcessor, error)
}

type PostProcessorSet interface {
	PostProcessorStore
	Set(name string, starter func() (packersdk.PostProcessor, error))
}

type DatasourceStore interface {
	BasicStore
	Start(name string) (packersdk.Datasource, error)
}

type DatasourceSet interface {
	DatasourceStore
	Set(name string, starter func() (packersdk.Datasource, error))
}

// ComponentFinder is a struct that contains the various function
// pointers necessary to look up components of Packer such as builders,
// commands, etc.
type ComponentFinder struct {
	Hook         HookFunc
	PluginConfig *PluginConfig
}

// NewCore creates a new Core.
func NewCore(c *CoreConfig) *Core {
	core := &Core{
		Template:   c.Template,
		Components: c.Components,
		Variables:  c.Variables,
		Version:    c.Version,
	}
	return core
}

// DetectPluginBinaries is used to load required plugins from the template,
// since it is unsupported in JSON, this is essentially a no-op.
func (c *Core) DetectPluginBinaries() hcl.Diagnostics {
	return nil
}

func (c *Core) Initialize() hcl.Diagnostics {
	err := c.initialize()
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}
	return nil
}

func (core *Core) initialize() error {
	if err := core.validate(); err != nil {
		return err
	}

	return nil
}

func (c *Core) PluginRequirements() (plugingetter.Requirements, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Summary:  "Packer plugins currently only works with HCL2 configuration templates",
			Detail:   "Please manually install plugins with the plugins command or use a HCL2 configuration that will do that for you.",
			Severity: hcl.DiagError,
		},
	}
}

// Context returns an interpolation context.
func (c *Core) Context() *interpolate.Context {
	return &interpolate.Context{
		TemplatePath:            c.Template.Path,
		UserVariables:           c.Variables,
		CorePackerVersionString: packerversion.FormattedVersion(),
	}
}

var ConsoleHelp = strings.TrimSpace(`
Packer console JSON Mode.
The Packer console allows you to experiment with Packer interpolations.
You may access variables in the Packer config you called the console with.

Type in the interpolation to test and hit <enter> to see the result.

"variables" will dump all available variables and their values.

"{{timestamp}}" will output the timestamp, for example "1559855090".

To exit the console, type "exit" and hit <enter>, or use Control-C.

/!\ If you would like to start console in hcl2 mode without a config you can
use the --config-type=hcl2 option.
`)

func (c *Core) EvaluateExpression(line string) (string, bool, hcl.Diagnostics) {
	switch {
	case line == "":
		return "", false, nil
	case line == "exit":
		return "", true, nil
	case line == "help":
		return ConsoleHelp, false, nil
	case line == "variables":
		varsstring := "\n"
		for k, v := range c.Context().UserVariables {
			varsstring += fmt.Sprintf("%s: %+v,\n", k, v)
		}

		return varsstring, false, nil
	default:
		ctx := c.Context()
		rendered, err := interpolate.Render(line, ctx)
		var diags hcl.Diagnostics
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary: "Interpolation error",
				Detail:  err.Error(),
			})
		}
		return rendered, false, diags
	}
}

func (c *Core) InspectConfig(opts InspectConfigOptions) int {

	// Convenience...
	ui := opts.Ui
	tpl := c.Template
	ui.Say("Packer Inspect: JSON mode")

	// Description
	if tpl.Description != "" {
		ui.Say("Description:\n")
		ui.Say(tpl.Description + "\n")
	}

	// Variables
	if len(tpl.Variables) == 0 {
		ui.Say("Variables:\n")
		ui.Say("  <No variables>")
	} else {
		requiredHeader := false
		for k, v := range tpl.Variables {
			for _, sensitive := range tpl.SensitiveVariables {
				if ok := strings.Compare(sensitive.Default, v.Default); ok == 0 {
					v.Default = "<sensitive>"
				}
			}
			if v.Required {
				if !requiredHeader {
					requiredHeader = true
					ui.Say("Required variables:\n")
				}

				ui.Machine("template-variable", k, v.Default, "1")
				ui.Say("  " + k)
			}
		}

		if requiredHeader {
			ui.Say("")
		}

		ui.Say("Optional variables and their defaults:\n")
		keys := make([]string, 0, len(tpl.Variables))
		max := 0
		for k := range tpl.Variables {
			keys = append(keys, k)
			if len(k) > max {
				max = len(k)
			}
		}

		sort.Strings(keys)

		for _, k := range keys {
			v := tpl.Variables[k]
			if v.Required {
				continue
			}
			for _, sensitive := range tpl.SensitiveVariables {
				if ok := strings.Compare(sensitive.Default, v.Default); ok == 0 {
					v.Default = "<sensitive>"
				}
			}

			padding := strings.Repeat(" ", max-len(k))
			output := fmt.Sprintf("  %s%s = %s", k, padding, v.Default)

			ui.Machine("template-variable", k, v.Default, "0")
			ui.Say(output)
		}
	}

	ui.Say("")

	// Builders
	ui.Say("Builders:\n")
	if len(tpl.Builders) == 0 {
		ui.Say("  <No builders>")
	} else {
		keys := make([]string, 0, len(tpl.Builders))
		max := 0
		for k := range tpl.Builders {
			keys = append(keys, k)
			if len(k) > max {
				max = len(k)
			}
		}

		sort.Strings(keys)

		for _, k := range keys {
			v := tpl.Builders[k]
			padding := strings.Repeat(" ", max-len(k))
			output := fmt.Sprintf("  %s%s", k, padding)
			if v.Name != v.Type {
				output = fmt.Sprintf("%s (%s)", output, v.Type)
			}

			ui.Machine("template-builder", k, v.Type)
			ui.Say(output)

		}
	}

	ui.Say("")

	// Provisioners
	ui.Say("Provisioners:\n")
	if len(tpl.Provisioners) == 0 {
		ui.Say("  <No provisioners>")
	} else {
		for _, v := range tpl.Provisioners {
			ui.Machine("template-provisioner", v.Type)
			ui.Say(fmt.Sprintf("  %s", v.Type))
		}
	}

	ui.Say("\nNote: If your build names contain user variables or template\n" +
		"functions such as 'timestamp', these are processed at build time,\n" +
		"and therefore only show in their raw form here.")

	return 0
}

func (c *Core) FixConfig(opts FixConfigOptions) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// Remove once we have support for the Inplace FixConfigMode
	if opts.Mode != Diff {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("FixConfig only supports template diff; FixConfigMode %d not supported", opts.Mode),
		})

		return diags
	}

	var rawTemplateData map[string]interface{}
	input := make(map[string]interface{})
	templateData := make(map[string]interface{})
	if err := json.Unmarshal(c.Template.RawContents, &rawTemplateData); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unable to read the contents of the JSON configuration file: %s", err),
			Detail:   err.Error(),
		})
		return diags
	}
	// Hold off on Diff for now - need to think about displaying to user.
	// delete empty top-level keys since the fixers seem to add them
	// willy-nilly
	for k := range input {
		ml, ok := input[k].([]map[string]interface{})
		if !ok {
			continue
		}
		if len(ml) == 0 {
			delete(input, k)
		}
	}
	// marshal/unmarshal to make comparable to templateData
	var fixedData map[string]interface{}
	// Guaranteed to be valid json, so we can ignore errors
	j, _ := json.Marshal(input)
	if err := json.Unmarshal(j, &fixedData); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unable to read the contents of the JSON configuration file: %s", err),
			Detail:   err.Error(),
		})

		return diags
	}

	if diff := cmp.Diff(templateData, fixedData); diff != "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Fixable configuration found.\nPlease run `packer fix` to get your build to run correctly.\nSee debug log for more information.",
			Detail:   diff,
		})
	}
	return diags
}

// validate does a full validation of the template.
//
// This will automatically call template.validate() in addition to doing
// richer semantic checks around variables and so on.
func (c *Core) validate() error {
	// First validate the template in general, we can't do anything else
	// unless the template itself is valid.
	if err := c.Template.Validate(); err != nil {
		return err
	}

	// Validate the minimum version is satisfied
	if c.Template.MinVersion != "" {
		versionActual, err := version.NewVersion(c.Version)
		if err != nil {
			// This shouldn't happen since we set it via the compiler
			panic(err)
		}

		versionMin, err := version.NewVersion(c.Template.MinVersion)
		if err != nil {
			return fmt.Errorf(
				"min_version is invalid: %s", err)
		}

		if versionActual.LessThan(versionMin) {
			return fmt.Errorf(
				"This template requires Packer version %s or higher; using %s",
				versionMin,
				versionActual)
		}
	}

	// Validate variables are set
	var err error
	for n, v := range c.Template.Variables {
		if v.Required {
			if _, ok := c.Variables[n]; !ok {
				err = multierror.Append(err, fmt.Errorf(
					"required variable not set: %s", n))
			}
		}
	}

	// TODO: validate all builders exist
	// TODO: ^^ provisioner
	// TODO: ^^ post-processor

	return err
}
