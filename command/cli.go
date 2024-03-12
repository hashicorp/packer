// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"flag"
	"strings"

	"github.com/hashicorp/packer/command/enumflag"
	kvflag "github.com/hashicorp/packer/command/flag-kv"
	sliceflag "github.com/hashicorp/packer/command/flag-slice"
)

//go:generate enumer -type configType -trimprefix ConfigType -transform snake
type configType int

const (
	ConfigTypeJSON configType = iota // default config type
	ConfigTypeHCL2
)

func (c *configType) Set(value string) error {
	v, err := configTypeString(value)
	if err == nil {
		*c = v
	}
	return err
}

// ConfigType tells what type of config we should use, it can return values
// like "hcl" or "json".
// Make sure Args was correctly set before.
func (ma *MetaArgs) GetConfigType() (configType, error) {
	if ma.Path == "" {
		return ma.ConfigType, nil
	}
	name := ma.Path
	if name == "-" {
		// TODO(azr): To allow piping HCL2 confs (when args is "-"), we probably
		// will need to add a setting that says "this is an HCL config".
		return ma.ConfigType, nil
	}
	if strings.HasSuffix(name, ".pkr.hcl") ||
		strings.HasSuffix(name, ".pkr.json") {
		return ConfigTypeHCL2, nil
	}
	isDir, err := isDir(name)
	if isDir {
		return ConfigTypeHCL2, err
	}
	return ma.ConfigType, err
}

// NewMetaArgs parses cli args and put possible values
func (ma *MetaArgs) AddFlagSets(fs *flag.FlagSet) {
	fs.Var((*sliceflag.StringFlag)(&ma.Only), "only", "")
	fs.Var((*sliceflag.StringFlag)(&ma.Except), "except", "")
	fs.Var((*kvflag.Flag)(&ma.Vars), "var", "")
	fs.Var((*kvflag.StringSlice)(&ma.VarFiles), "var-file", "")
	fs.Var(&ma.ConfigType, "config-type", "set to 'hcl2' to run in hcl2 mode when no file is passed.")
}

// MetaArgs defines commonalities between all commands
type MetaArgs struct {
	// TODO(azr): in the future, I want to allow passing multiple path to
	// merge HCL confs together; but this will probably need an RFC first.
	Path         string
	Only, Except []string
	Vars         map[string]string
	VarFiles     []string
	// set to "hcl2" to force hcl2 mode
	ConfigType configType

	// WarnOnUndeclared does not have a common default, as the default varies per sub-command usage.
	// Refer to individual command FlagSets for usage.
	WarnOnUndeclaredVar bool
}

func (ba *BuildArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&ba.Color, "color", true, "")
	flags.BoolVar(&ba.Debug, "debug", false, "")
	flags.BoolVar(&ba.Force, "force", false, "")
	flags.BoolVar(&ba.TimestampUi, "timestamp-ui", false, "")
	flags.BoolVar(&ba.MachineReadable, "machine-readable", false, "")

	flags.Int64Var(&ba.ParallelBuilds, "parallel-builds", 0, "")

	flagOnError := enumflag.New(&ba.OnError, "cleanup", "abort", "ask", "run-cleanup-provisioner")
	flags.Var(flagOnError, "on-error", "")

	flags.BoolVar(&ba.MetaArgs.WarnOnUndeclaredVar, "warn-on-undeclared-var", false, "Show warnings for variable files containing undeclared variables.")

	flags.BoolVar(&ba.ReleaseOnly, "ignore-prerelease-plugins", false, "Disable the loading of prerelease plugin binaries (x.y.z-dev).")

	ba.MetaArgs.AddFlagSets(flags)
}

// BuildArgs represents a parsed cli line for a `packer build`
type BuildArgs struct {
	MetaArgs
	Debug, Force                        bool
	Color, TimestampUi, MachineReadable bool
	ParallelBuilds                      int64
	OnError                             string
	ReleaseOnly                         bool
}

func (ia *InitArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&ia.Upgrade, "upgrade", false, "upgrade any present plugin to the highest allowed version.")
	flags.BoolVar(&ia.Force, "force", false, "force installation of a plugin, even if already installed")

	ia.MetaArgs.AddFlagSets(flags)
}

// InitArgs represents a parsed cli line for a `packer init <path>`
type InitArgs struct {
	MetaArgs
	Upgrade bool
	Force   bool
}

// PluginsRequiredArgs represents a parsed cli line for a `packer plugins required <path>`
type PluginsRequiredArgs struct {
	MetaArgs
}

// ConsoleArgs represents a parsed cli line for a `packer console`
type ConsoleArgs struct {
	MetaArgs
}

func (fa *FixArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&fa.Validate, "validate", true, "")

	fa.MetaArgs.AddFlagSets(flags)
}

// FixArgs represents a parsed cli line for a `packer fix`
type FixArgs struct {
	MetaArgs
	Validate bool
}

func (va *ValidateArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&va.SyntaxOnly, "syntax-only", false, "check syntax only")
	flags.BoolVar(&va.NoWarnUndeclaredVar, "no-warn-undeclared-var", false, "Ignore warnings for variable files containing undeclared variables.")
	flags.BoolVar(&va.EvaluateDatasources, "evaluate-datasources", false, "evaluate datasources for validation (HCL2 only, may incur costs)")
	flags.BoolVar(&va.ReleaseOnly, "ignore-prerelease-plugins", false, "Disable the loading of prerelease plugin binaries (x.y.z-dev).")

	va.MetaArgs.AddFlagSets(flags)
}

// ValidateArgs represents a parsed cli line for a `packer validate`
type ValidateArgs struct {
	MetaArgs
	SyntaxOnly, NoWarnUndeclaredVar bool
	EvaluateDatasources             bool
	ReleaseOnly                     bool
}

func (va *InspectArgs) AddFlagSets(flags *flag.FlagSet) {
	va.MetaArgs.AddFlagSets(flags)
}

// InspectArgs represents a parsed cli line for a `packer inspect`
type InspectArgs struct {
	MetaArgs
}

func (va *HCL2UpgradeArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.StringVar(&va.OutputFile, "output-file", "", "File where to put the hcl2 generated config. Defaults to JSON_TEMPLATE.pkr.hcl")
	flags.BoolVar(&va.WithAnnotations, "with-annotations", false, "Adds helper annotations with information about the generated HCL2 blocks.")

	va.MetaArgs.AddFlagSets(flags)
}

// HCL2UpgradeArgs represents a parsed cli line for a `packer hcl2_upgrade`
type HCL2UpgradeArgs struct {
	MetaArgs
	OutputFile      string
	WithAnnotations bool
}

func (va *FormatArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&va.Check, "check", false, "check if the input is formatted")
	flags.BoolVar(&va.Diff, "diff", false, "display the diff of formatting changes")
	flags.BoolVar(&va.Write, "write", true, "overwrite source files instead of writing to stdout")
	flags.BoolVar(&va.Recursive, "recursive", false, "Also process files in subdirectories")
	va.MetaArgs.AddFlagSets(flags)
}

// FormatArgs represents a parsed cli line for `packer fmt`
type FormatArgs struct {
	MetaArgs
	Check, Diff, Write, Recursive bool
}
