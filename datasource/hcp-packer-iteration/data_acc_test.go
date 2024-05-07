// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcp_packer_iteration

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer/internal/hcp/env"
)

//go:embed test-fixtures/template.pkr.hcl
var testDatasourceBasic string

//go:embed test-fixtures/hcp-setup-build.pkr.hcl
var testHCPBuild string

// Acceptance tests for data sources.
//
// Your HCP credentials must be provided through your runtime
// environment because the template this test uses does not set them.
func TestAccDatasource_HCPPackerIteration(t *testing.T) {
	if os.Getenv(env.HCPClientID) == "" && os.Getenv(env.HCPClientSecret) == "" {
		t.Skipf(fmt.Sprintf("Acceptance tests skipped unless envs %q and %q are set", env.HCPClientID, env.HCPClientSecret))
		return
	}

	tmpFile := filepath.Join(t.TempDir(), "hcp-target-file")
	testSetup := acctest.PluginTestCase{
		Template: fmt.Sprintf(testHCPBuild, tmpFile),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, &testSetup)

	testCase := acctest.PluginTestCase{
		Name:     "hcp_packer_iteration_datasource_basic_test",
		Template: fmt.Sprintf(testDatasourceBasic, filepath.Dir(tmpFile)),
		Setup: func() error {
			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				return err
			}
			return nil
		},
		// TODO have acc test write iteration id to a file and check it to make
		// sure it isn't empty.
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, &testCase)
}
