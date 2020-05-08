package packer

import "github.com/hashicorp/hcl/v2"

type GetBuildsOptions struct {
	// Get builds except the ones that match with except and with only the ones
	// that match with Only. When those are empty everything matches.
	Except, Only []string
}

type BuildGetter interface {
	// GetBuilds return all possible builds for a config. It also starts them.
	// TODO(azr): rename to builder starter ?
	GetBuilds(GetBuildsOptions) ([]Build, hcl.Diagnostics)
}

//go:generate enumer -type FixConfigMode
type FixConfigMode int

const (
	Stdout FixConfigMode = iota
	Inplace
	Diff
)

type FixConfigOptions struct {
	DiffOnly bool
}

type OtherInterfaceyMacOtherInterfaceFace interface {
	// FixConfig will output the config in a fixed manner.
	FixConfig(FixConfigOptions) hcl.Diagnostics
}
