// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package powershell_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest/provisioneracc"
)

const TestProvisionerType = "powershell"

func powershellIsCompatible(builder string, vmOS string) bool {
	return vmOS == "windows"
}

func fixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "test-fixtures")
}

func LoadProvisionerFragment(templateFragmentPath string) (string, error) {
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

func TestAccPowershellProvisioner_basic(t *testing.T) {
	templateString, err := LoadProvisionerFragment("powershell-provisioner-cleanup.txt")
	if err != nil {
		t.Fatalf("Couldn't load test fixture; %s", err.Error())
	}
	testCase := &provisioneracc.ProvisionerTestCase{
		IsCompatible: powershellIsCompatible,
		Name:         "powershell-provisioner-cleanup",
		Template:     templateString,
		Type:         TestProvisionerType,
		Check: func(buildcommand *exec.Cmd, logfile string) error {
			if buildcommand.ProcessState != nil {
				if buildcommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}

	provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}

func TestAccPowershellProvisioner_Inline(t *testing.T) {
	templateString, err := LoadProvisionerFragment("powershell-inline-provisioner.txt")
	if err != nil {
		t.Fatalf("Couldn't load test fixture; %s", err.Error())
	}
	testCase := &provisioneracc.ProvisionerTestCase{
		IsCompatible: powershellIsCompatible,
		Name:         "powershell-provisioner-inline",
		Template:     templateString,
		Type:         TestProvisionerType,
		Check: func(buildcommand *exec.Cmd, logfile string) error {
			if buildcommand.ProcessState != nil {
				if buildcommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}

	provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}

func TestAccPowershellProvisioner_Script(t *testing.T) {
	templateString, err := LoadProvisionerFragment("powershell-script-provisioner.txt")
	if err != nil {
		t.Fatalf("Couldn't load test fixture; %s", err.Error())
	}
	testCase := &provisioneracc.ProvisionerTestCase{
		IsCompatible: powershellIsCompatible,
		Name:         "powershell-provisioner-script",
		Template:     templateString,
		Type:         TestProvisionerType,
		Check: func(buildcommand *exec.Cmd, logfile string) error {
			if buildcommand.ProcessState != nil {
				if buildcommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}

	provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}
