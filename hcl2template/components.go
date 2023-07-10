// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcl2template

import "fmt"

// ComponentKind helps enumerate what kind of components exist in this Package.
type ComponentKind int

const (
	Builder ComponentKind = iota
	Provisioner
	PostProcessor
	Datasource
)

func (k ComponentKind) String() string {
	switch k {
	case Builder:
		return "builder"
	case Provisioner:
		return "provisioner"
	case PostProcessor:
		return "post-processor"
	case Datasource:
		return "datasource"
	}

	panic(fmt.Sprintf("unknown component kind %d, this is a Packer bug which should be reported as an issue.", k))
}
