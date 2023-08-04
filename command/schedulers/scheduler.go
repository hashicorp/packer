package schedulers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

type Scheduler interface {
	Run() hcl.Diagnostics
}

type SchedulerOptions struct {
	Only, Except []string

	// Build-specific options
	Debug, Force                        bool
	Color, TimestampUi, MachineReadable bool
	ParallelBuilds                      int64
	OnError                             string
}

func (so SchedulerOptions) toPackerBuildOpts() packer.GetBuildsOptions {
	opts := packer.GetBuildsOptions{
		Except:  so.Except,
		Only:    so.Only,
		Debug:   so.Debug,
		Force:   so.Force,
		OnError: so.OnError,
	}

	return opts
}
