// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	kvflag "github.com/hashicorp/packer/command/flag-kv"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/helper/wrappedstreams"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/version"
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet
type FlagSetFlags uint

const (
	FlagSetNone        FlagSetFlags = 0
	FlagSetBuildFilter FlagSetFlags = 1 << iota
	FlagSetVars
)

// Meta contains the meta-options and functionality that nearly every
// Packer command inherits.
type Meta struct {
	CoreConfig *packer.CoreConfig
	Ui         packersdk.Ui
	Version    string
}

// Core returns the core for the given template given the configured
// CoreConfig and user variables on this Meta.
func (m *Meta) Core(tpl *template.Template, cla *MetaArgs) (*packer.Core, error) {
	// Copy the config so we don't modify it
	config := *m.CoreConfig
	config.Template = tpl

	fj := &kvflag.FlagJSON{}
	// First populate fj with contents from var files
	for _, file := range cla.VarFiles {
		err := fj.Set(file)
		if err != nil {
			return nil, err
		}
	}
	// Now read fj values back into flagvars and set as config.Variables. Only
	// add to flagVars if the key doesn't already exist, because flagVars comes
	// from the command line and should not be overridden by variable files.
	if cla.Vars == nil {
		cla.Vars = map[string]string{}
	}
	for k, v := range *fj {
		if _, exists := cla.Vars[k]; !exists {
			cla.Vars[k] = v
		}
	}
	config.Variables = cla.Vars

	core := packer.NewCore(&config)
	return core, nil
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// build settings on the commands that don't handle builds.
func (m *Meta) FlagSet(n string, _ FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// Create an io.Writer that writes to our Ui properly for errors.
	// This is kind of a hack, but it does the job. Basically: create
	// a pipe, use a scanner to break it into lines, and output each line
	// to the UI. Do this forever.
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.Ui.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	return f
}

// ValidateFlags should be called after parsing flags to validate the
// given flags
func (m *Meta) ValidateFlags() error {
	// TODO
	return nil
}

// StdinPiped returns true if the input is piped.
func (m *Meta) StdinPiped() bool {
	fi, err := wrappedstreams.Stdin().Stat()
	if err != nil {
		// If there is an error, let's just say its not piped
		return false
	}

	return fi.Mode()&os.ModeNamedPipe != 0
}

func (m *Meta) GetConfig(cla *MetaArgs) (packer.Handler, int) {
	cfgType, err := cla.GetConfigType()
	if err != nil {
		m.Ui.Error(fmt.Sprintf("%q: %s", cla.Path, err))
		return nil, 1
	}

	switch cfgType {
	case ConfigTypeHCL2:
		packer.CheckpointReporter.SetTemplateType(packer.HCL2Template)
		// TODO(azr): allow to pass a slice of files here.
		return m.GetConfigFromHCL(cla)
	default:
		packer.CheckpointReporter.SetTemplateType(packer.JSONTemplate)
		// TODO: uncomment once we've polished HCL a bit more.
		// c.Ui.Say(`Legacy JSON Configuration Will Be Used.
		// The template will be parsed in the legacy configuration style. This style
		// will continue to work but users are encouraged to move to the new style.
		// See: https://packer.io/guides/hcl
		// `)
		return m.GetConfigFromJSON(cla)
	}
}

func (m *Meta) GetConfigFromHCL(cla *MetaArgs) (*hcl2template.PackerConfig, int) {
	parser := &hcl2template.Parser{
		CorePackerVersion:       version.SemVer,
		CorePackerVersionString: version.FormattedVersion(),
		Parser:                  hclparse.NewParser(),
		PluginConfig:            m.CoreConfig.Components.PluginConfig,
		ValidationOptions: hcl2template.ValidationOptions{
			WarnOnUndeclaredVar: cla.WarnOnUndeclaredVar,
		},
	}
	cfg, diags := parser.Parse(cla.Path, cla.VarFiles, cla.Vars)
	return cfg, writeDiags(m.Ui, parser.Files(), diags)
}

func (m *Meta) GetConfigFromJSON(cla *MetaArgs) (packer.Handler, int) {
	// Parse the template
	var tpl *template.Template
	var err error
	if cla.Path == "" {
		// here cla validation passed so this means we want a default builder
		// and we probably are in the console command
		tpl, err = template.Parse(TiniestBuilder)
	} else {
		tpl, err = template.ParseFile(cla.Path)
	}

	if err != nil {
		m.Ui.Error(fmt.Sprintf("Failed to parse file as legacy JSON template: "+
			"if you are using an HCL template, check your file extensions; they "+
			"should be either *.pkr.hcl or *.pkr.json; see the docs for more "+
			"details: https://www.packer.io/docs/templates/hcl_templates. \n"+
			"Original error: %s", err))
		return nil, 1
	}

	// Get the core
	core, err := m.Core(tpl, cla)
	ret := 0
	if err != nil {
		m.Ui.Error(err.Error())
		ret = 1
	}
	return core, ret
}

func (m *Meta) DetectBundledPlugins(handler packer.Handler) hcl.Diagnostics {
	var plugins []string

	switch h := handler.(type) {
	case *packer.Core:
		plugins = m.detectBundledPluginsJSON(h)
	case *hcl2template.PackerConfig:
		plugins = m.detectBundledPluginsHCL2(handler.(*hcl2template.PackerConfig))
	}

	if len(plugins) == 0 {
		return nil
	}

	packer.CheckpointReporter.SetBundledUsage()

	buf := &strings.Builder{}
	buf.WriteString("This template relies on the use of plugins bundled into the Packer binary.\n")
	buf.WriteString("The practice of bundling external plugins into Packer will be removed in an upcoming version.\n\n")
	switch h := handler.(type) {
	case *packer.Core:
		buf.WriteString("To remove this warning and ensure builds keep working you can install these external plugins with the 'packer plugins install' command\n\n")

		for _, plugin := range plugins {
			fmt.Fprintf(buf, "* packer plugins install %s\n", plugin)
		}

		buf.WriteString("\nAlternatively, if you upgrade your templates to HCL2, you can use 'packer init' with a 'required_plugins' block to automatically install external plugins.\n\n")
		fmt.Fprintf(buf, "You can try HCL2 by running 'packer hcl2_upgrade %s'", h.Template.Path)
	case *hcl2template.PackerConfig:
		buf.WriteString("To remove this warning, add the following section to your template:\n")
		buf.WriteString(m.fixRequiredPlugins(h))
		buf.WriteString("\nThen run 'packer init' to manage installation of the plugins")
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Bundled plugins used",
			Detail:   buf.String(),
		},
	}
}

