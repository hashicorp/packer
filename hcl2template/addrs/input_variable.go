// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package addrs

// InputVariable is the address of an input variable.
type InputVariable struct {
	referenceable
	Name string
}

func (v InputVariable) String() string {
	return "var." + v.Name
}
