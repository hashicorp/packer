// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest/provisioneracc"
	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
)

func fixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "test-fixtures")
}

func loadFile(templateFragmentPath string) (string, error) {
	dir := fixtureDir()
	fragmentAbsPath := filepath.Join(dir, templateFragmentPath)
	fragmentFile, err := os.Open(fragmentAbsPath)
	if err != nil {
		return "", fmt.Errorf("Unable find %s", fragmentAbsPath)
	}
	defer fragmentFile.Close()

	fragmentString, err := io.ReadAll(fragmentFile)
	if err != nil {
		return "", fmt.Errorf("Unable to read %s", fragmentAbsPath)
	}

	return string(fragmentString), nil
}

func IsCompatible(builder string, vmOS string) bool {
	return vmOS == "linux"
}

func TestAccShellProvisioner_basic(t *testing.T) {
	templateString, err := loadFile("shell-local-provisioner.txt")
	if err != nil {
		t.Fatalf("Couldn't load test fixture; %s", err.Error())
	}

	testCase := &provisioneracc.ProvisionerTestCase{
		IsCompatible: IsCompatible,
		Name:         "shell-local-provisioner-basic",
		Teardown: func() error {
			testutils.CleanupFiles("test-fixtures/file.txt")
			return nil
		},
		Template: templateString,
		Type:     "shell-local",
		Check: func(buildcommand *exec.Cmd, logfile string) error {
			if buildcommand.ProcessState != nil {
				if buildcommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s\n", logfile)
				}
			}
			filecontents, err := loadFile("file.txt")
			if err != nil {
				return err
			}
			if !strings.Contains(filecontents, "hello") {
				return fmt.Errorf("file contents were wrong: %s", filecontents)
			}
			return nil
		},
	}

	provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}
