// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

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
