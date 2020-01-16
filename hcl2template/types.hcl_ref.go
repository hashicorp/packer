package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
)

// reference to the source definition in configuration text file
type HCL2Ref struct {
	// reference to the source definition in configuration text file
	DefRange     hcl.Range
	TypeRange    hcl.Range
	LabelsRanges []hcl.Range

	// remainder of unparsed body
	Rest hcl.Body
}

func newHCL2Ref(block *hcl.Block, rest hcl.Body) HCL2Ref {
	return HCL2Ref{
		Rest:         rest,
		DefRange:     block.DefRange,
		TypeRange:    block.TypeRange,
		LabelsRanges: block.LabelRanges,
	}
}
