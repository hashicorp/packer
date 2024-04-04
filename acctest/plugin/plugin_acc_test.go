// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// plugin_acc_test.go should contain acceptance tests for features related to
// installing, discovering and running plugins.
package plugin

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/hashicorp/packer/packer"
)

//go:embed test-fixtures/basic-amazon-ebs.pkr.hcl
var basicAmazonEbsHCL2Template string

func TestAccInitAndBuildBasicAmazonEbs(t *testing.T) {
	plugin := addrs.Plugin{
		Source: "github.com/hashicorp/amazon",
	}
	testCase := &acctest.PluginTestCase{
		Name: "amazon-ebs_basic_plugin_init_and_build_test",
		Setup: func() error {
			return cleanupPluginInstallation(plugin)
		},
		Template: basicAmazonEbsHCL2Template,
		Type:     "amazon-ebs",
		Init:     true,
		CheckInit: func(initCommand *exec.Cmd, logfile string) error {
			if initCommand.ProcessState != nil {
				if initCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := io.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			initOutput := string(logsBytes)
			return checkPluginInstallation(initOutput, plugin)
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func pluginDirectory(plugin addrs.Plugin) (string, error) {
	pluginDir, err := packer.PluginFolder()
	if err != nil {
		return "", err
	}

	pluginParts := []string{pluginDir}
	pluginParts = append(pluginParts, plugin.Parts()...)
	return filepath.Join(pluginParts...), nil
}

func cleanupPluginInstallation(plugin addrs.Plugin) error {
	pluginPath, err := pluginDirectory(plugin)
	if err != nil {
		return err
	}
	testutils.CleanupFiles(pluginPath)
	return nil
}

func checkPluginInstallation(initOutput string, plugin addrs.Plugin) error {
	expectedInitLog := "Installed plugin " + plugin.String()
	if matched, _ := regexp.MatchString(expectedInitLog+".*", initOutput); !matched {
		return fmt.Errorf("logs doesn't contain expected foo value %q", initOutput)
	}

	pluginPath, err := pluginDirectory(plugin)
	if err != nil {
		return err
	}

	if !testutils.FileExists(pluginPath) {
		return fmt.Errorf("%s plugin installation not found", plugin.String())
	}
	return nil
}
