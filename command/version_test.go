// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestVersionCommand_implements(t *testing.T) {
	var _ cli.Command = &VersionCommand{}
}
