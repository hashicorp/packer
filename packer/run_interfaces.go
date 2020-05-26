package packer

import "github.com/hashicorp/hcl/v2"

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
	GetBuilds(GetBuildsOptions) ([]Build, hcl.Diagnostics)
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
	DiffOnly bool
}

type ConfigFixer interface {
	// FixConfig will output the config in a fixed manner.
	FixConfig(FixConfigOptions) hcl.Diagnostics
}
