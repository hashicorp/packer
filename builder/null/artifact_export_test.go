// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package null

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packersdk.Artifact = new(NullArtifact)
}
