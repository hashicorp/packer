// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildScrubsSensitiveMultilineShellLocalOutput(t *testing.T) {
	template := `variable "secret_multiline" {
					type      = string
					sensitive = true
					default = "line-one-secret\nline-two-secret\nline-three-secret"
				}

				source "null" "example" {
					communicator = "none"
				}

				build {
					sources = ["sources.null.example"]

					provisioner "shell-local" {
						inline = [
							"printf 'BEGIN\n%s\nEND\n' '${var.secret_multiline}'"
						]
					}
				}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "multi-pwd.pkr.hcl")
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

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
