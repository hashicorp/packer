package hcp_sbom

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestConfigPrepare(t *testing.T) {
	tests := []struct {
		name               string
		inputConfig        map[string]interface{}
		interpolateContext interpolate.Context
		expectConfig       *Config
		expectError        bool
		errorContains      string
	}{
		{
			"empty config, should error without a source",
			map[string]interface{}{},
			interpolate.Context{},
			nil,
			true,
			"source must be specified",
		},
		{
			"config with full context for interpolation: success",
			map[string]interface{}{
				"source": "{{ .Name }}",
			},
			interpolate.Context{
				Data: &struct {
					Name string
				}{
					Name: "testInterpolate",
				},
			},
			&Config{
				Source: "testInterpolate",
			},
			false,
			"",
		},
		{
			// Note: this will look weird to reviewers, but is actually
			// expected for the moment.
			// Refer to the comment in `Prepare` for context as to WHY
			// this cannot be considered an error.
			"config with sbom name as interpolated value, without it in context, replace with a placeholder",
			map[string]interface{}{
				"source":    "test",
				"sbom_name": "{{ .Name }}",
			},
			interpolate.Context{},
			&Config{
				Source:   "test",
				SbomName: "<no value>",
			},
			false,
			"",
		},
		{
			"auto_generate enabled with defaults",
			map[string]interface{}{
				"auto_generate": true,
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:   true,
				ScanPath:       "/",
				ScannerArgs:    []string{"-o", "cyclonedx-json", "-q"},
				ExecuteCommand: "chmod +x {{.Path}} && sudo {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with custom scanner URL",
			map[string]interface{}{
				"auto_generate": true,
				"scanner_url":   "https://example.com/scanner",
				"scan_path":     "/opt/app",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:   true,
				ScannerURL:     "https://example.com/scanner",
				ScanPath:       "/opt/app",
				ScannerArgs:    []string{"-o", "cyclonedx-json", "-q"},
				ExecuteCommand: "chmod +x {{.Path}} && sudo {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with scanner checksum and URL",
			map[string]interface{}{
				"auto_generate":    true,
				"scanner_url":      "https://example.com/scanner",
				"scanner_checksum": "abc123def456",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:    true,
				ScannerURL:      "https://example.com/scanner",
				ScannerChecksum: "abc123def456",
				ScanPath:        "/",
				ScannerArgs:     []string{"-o", "cyclonedx-json", "-q"},
				ExecuteCommand:  "chmod +x {{.Path}} && sudo {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with custom execute_command",
			map[string]interface{}{
				"auto_generate":   true,
				"execute_command": "{{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:   true,
				ScanPath:       "/",
				ScannerArgs:    []string{"-o", "cyclonedx-json", "-q"},
				ExecuteCommand: "{{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with elevated user and password",
			map[string]interface{}{
				"auto_generate":     true,
				"elevated_user":     "admin",
				"elevated_password": "password123",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:     true,
				ElevatedUser:     "admin",
				ElevatedPassword: "password123",
				ScanPath:         "/",
				ScannerArgs:      []string{"-o", "cyclonedx-json", "-q"},
				ExecuteCommand:   "chmod +x {{.Path}} && sudo {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"source and auto_generate both set - should error",
			map[string]interface{}{
				"source":        "sbom.json",
				"auto_generate": true,
			},
			interpolate.Context{},
			nil,
			true,
			"source and auto_generate are mutually exclusive",
		},
		{
			"scanner_checksum without scanner_url - should error",
			map[string]interface{}{
				"auto_generate":    true,
				"scanner_checksum": "abc123",
			},
			interpolate.Context{},
			nil,
			true,
			"scanner_checksum requires scanner_url",
		},
		{
			"elevated_password without elevated_user - should error",
			map[string]interface{}{
				"auto_generate":     true,
				"elevated_password": "password123",
			},
			interpolate.Context{},
			nil,
			true,
			"elevated_user must be specified if elevated_password is provided",
		},
		{
			"scanner_url without auto_generate - should error",
			map[string]interface{}{
				"source":      "sbom.json",
				"scanner_url": "https://example.com/scanner",
			},
			interpolate.Context{},
			nil,
			true,
			"scanner_url can only be used when auto_generate is enabled",
		},
		{
			"scanner_checksum without auto_generate - should error",
			map[string]interface{}{
				"source":           "sbom.json",
				"scanner_checksum": "abc123",
			},
			interpolate.Context{},
			nil,
			true,
			"scanner_checksum can only be used when auto_generate is enabled",
		},
		{
			"scanner_args without auto_generate - should error",
			map[string]interface{}{
				"source":       "sbom.json",
				"scanner_args": []string{"-o", "json"},
			},
			interpolate.Context{},
			nil,
			true,
			"scanner_args can only be used when auto_generate is enabled",
		},
		{
			"scan_path without auto_generate - should error",
			map[string]interface{}{
				"source":    "sbom.json",
				"scan_path": "/opt/app",
			},
			interpolate.Context{},
			nil,
			true,
			"scan_path can only be used when auto_generate is enabled",
		},
		{
			"execute_command without auto_generate - should error",
			map[string]interface{}{
				"source":          "sbom.json",
				"execute_command": "{{.Path}} {{.Args}}",
			},
			interpolate.Context{},
			nil,
			true,
			"execute_command can only be used when auto_generate is enabled",
		},
		{
			"elevated_user without auto_generate - should error",
			map[string]interface{}{
				"source":        "sbom.json",
				"elevated_user": "admin",
			},
			interpolate.Context{},
			nil,
			true,
			"elevated_user can only be used when auto_generate is enabled",
		},
		{
			"elevated_password without auto_generate - should error",
			map[string]interface{}{
				"source":            "sbom.json",
				"elevated_password": "password123",
			},
			interpolate.Context{},
			nil,
			true,
			"elevated_password can only be used when auto_generate is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov := &Provisioner{}
			prov.config.ctx = tt.interpolateContext
			err := prov.Prepare(tt.inputConfig)
			if err != nil && !tt.expectError {
				t.Fatalf("configuration unexpectedly failed to prepare: %s", err)
			}

			if err == nil && tt.expectError {
				t.Fatalf("configuration succeeded to prepare, but should have failed")
			}

			if err != nil {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %s", tt.errorContains, err)
				}
				t.Logf("config had error %q", err)
				return
			}

			diff := cmp.Diff(prov.config, *tt.expectConfig, cmpopts.IgnoreUnexported(Config{}))
			if diff != "" {
				t.Errorf("configuration returned by `Prepare` is different from what was expected: %s", diff)
			}
		})
	}
}
