package ami

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	amazonacc "github.com/hashicorp/packer/builder/amazon/ebs/acceptance"
)

func TestAmazonAmi(t *testing.T) {
	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_ami_datasource_basic_test",
		Teardown: func() error {
			helper := amazonacc.AWSHelper{
				Region:  "us-west-2",
				AMIName: "packer-amazon-ami-test",
			}
			return helper.CleanUpAmi()
		},
		Template: testDatasourceBasic,
		Type:     "amazon-ami",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestDatasource(t, testCase)
}

const testDatasourceBasic = `
data "amazon-ami" "test" {
  filters = {
    virtualization-type = "hvm"
    name                = "Windows_Server-2016-English-Full-Base-*"
    root-device-type    = "ebs"
  }
  most_recent = true
  owners = ["801119661308"]
}

source "amazon-ebs" "basic-example" {
  user_data_file = "./test-fixtures/configure-source-ssh.ps1"
  region = "us-west-2"
  source_ami = data.amazon-ami.test.id
  instance_type =  "t2.small"
  ssh_agent_auth = false
  ami_name =  "packer-amazon-ami-test"
  communicator = "ssh"
  ssh_timeout = "10m"
  ssh_username = "Administrator"
}

build {
  sources = [
    "source.amazon-ebs.basic-example"
  ]
}
`
