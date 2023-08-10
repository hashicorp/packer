// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package kvflag

import (
	"encoding/json"
	"fmt"
	"os"
)

// FlagJSON is a flag.Value implementation for parsing user variables
// from the command-line using JSON files.
type FlagJSON map[string]string

func (v *FlagJSON) String() string {
	return ""
}

func (v *FlagJSON) Set(raw string) error {
	f, err := os.Open(raw)
	if err != nil {
		return err
	}
	defer f.Close()

	if *v == nil {
		*v = make(map[string]string)
	}

	if err := json.NewDecoder(f).Decode(v); err != nil {
		return fmt.Errorf(
			"Error reading variables in '%s': %s", raw, err)
	}

	return nil
}
