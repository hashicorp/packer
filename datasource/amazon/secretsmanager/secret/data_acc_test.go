package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer/datasource/amazon/secretsmanager/acceptance"
)

func TestAmazonSecretsManagerSecret(t *testing.T) {
	secret := &acceptance.AmazonSecret{
		Name:        "packer_datasource_secret_test_secret",
		Key:         "packer_test_key",
		Value:       "this_is_the_packer_test_secret_value",
		Description: "this is a secret used in a packer acc test",
	}

	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_secretsmanager-secret_datasource_basic_test",
		Setup: func() error {
			return secret.Create()
		},
		Teardown: func() error {
			return secret.Delete()
		},
		Template: testDatasourceBasic,
		Type:     "amazon-secrestmanager-secret",
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

			arnLog := fmt.Sprintf("null.basic-example: secret arn: %s", aws.StringValue(secret.Info.ARN))
			idLog := fmt.Sprintf("null.basic-example: secret id: %s", aws.StringValue(secret.Info.ARN))
			nameLog := fmt.Sprintf("null.basic-example: secret name: %s", aws.StringValue(secret.Info.Name))
			descriptionLog := fmt.Sprintf("null.basic-example: secret description: %s", secret.Description)

			if matched, _ := regexp.MatchString(arnLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected arn %q", logsString)
			}
			if matched, _ := regexp.MatchString(idLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected id %q", logsString)
			}
			if matched, _ := regexp.MatchString(nameLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected name %q", logsString)
			}
			if matched, _ := regexp.MatchString(descriptionLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected description %q", logsString)
			}
			return nil
		},
	}
	acctest.TestDatasource(t, testCase)
}

const testDatasourceBasic = `
data "amazon-secretsmanager-secret" "test" {
    name = "packer_datasource_secret_test_secret"
}

locals { 
	arn = data.amazon-secretsmanager-secret.test.arn
	id = data.amazon-secretsmanager-secret.test.id
	name = data.amazon-secretsmanager-secret.test.name
	description = data.amazon-secretsmanager-secret.test.description
}

source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "shell-local" {
    inline  = [
		"echo secret arn: ${local.arn}",
		"echo secret id: ${local.id}",
		"echo secret name: ${local.name}",
		"echo secret description: ${local.description}",
	]
  }
}
`
