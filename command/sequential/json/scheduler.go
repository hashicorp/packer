package json

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

type JSONSequentialScheduler struct {
	config *packer.Core
}

func NewScheduler(config *packer.Core) *JSONSequentialScheduler {
	return &JSONSequentialScheduler{
		config: config,
	}
}

// EvaluateDataSources is a noop in JSON as data sources are not supported in this mode.
func (s *JSONSequentialScheduler) ExecuteDataSources(bool) hcl.Diagnostics {
	return nil
}

// In a JSON template, there are no HCL files, so this is always nil.
func (s *JSONSequentialScheduler) FileMap() map[string]*hcl.File {
	return nil
}
