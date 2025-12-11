// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package file

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packersdk.Artifact = new(FileArtifact)
}
