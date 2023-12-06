// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"path/filepath"
	"testing"
)

func TestBuildWithCleanupScript(t *testing.T) {
	c := &BuildCommand{
		Meta: TestMetaFile(t),
	}

	args := []string{
		"-parallel-builds=1",
		filepath.Join(testFixture("cleanup-script"), "template.json"),
	}

	defer cleanup()

	// build should exit with error code!
	if code := c.Run(args); code == 0 {
		fatalCommand(t, c.Meta)
	}

	if !fileExists("ducky.txt") {
		t.Errorf("Expected to find ducky.txt")
	}

}
