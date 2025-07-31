// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package env

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_IsHCPDisabled(t *testing.T) {
	tcs := []struct {
		name           string
		registry_value string
		output         bool
	}{
		{
			name:           "nothing set",
			registry_value: "",
			output:         false,
		},
		{
			name:           "registry set with 1",
			registry_value: "1",
			output:         false,
		},
		{
			name:           "registry set with 0",
			registry_value: "0",
			output:         true,
		},
		{
			name:           "registry set with OFF",
			registry_value: "OFF",
			output:         true,
		},
		{
			name:           "registry set with off",
			registry_value: "off",
			output:         true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(HCPPackerRegistry, tc.registry_value)
			out := IsHCPDisabled()
			if out != tc.output {
				t.Fatalf("unexpected output: %t", out)
			}
		})
	}
}
func Test_HasHCPAuth(t *testing.T) {
	origClientID := os.Getenv(HCPClientID)
	origClientSecret := os.Getenv(HCPClientSecret)
	origCredFile := os.Getenv(HCPCredFile)
	origDefaultCredFilePath := ""

	// Save and restore default cred file at ~/.config/hcp/cred_file.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	credDir := filepath.Join(homeDir, HCPDefaultCredFilePath)
	defaultCredPath := filepath.Join(credDir, HCPDefaultCredFile)

	origDefaultCredFileExists := false
	if _, err := os.Stat(defaultCredPath); err == nil {
		tmpFile, err := os.CreateTemp("", "orig_cred_file.json")
		if err != nil {
			t.Fatalf("failed to create temp file for original cred file: %v", err)
		}
		tmpFile.Close()
		origDefaultCredFilePath = tmpFile.Name()
		if err := os.Rename(defaultCredPath, origDefaultCredFilePath); err != nil {
			t.Fatalf("failed to move original cred file: %v", err)
		}
	}
	if _, err := os.ReadFile(defaultCredPath); err == nil {
		origDefaultCredFileExists = true
	}
	type setupFunc func(t *testing.T)

	tmpCredFile := func(t *testing.T) string {
		f, err := os.CreateTemp("", "cred_file.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		f.Close()
		t.Cleanup(func() { os.Remove(f.Name()) })
		return f.Name()
	}

	tmpDefaultCredFile := func(t *testing.T) string {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get home dir: %v", err)
		}
		credDir := filepath.Join(homeDir, HCPDefaultCredFilePath)
		os.MkdirAll(credDir, 0755)
		credPath := filepath.Join(credDir, HCPDefaultCredFile)
		f, err := os.Create(credPath)
		if err != nil {
			t.Fatalf("failed to create default cred file: %v", err)
		}
		f.Close()
		t.Cleanup(func() { os.Remove(credPath) })
		return credPath
	}

	tcs := []struct {
		name    string
		setup   setupFunc
		want    bool
		wantErr bool
	}{
		{
			name: "neither credentials nor certificate present",
			setup: func(t *testing.T) {
				os.Unsetenv(HCPClientID)
				os.Unsetenv(HCPClientSecret)
				os.Unsetenv(HCPCredFile)
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "only credentials present",
			setup: func(t *testing.T) {
				os.Unsetenv(HCPCredFile)
				os.Setenv(HCPClientID, "foo")
				os.Setenv(HCPClientSecret, "bar")
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "only certificate present via env var",
			setup: func(t *testing.T) {
				os.Unsetenv(HCPClientID)
				os.Unsetenv(HCPClientSecret)
				os.Setenv(HCPCredFile, tmpCredFile(t))
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "only certificate present via default path",
			setup: func(t *testing.T) {
				os.Unsetenv(HCPClientID)
				os.Unsetenv(HCPClientSecret)
				os.Unsetenv(HCPCredFile)
				tmpDefaultCredFile(t)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "both credentials and certificate present",
			setup: func(t *testing.T) {
				os.Setenv(HCPClientID, "foo")
				os.Setenv(HCPClientSecret, "bar")
				os.Setenv(HCPCredFile, tmpCredFile(t))
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "certificate file check returns error",
			setup: func(t *testing.T) {
				os.Unsetenv(HCPClientID)
				os.Unsetenv(HCPClientSecret)
				os.Setenv(HCPCredFile, "/dev/null/doesnotexist")
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			got, err := HasHCPAuth()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}

	// Restore original env vars
	if origClientID != "" {
		os.Setenv(HCPClientID, origClientID)
	} else {
		os.Unsetenv(HCPClientID)
	}
	if origClientSecret != "" {
		os.Setenv(HCPClientSecret, origClientSecret)
	} else {
		os.Unsetenv(HCPClientSecret)
	}
	if origCredFile != "" {
		os.Setenv(HCPCredFile, origCredFile)
	} else {
		os.Unsetenv(HCPCredFile)
	}
	os.Remove(defaultCredPath)
	// Restore original default cred file if it was present before test run
	if origDefaultCredFileExists {
		if err := os.Rename(origDefaultCredFilePath, defaultCredPath); err != nil {
			t.Fatalf("failed to delete temp cred file: %v", err)
		}
	}
}
