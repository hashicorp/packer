// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package enumflag

import "fmt"

type enumFlag struct {
	target  *string
	options []string
}

// New returns a flag.Value implementation for parsing flags with a one-of-a-set value
func New(target *string, options ...string) *enumFlag {
	return &enumFlag{target: target, options: options}
}

func (f *enumFlag) String() string {
	return *f.target
}

func (f *enumFlag) Set(value string) error {
	for _, v := range f.options {
		if v == value {
			*f.target = value
			return nil
		}
	}

	return fmt.Errorf("expected one of %q", f.options)
}
