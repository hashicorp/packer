// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kvflag

import (
	"strings"
)

type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
