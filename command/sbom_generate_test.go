// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/internal/sbom"
)

func TestSBOMGenerateCommand_ParseArgs_ExcludeAndScope(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := &SBOMGenerateCommand{
		Meta: Meta{
			Ui: &packersdk.BasicUi{
				Writer:      &out,
				ErrorWriter: &errOut,
			},
		},
	}

	cfg, ret := cmd.ParseArgs([]string{
		"--exclude", "/tmp/**",
		"--exclude=/var/cache/**",
		"--scope", "all-layers",
		"-o", "cyclonedx-json",
		"/",
	})
	if ret != 0 {
		t.Fatalf("expected parse success, got ret=%d err=%q", ret, errOut.String())
	}

	if cfg.Scope != sbom.ScopeAllLayers {
		t.Fatalf("expected scope %q, got %q", sbom.ScopeAllLayers, cfg.Scope)
	}

	if len(cfg.Exclude) != 2 {
		t.Fatalf("expected 2 exclude entries, got %d", len(cfg.Exclude))
	}
	if cfg.Exclude[0] != "/tmp/**" || cfg.Exclude[1] != "/var/cache/**" {
		t.Fatalf("unexpected exclude values: %#v", cfg.Exclude)
	}
}

func TestSBOMGenerateCommand_ParseArgs_InvalidScope(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := &SBOMGenerateCommand{
		Meta: Meta{
			Ui: &packersdk.BasicUi{
				Writer:      &out,
				ErrorWriter: &errOut,
			},
		},
	}

	_, ret := cmd.ParseArgs([]string{"--scope", "bad-scope"})
	if ret == 0 {
		t.Fatalf("expected parse failure for invalid scope")
	}
}

func TestSBOMGenerateCommand_ParseArgs_UnsupportedFlagIgnored(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := &SBOMGenerateCommand{
		Meta: Meta{
			Ui: &packersdk.BasicUi{
				Writer:      &out,
				ErrorWriter: &errOut,
			},
		},
	}

	cfg, ret := cmd.ParseArgs([]string{"-q", "/opt/app"})
	if ret != 0 {
		t.Fatalf("expected parse success, got ret=%d err=%q", ret, errOut.String())
	}

	if cfg.ScanPath != "/opt/app" {
		t.Fatalf("expected scan path /opt/app, got %q", cfg.ScanPath)
	}

	if out.Len() != 0 {
		t.Fatalf("expected no stdout output for ignored arg, got %q", out.String())
	}
}
