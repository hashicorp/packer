package schedulers

import (
	opts "github.com/hashicorp/packer/command/schedulers/opts"
	"github.com/hashicorp/packer/hcl2template"
)

type HCLSequentialScheduler struct {
	config *hcl2template.PackerConfig
	opts   *opts.SchedulerOptions
}

func (s *HCLSequentialScheduler) Options() *opts.SchedulerOptions {
	return s.opts
}

func NewScheduler(config *hcl2template.PackerConfig, opts *opts.SchedulerOptions) *HCLSequentialScheduler {
	return &HCLSequentialScheduler{
		config: config,
		opts:   opts,
	}
}
