// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildScrubsSensitiveMultilineShellLocalOutput(t *testing.T) {
	templatePath := filepath.Join(testFixture("repro-sensitive-multiline"), testBuildSensitiveMultilineShellLocalFixture(runtime.GOOS))

	c := &BuildCommand{
		Meta: TestMetaFile(t),
	}

	if exitCode := c.Run([]string{templatePath}); exitCode != 0 {
		out, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
		t.Fatalf("build failed with exit code %d\nstdout: %q\nstderr: %q", exitCode, out, stderr)
	}

	out, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
	output := out + "\n" + stderr
	secret := "line-one-secret\nline-two-secret\nline-three-secret"

	if strings.Contains(output, secret) {
		t.Fatalf("multiline sensitive value leaked to build output: %q", output)
	}
	if strings.Contains(output, "line-one-secret") {
		t.Fatalf("sensitive line leaked to build output: %q", output)
	}
	if !strings.Contains(output, "<sensitive>") {
		t.Fatalf("expected scrubbed output, got: %q", output)
	}
}

func testBuildSensitiveMultilineShellLocalFixture(goos string) string {
	if goos == "windows" {
		return "multi-pwd.windows.pkr.hcl"
	}

	return "multi-pwd.unix.pkr.hcl"
}
