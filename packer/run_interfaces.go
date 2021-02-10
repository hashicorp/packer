package packer

import (
	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

type GetBuildsOptions struct {
	// Get builds except the ones that match with except and with only the ones
	// that match with Only. When those are empty everything matches.
	Except, Only []string
	Debug, Force bool
	OnError      string
}

type BuildGetter interface {
	// GetBuilds return all possible builds for a config. It also starts all
	// builders.
	// TODO(azr): rename to builder starter ?
	GetBuilds(GetBuildsOptions) ([]packersdk.Build, hcl.Diagnostics)
}

type Evaluator interface {
	// EvaluateExpression is meant to be used in the `packer console` command.
	// It parses the input string and returns what needs to be displayed. In
	// case of an error the error should be displayed.
	EvaluateExpression(expr string) (output string, exit bool, diags hcl.Diagnostics)
}

type InitializeOptions struct {
	// When set, the execution of datasources will be skipped and the datasource will provide
	// a output spec that will be used for validation only.
	SkipDatasourcesExecution bool
}

// The Handler handles all Packer things. This interface reflects the Packer
// commands, ex: init, console ( evaluate ), fix config, inspect config, etc. To
// run a build we will start the builds and then the core of Packer handles
// execution.
type Handler interface {
	Initialize(InitializeOptions) hcl.Diagnostics
	// PluginRequirements returns the list of plugin Requirements from the
	// config file.
	PluginRequirements() (plugingetter.Requirements, hcl.Diagnostics)
	Evaluator
	BuildGetter
	ConfigFixer
	ConfigInspector
}

//go:generate enumer -type FixConfigMode
type FixConfigMode int

const (
	// Stdout will make FixConfig simply print what the config should be; it
	// will only work when a single file is passed.
	Stdout FixConfigMode = iota
	// Inplace fixes your files on the spot.
	Inplace
	// Diff shows a full diff.
	Diff
)

type FixConfigOptions struct {
	Mode FixConfigMode
}

type ConfigFixer interface {
	// FixConfig will output the config in a fixed manner.
	FixConfig(FixConfigOptions) hcl.Diagnostics
}

type InspectConfigOptions struct {
	packersdk.Ui
}

type ConfigInspector interface {
	// Inspect will output self inspection for a configuration
	InspectConfig(InspectConfigOptions) (ret int)
}
