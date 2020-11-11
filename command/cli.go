package command

import (
	"flag"
	"strings"

	"github.com/hashicorp/packer/helper/enumflag"
	kvflag "github.com/hashicorp/packer/helper/flag-kv"
	sliceflag "github.com/hashicorp/packer/helper/flag-slice"
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

// MetaArgs defines commonalities between all comands
type MetaArgs struct {
	// TODO(azr): in the future, I want to allow passing multiple path to
	// merge HCL confs together; but this will probably need an RFC first.
	Path         string
	Only, Except []string
	Vars         map[string]string
	VarFiles     []string
	// set to "hcl2" to force hcl2 mode
	ConfigType configType
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

	ba.MetaArgs.AddFlagSets(flags)
}

// BuildArgs represents a parsed cli line for a `packer build`
type BuildArgs struct {
	MetaArgs
	Color, Debug, Force, TimestampUi, MachineReadable bool
	ParallelBuilds                                    int64
	OnError                                           string
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

	va.MetaArgs.AddFlagSets(flags)
}

// ValidateArgs represents a parsed cli line for a `packer validate`
type ValidateArgs struct {
	MetaArgs
	SyntaxOnly bool
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

	va.MetaArgs.AddFlagSets(flags)
}

// HCL2UpgradeArgs represents a parsed cli line for a `packer hcl2_upgrade`
type HCL2UpgradeArgs struct {
	MetaArgs
	OutputFile string
}

func (va *FormatArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&va.Check, "check", false, "check if the input is formatted")
	flags.BoolVar(&va.Diff, "diff", false, "display the diff of formatting changes")
	flags.BoolVar(&va.Write, "write", true, "overwrite source files instead of writing to stdout")

	va.MetaArgs.AddFlagSets(flags)
}

// FormatArgs represents a parsed cli line for `packer fmt`
type FormatArgs struct {
	MetaArgs
	Check, Diff, Write bool
}
