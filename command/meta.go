// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	kvflag "github.com/hashicorp/packer/command/flag-kv"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/helper/wrappedstreams"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/version"
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

// FlagSet returns a FlagSet with Packer SDK Ui support built-in
func (m *Meta) FlagSet(n string) *flag.FlagSet {
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
