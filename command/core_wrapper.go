package command

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

type CoreWrapper struct {
	*packer.Core
}

func (c *CoreWrapper) Initialize() hcl.Diagnostics {
	err := c.Core.Initialize()
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Detail: err.Error(),
			},
		}
	}
	return nil
}
