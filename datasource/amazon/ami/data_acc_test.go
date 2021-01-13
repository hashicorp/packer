package ami

import (
	"fmt"
	"os/exec"
	"testing"

	amazon_acc "github.com/hashicorp/packer/builder/amazon/ebs/acceptance"
	"github.com/hashicorp/packer/datasource/amazon/ami/acceptance"
)

func TestAmazonAmi(t *testing.T) {
	testCase := &acceptance.DatasourceTestCase{
		Name: "amazon_ami_datasource_basic_test",
		Teardown: func() error {
			helper := amazon_acc.AWSHelper{
				Region:  "us-east-1",
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
	acceptance.TestDatasource(t, testCase)
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
  region = "us-east-1"
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
