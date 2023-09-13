package schedulers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template"
)

type HCLSequentialScheduler struct {
	config *hcl2template.PackerConfig
}

func NewScheduler(config *hcl2template.PackerConfig) *HCLSequentialScheduler {
	return &HCLSequentialScheduler{
		config: config,
	}
}

func (s *HCLSequentialScheduler) FileMap() map[string]*hcl.File {
	return s.config.Files()
}
