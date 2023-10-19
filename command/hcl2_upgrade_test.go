// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_hcl2_upgrade(t *testing.T) {

	tc := []struct {
		folder    string
		flags     []string
		exitCode  int
		exitEarly bool
	}{
		{folder: "unknown_builder", flags: []string{}, exitCode: 1}, // warn for unknown components not tracked in knownPluginPrefixes
		{folder: "complete", flags: []string{"-with-annotations"}, exitCode: 0},
		{folder: "without-annotations", flags: []string{}, exitCode: 0},
		{folder: "minimal", flags: []string{"-with-annotations"}, exitCode: 0},
		{folder: "source-name", flags: []string{"-with-annotations"}, exitCode: 0},
		{folder: "error-cleanup-provisioner", flags: []string{"-with-annotations"}, exitCode: 0},
		{folder: "aws-access-config", flags: []string{}, exitCode: 0},
		{folder: "escaping", flags: []string{}, exitCode: 0},
		{folder: "vsphere_linux_options_and_network_interface", flags: []string{}, exitCode: 0}, //do not warn for known uninstalled plugins components
		{folder: "nonexistent", flags: []string{}, exitCode: 1, exitEarly: true},
		{folder: "placeholders", flags: []string{}, exitCode: 0},
		{folder: "ami_test", flags: []string{}, exitCode: 0},
		{folder: "azure_shg", flags: []string{}, exitCode: 0},
		{folder: "variables-only", flags: []string{}, exitCode: 0},
		{folder: "variables-with-variables", flags: []string{}, exitCode: 0},
		{folder: "complete-variables-with-template-engine", flags: []string{}, exitCode: 0},
		{folder: "undeclared-variables", flags: []string{}, exitCode: 0},
		{folder: "varfile-with-no-variables-block", flags: []string{}, exitCode: 0},
		{folder: "bundled-plugin-used", flags: []string{}, exitCode: 0},
	}

	for _, tc := range tc {
		t.Run(tc.folder, func(t *testing.T) {
			inputPath := filepath.Join(testFixture("hcl2_upgrade", tc.folder, "input.json"))
			outputPath := inputPath + ".pkr.hcl"
			expectedPath := filepath.Join(testFixture("hcl2_upgrade", tc.folder, "expected.pkr.hcl"))
			args := []string{"hcl2_upgrade"}
			if len(tc.flags) > 0 {
				args = append(args, tc.flags...)
			}
			args = append(args, inputPath)
			p := helperCommand(t, args...)
			err := p.Run()
			defer os.Remove(outputPath)
			if err != nil {
				t.Logf("run returned an error: %s", err)
			}
			actualExitCode := p.ProcessState.ExitCode()
			if tc.exitCode != actualExitCode {
				t.Fatalf("unexpected exit code: %d found; expected %d ", actualExitCode, tc.exitCode)
			}
			if tc.exitEarly {
				return
			}
			expected := string(mustBytes(os.ReadFile(expectedPath)))
			actual := string(mustBytes(os.ReadFile(outputPath)))

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Fatalf("unexpected output: %s", diff)
			}
		})
	}
}

func mustBytes(b []byte, e error) []byte {
	if e != nil {
		panic(e)
	}
	return b
}
