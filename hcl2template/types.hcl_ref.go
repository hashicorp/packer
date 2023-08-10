// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
)

// HCL2Ref references to the source definition in configuration text file. It
// is used to tell were something was wrong, - like a warning or an error -
// long after it was parsed; allowing to give pointers as to where change/fix
// things in a file.
type HCL2Ref struct {
	// references
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
