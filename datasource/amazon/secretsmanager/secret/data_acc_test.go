package secret

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestAmazonAmi(t *testing.T) {
	// create secret

	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_secretsmanager-secret_datasource_basic_test",
		Teardown: func() error {
			// remove secret
			return nil
		},
		Template: testDatasourceBasic,
		Type:     "amazon-ami",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			// check if log contains the secret
			return nil
		},
	}
	acctest.TestDatasource(t, testCase)
}

const testDatasourceBasic = `
data "amazon-secretsmanager-secret" "secret" {
    name = "packer_test_secret"
}

data "amazon-secretsmanager-secret-version" "by-version" {
    secret_id = data.amazon-secretsmanager-secret.secret.id
}

locals { password = jsondecode(data.amazon-secretsmanager-secret-version.by-version.secret_string)["packer_test_key"] }

source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "shell-local" {
    inline  = ["echo the password is: ${local.password}"]
  }
}
`
