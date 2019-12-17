package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/packer/template"
)

type PackerV1Variables map[string]string

// decodeConfig decodes a "variables" section the way packer 1 used to
func (variables *PackerV1Variables) decodeConfig(block *hcl.Block) hcl.Diagnostics {
	return gohcl.DecodeBody(block.Body, nil, variables)
}

func (variables PackerV1Variables) Variables() map[string]*template.Variable {
	res := map[string]*template.Variable{}

	for k, v := range variables {
		res[k] = &template.Variable{
			Key:     k,
			Default: v,
		}
	}

	return res
}
