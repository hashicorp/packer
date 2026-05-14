package hcp_sbom

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
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
				ScannerArgs:    []string{"-o", "cyclonedx-json"},
				ExecuteCommand: "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with custom scan path",
			map[string]interface{}{
				"auto_generate": true,
				"scan_path":     "/opt/app",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:   true,
				ScanPath:       "/opt/app",
				ScannerArgs:    []string{"-o", "cyclonedx-json"},
				ExecuteCommand: "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with custom execute_command",
			map[string]interface{}{
				"auto_generate":   true,
				"execute_command": "{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			interpolate.Context{},
			&Config{
				AutoGenerate:   true,
				ScanPath:       "/",
				ScannerArgs:    []string{"-o", "cyclonedx-json"},
				ExecuteCommand: "{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"auto_generate with deprecated scanner_url (should warn but not fail)",
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
				ScannerArgs:    []string{"-o", "cyclonedx-json"},
				ExecuteCommand: "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"deprecated scanner_checksum with scanner_url (should warn but not fail)",
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
				ScannerArgs:     []string{"-o", "cyclonedx-json"},
				ExecuteCommand:  "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			},
			false,
			"",
		},
		{
			"deprecated scanner_checksum without scanner_url - should still error for clarity",
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
				ScannerArgs:      []string{"-o", "cyclonedx-json"},
				ExecuteCommand:   "chmod +x {{.Path}} && sudo {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
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
			"source mode with scanner fields - should succeed (allows toggling auto_generate)",
			map[string]interface{}{
				"source":            "sbom.json",
				"scanner_args":      []string{"-o", "json"},
				"scan_path":         "/opt/app",
				"execute_command":   "{{.Path}} sbom-generate {{.Args}}",
				"elevated_user":     "admin",
				"elevated_password": "password123",
			},
			interpolate.Context{},
			&Config{
				Source:           "sbom.json",
				ScannerArgs:      []string{"-o", "json"},
				ScanPath:         "/opt/app",
				ExecuteCommand:   "{{.Path}} sbom-generate {{.Args}}",
				ElevatedUser:     "admin",
				ElevatedPassword: "password123",
			},
			false,
			"",
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

func TestNormalizeScannerExecuteCommand(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "injects when path directly followed by args",
			in:   "chmod +x {{.Path}} && {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}",
			want: "chmod +x {{.Path}} && {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
		},
		{
			name: "injects when path directly followed by scan path",
			in:   "sudo {{.Path}} {{.ScanPath}} > {{.Output}}",
			want: "sudo {{.Path}} sbom-generate {{.ScanPath}} > {{.Output}}",
		},
		{
			name: "keeps command when sbom-generate already present",
			in:   "{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
			want: "{{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}",
		},
		{
			name: "does not modify non-scan invocation",
			in:   "chmod +x {{.Path}} && {{.Path}} version",
			want: "chmod +x {{.Path}} && {{.Path}} version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeScannerExecuteCommand(tt.in)
			if got != tt.want {
				t.Fatalf("unexpected normalized command:\nwant: %q\n got: %q", tt.want, got)
			}
		})
	}
}

func TestExpectedZipSHA256FromSums(t *testing.T) {
	tests := []struct {
		name        string
		sumsContent string
		fileName    string
		want        string
		wantErr     string
	}{
		{
			name: "matches standard sums line",
			sumsContent: strings.Join([]string{
				"1111111111111111111111111111111111111111111111111111111111111111  packer_1.12.0_linux_amd64.zip",
				"2222222222222222222222222222222222222222222222222222222222222222  packer_1.12.0_linux_arm64.zip",
			}, "\n"),
			fileName: "packer_1.12.0_linux_arm64.zip",
			want:     "2222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			name: "matches starred filename",
			sumsContent: strings.Join([]string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa *packer_1.12.0_windows_amd64.zip",
				"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb *packer_1.12.0_linux_amd64.zip",
			}, "\n"),
			fileName: "packer_1.12.0_windows_amd64.zip",
			want:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name: "rejects malformed checksum format",
			sumsContent: strings.Join([]string{
				"not-a-valid-sha256  packer_1.12.0_linux_amd64.zip",
			}, "\n"),
			fileName: "packer_1.12.0_linux_amd64.zip",
			wantErr:  "invalid SHA256 checksum format",
		},
		{
			name: "returns not found for missing file",
			sumsContent: strings.Join([]string{
				"1111111111111111111111111111111111111111111111111111111111111111  packer_1.12.0_linux_amd64.zip",
			}, "\n"),
			fileName: "packer_1.12.0_freebsd_amd64.zip",
			wantErr:  "not found in SHA256SUMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expectedZipSHA256FromSums(tt.sumsContent, tt.fileName)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected checksum: want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestIsValidSHA256Hex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "valid lowercase sha256",
			in:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			want: true,
		},
		{
			name: "valid uppercase sha256",
			in:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			want: true,
		},
		{
			name: "rejects non-hex characters",
			in:   "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			want: false,
		},
		{
			name: "rejects short length",
			in:   "abc123",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidSHA256Hex(tt.in); got != tt.want {
				t.Fatalf("unexpected result: want %t, got %t", tt.want, got)
			}
		})
	}
}

func TestFileSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.bin")
	content := []byte("packer checksum test payload")

	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("failed to write temp file: %s", err)
	}

	got, err := fileSHA256(path)
	if err != nil {
		t.Fatalf("unexpected error hashing file: %s", err)
	}

	wantBytes := sha256.Sum256(content)
	want := hex.EncodeToString(wantBytes[:])
	if got != want {
		t.Fatalf("unexpected hash: want %q, got %q", want, got)
	}
}

func TestChecksumMismatchDetection(t *testing.T) {
	fileName := "packer_1.12.0_linux_amd64.zip"
	sumsContent := "1111111111111111111111111111111111111111111111111111111111111111  " + fileName

	expected, err := expectedZipSHA256FromSums(sumsContent, fileName)
	if err != nil {
		t.Fatalf("unexpected error resolving expected checksum: %s", err)
	}

	actual := "2222222222222222222222222222222222222222222222222222222222222222"
	if strings.EqualFold(expected, actual) {
		t.Fatalf("expected checksum mismatch, but checksums compared equal")
	}
}
