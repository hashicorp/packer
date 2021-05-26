package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_hcl2_upgrade(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	_ = cwd

	tc := []struct {
		folder   string
		flags    []string
		exitCode int
	}{
		{folder: "unknown_builder", flags: []string{}, exitCode: 1},
		{folder: "complete", flags: []string{"-with-annotations"}},
		{folder: "without-annotations", flags: []string{}},
		{folder: "minimal", flags: []string{"-with-annotations"}},
		{folder: "source-name", flags: []string{"-with-annotations"}},
		{folder: "error-cleanup-provisioner", flags: []string{"-with-annotations"}},
		{folder: "aws-access-config", flags: []string{}},
		{folder: "variables-only", flags: []string{}},
		{folder: "variables-with-variables", flags: []string{}},
		{folder: "complete-variables-with-template-engine", flags: []string{}},
		{folder: "escaping", flags: []string{}},
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
			if err != nil {
				t.Logf("run returned an error: %s", err)
			}
			expected := string(mustBytes(ioutil.ReadFile(expectedPath)))
			actual := string(mustBytes(ioutil.ReadFile(outputPath)))

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Fatalf("unexpected output: %s", diff)
			}
			actualExitCode := p.ProcessState.ExitCode()
			if tc.exitCode != actualExitCode {
				t.Fatalf("unexpected exit code: %d found; expected %d ", actualExitCode, tc.exitCode)
			}
			os.Remove(outputPath)
		})
	}
}

func mustBytes(b []byte, e error) []byte {
	if e != nil {
		panic(e)
	}
	return b
}
