// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"math"
	"strings"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	sequential "github.com/hashicorp/packer/command/schedulers/sequential"

	"github.com/posener/complete"
)

type BuildCommand struct {
	Meta
}

func (c *BuildCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *BuildCommand) ParseArgs(args []string) (*BuildArgs, int) {
	var cfg BuildArgs
	flags := c.Meta.FlagSet("build", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	if cfg.ParallelBuilds < 1 {
		cfg.ParallelBuilds = math.MaxInt64
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.Path = args[0]
	return &cfg, 0
}

func writeDiags(ui packersdk.Ui, files map[string]*hcl.File, diags hcl.Diagnostics) int {
	// write HCL errors/diagnostics if any.
	b := bytes.NewBuffer(nil)
	err := hcl.NewDiagnosticTextWriter(b, files, 80, false).WriteDiagnostics(diags)
	if err != nil {
		ui.Error("could not write diagnostic: " + err.Error())
		return 1
	}
	if b.Len() != 0 {
		if diags.HasErrors() {
			ui.Error(b.String())
			return 1
		}
		ui.Say(b.String())
	}
	return 0
}

func (c *BuildCommand) RunContext(buildCtx context.Context, cla *BuildArgs) int {
	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	diags := packerStarter.DetectPluginBinaries()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	diags = packerStarter.Initialize()
	bundledDiags := c.DetectBundledPlugins(packerStarter)
	diags = append(bundledDiags, diags...)
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	scheduler := sequential.NewSequentialScheduler(packerStarter, cla.ToSchedulerOptions()).
		WithContext(buildCtx).
		WithUi(c.Meta.Ui).
		WithBuilds().
		WithHCPRegistry()

	diags = scheduler.Run()

	return writeDiags(c.Ui, nil, diags)
}

func (*BuildCommand) Help() string {
	helpText := `
Usage: packer build [options] TEMPLATE

  Will execute multiple builds in parallel as defined in the template.
  The various artifacts created by the template will be outputted.

Options:

  -color=false                  Disable color output. (Default: color)
  -debug                        Debug mode enabled for builds.
  -except=foo,bar,baz           Run all builds and post-processors other than these.
  -only=foo,bar,baz             Build only the specified builds.
  -force                        Force a build to continue if artifacts exist, deletes existing artifacts.
  -machine-readable             Produce machine-readable output.
  -on-error=[cleanup|abort|ask|run-cleanup-provisioner] If the build fails do: clean up (default), abort, ask, or run-cleanup-provisioner.
  -parallel-builds=1            Number of builds to run in parallel. 1 disables parallelization. 0 means no limit (Default: 0)
  -timestamp-ui                 Enable prefixing of each ui output with an RFC3339 timestamp.
  -var 'key=value'              Variable for templates, can be used multiple times.
  -var-file=path                JSON or HCL2 file containing user variables, can be used multiple times.
  -warn-on-undeclared-var       Display warnings for user variable files containing undeclared variables.
`

	return strings.TrimSpace(helpText)
}

func (*BuildCommand) Synopsis() string {
	return "build image(s) from template"
}

func (*BuildCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*BuildCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-color":            complete.PredictNothing,
		"-debug":            complete.PredictNothing,
		"-except":           complete.PredictNothing,
		"-only":             complete.PredictNothing,
		"-force":            complete.PredictNothing,
		"-machine-readable": complete.PredictNothing,
		"-on-error":         complete.PredictNothing,
		"-parallel":         complete.PredictNothing,
		"-timestamp-ui":     complete.PredictNothing,
		"-var":              complete.PredictNothing,
		"-var-file":         complete.PredictNothing,
	}
}
