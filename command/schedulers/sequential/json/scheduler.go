package json

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/command/schedulers/opts"
	"github.com/hashicorp/packer/packer"
)

type JSONSequentialScheduler struct {
	config *packer.Core
	opts   *opts.SchedulerOptions
}

func NewScheduler(config *packer.Core, opts *opts.SchedulerOptions) *JSONSequentialScheduler {
	return &JSONSequentialScheduler{
		config: config,
		opts:   opts,
	}
}

func (s *JSONSequentialScheduler) Options() *opts.SchedulerOptions {
	return s.opts
}

// EvaluateDataSources is a noop in JSON as data sources are not supported in this mode.
func (s *JSONSequentialScheduler) EvaluateDataSources() hcl.Diagnostics {
	return nil
}