func (m *Meta) detectBundledPluginsJSON(core *packer.Core) []string {
	bundledPlugins := map[string]struct{}{}

	tmpl := core.Template
	if tmpl == nil {
		panic("No template parsed. This is a Packer bug which should be reported, please open an issue on the project's issue tracker.")
	}

	for _, b := range tmpl.Builders {
		builderType := fmt.Sprintf("packer-builder-%s", b.Type)
		if bundledStatus[builderType] {
			bundledPlugins[builderType] = struct{}{}
		}
	}

	for _, p := range tmpl.Provisioners {
		provisionerType := fmt.Sprintf("packer-provisioner-%s", p.Type)
		if bundledStatus[provisionerType] {
			bundledPlugins[provisionerType] = struct{}{}
		}
	}

	for _, pps := range tmpl.PostProcessors {
		for _, pp := range pps {
			postProcessorType := fmt.Sprintf("packer-post-processor-%s", pp.Type)
			if bundledStatus[postProcessorType] {
				bundledPlugins[postProcessorType] = struct{}{}
			}
		}
	}

	return compileBundledPluginList(bundledPlugins)
}

var knownPluginPrefixes = map[string]string{
	"amazon":        "github.com/hashicorp/amazon",
	"ansible":       "github.com/hashicorp/ansible",
	"azure":         "github.com/hashicorp/azure",
	"docker":        "github.com/hashicorp/docker",
	"googlecompute": "github.com/hashicorp/googlecompute",
	"qemu":          "github.com/hashicorp/qemu",
	"vagrant":       "github.com/hashicorp/vagrant",
	"vmware":        "github.com/hashicorp/vmware",
	"vsphere":       "github.com/hashicorp/vsphere",
}

func (m *Meta) fixRequiredPlugins(config *hcl2template.PackerConfig) string {
	plugins := map[string]struct{}{}

	for _, b := range config.Builds {
		for _, b := range b.Sources {
			for prefix, plugin := range knownPluginPrefixes {
				if strings.HasPrefix(b.Type, prefix) {
					plugins[plugin] = struct{}{}
				}
			}
		}

		for _, p := range b.ProvisionerBlocks {
			for prefix, plugin := range knownPluginPrefixes {
				if strings.HasPrefix(p.PType, prefix) {
					plugins[plugin] = struct{}{}
				}
			}
		}

		for _, pps := range b.PostProcessorsLists {
			for _, pp := range pps {
				for prefix, plugin := range knownPluginPrefixes {
					if strings.HasPrefix(pp.PType, prefix) {
						plugins[plugin] = struct{}{}
					}
				}
			}
		}
	}

	for _, ds := range config.Datasources {
		for prefix, plugin := range knownPluginPrefixes {
			if strings.HasPrefix(ds.Type, prefix) {
				plugins[plugin] = struct{}{}
			}
		}
	}

	retPlugins := make([]string, 0, len(plugins))
	for plugin := range plugins {
		retPlugins = append(retPlugins, plugin)
	}

	return generateRequiredPluginsBlock(retPlugins)
}

func (m *Meta) detectBundledPluginsHCL2(config *hcl2template.PackerConfig) []string {
	bundledPlugins := map[string]struct{}{}

	for _, b := range config.Builds {
		for _, src := range b.Sources {
			builderType := fmt.Sprintf("packer-builder-%s", src.Type)
			if bundledStatus[builderType] {
				bundledPlugins[builderType] = struct{}{}
			}
		}

		for _, p := range b.ProvisionerBlocks {
			provisionerType := fmt.Sprintf("packer-provisioner-%s", p.PType)
			if bundledStatus[provisionerType] {
				bundledPlugins[provisionerType] = struct{}{}
			}
		}

		for _, pps := range b.PostProcessorsLists {
			for _, pp := range pps {
				postProcessorType := fmt.Sprintf("packer-post-processor-%s", pp.PType)
				if bundledStatus[postProcessorType] {
					bundledPlugins[postProcessorType] = struct{}{}
				}
			}
		}
	}

	for _, ds := range config.Datasources {
		dsType := fmt.Sprintf("packer-datasource-%s", ds.Type)
		if bundledStatus[dsType] {
			bundledPlugins[dsType] = struct{}{}
		}
	}

	return compileBundledPluginList(bundledPlugins)
}
