// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package compress

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_ImplementsArtifact(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact should be a Artifact!")
	}
}
