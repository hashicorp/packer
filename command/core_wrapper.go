package command

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

// CoreWrapper wraps a packer.Core in order to have it's Initialize func return
// a diagnostic.
type CoreWrapper struct {
	*packer.Core
}

func (c *CoreWrapper) Initialize(_ packer.InitializeOptions) hcl.Diagnostics {
	err := c.Core.Initialize()
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}
	return nil
}
