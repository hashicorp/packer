// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"testing"
)

func TestFix_allFixersEnabled(t *testing.T) {
	f := Fixers
	o := FixerOrder

	if len(f) != len(o) {
		t.Fatalf("Fixers length (%d) does not match FixerOrder length (%d)", len(f), len(o))
	}

	for fixer := range f {
		found := false

		for _, orderedFixer := range o {
			if orderedFixer == fixer {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Did not find Fixer %s in FixerOrder", fixer)
		}
	}
}
