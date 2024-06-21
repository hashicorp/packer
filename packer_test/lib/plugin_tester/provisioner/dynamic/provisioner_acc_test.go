// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template.pkr.hcl
var testProvisionerHCL2Basic string

// Run with: PACKER_ACC=1 go test -count 1 -v ./provisioner/scaffolding/provisioner_acc_test.go  -timeout=120m
func TestAccScaffoldingProvisioner(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "scaffolding_provisioner_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testProvisionerHCL2Basic,
		Type:     "scaffolding-my-provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			provisionerOutputLog := "null.basic-example: provisioner mock: my-mock-config"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected foo value %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
