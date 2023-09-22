// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	templateString, err := loadFile("shell-provisioner.txt")
	if err != nil {
		t.Fatalf("Couldn't load test fixture; %s", err.Error())
	}

	testCase := &provisioneracc.ProvisionerTestCase{
		IsCompatible: IsCompatible,
		Name:         "shell-provisioner-basic",
		Teardown: func() error {
			testutils.CleanupFiles("test-fixtures/provisioner.shell.txt")
			return nil
		},
		Template: templateString,
		Type:     "shell",
		Check: func(buildcommand *exec.Cmd, logfile string) error {
			if buildcommand.ProcessState != nil {
				if buildcommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			filecontents, err := loadFile("provisioner.shell.txt")
			if err != nil {
				return err
			}
			re := regexp.MustCompile(`build ID is .* and build UUID is [[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}`)
			if !re.MatchString(filecontents) {
				return fmt.Errorf("Bad file contents \"%s\"", filecontents)
			}
			return nil
		},
	}

	provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}
