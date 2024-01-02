// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcp_packer_iteration

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer/internal/hcp/env"
)

//go:embed test-fixtures/template.pkr.hcl
var testDatasourceBasic string

// Acceptance tests for data sources.
//
// To be successful, the HCP project you're providing credentials for must
// contain a bucket named "hardened-ubuntu-16-04", with a channel named
// "packer-acc-test". It must contain a build that references an image in AWS
// region "us-east-1". Your HCP credentials must be provided through your
// runtime environment because the template this test uses does not set them.
//
// TODO: update this acceptance to create and clean up the HCP resources this
// data source queries, to prevent plugin developers from having to have images
// as defined above.

func TestAccDatasource_HCPPackerIteration(t *testing.T) {
	if os.Getenv(env.HCPClientID) == "" && os.Getenv(env.HCPClientSecret) == "" {
		t.Skipf(fmt.Sprintf("Acceptance tests skipped unless envs %q and %q are set", env.HCPClientID, env.HCPClientSecret))
		return
	}

	testCase := &acctest.PluginTestCase{
		Name:     "hcp_packer_iteration_datasource_basic_test",
		Template: testDatasourceBasic,
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
	acctest.TestPlugin(t, testCase)
}
