// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package null

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestBuilder_implBuilder(t *testing.T) {
	var _ packersdk.Builder = new(Builder)
}
