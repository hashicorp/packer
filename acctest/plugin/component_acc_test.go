// component_acc_test.go should contain acceptance tests for plugin components
// to make sure all component types can be discovered and started.
package plugin

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	amazonacc "github.com/hashicorp/packer-plugin-amazon/builder/ebs/acceptance"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

//go:embed test-fixtures/basic-amazon-ami-datasource.pkr.hcl
var basicAmazonAmiDatasourceHCL2Template string

func TestAccInitAndBuildBasicAmazonAmiDatasource(t *testing.T) {
	plugin := addrs.Plugin{
		Hostname:  "github.com",
		Namespace: "hashicorp",
		Type:      "amazon",
	}
	testCase := &acctest.PluginTestCase{
		Name: "amazon-ami_basic_datasource_test",
		Setup: func() error {
			return cleanupPluginInstallation(plugin)
		},
		Teardown: func() error {
			helper := amazonacc.AMIHelper{
				Region: "us-west-2",
				Name:   "packer-amazon-ami-test",
			}
			return helper.CleanUpAmi()
		},
		Template: basicAmazonAmiDatasourceHCL2Template,
		Type:     "amazon-ami",
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

			logsBytes, err := ioutil.ReadAll(logs)
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
