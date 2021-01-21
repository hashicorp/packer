package secret_version

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

func TestAmazonSecretsManagerSecretVersion(t *testing.T) {
	secret := &acceptance.AmazonSecret{
		Name:        "packer_datasource_secret_version_test_secret",
		Key:         "packer_test_key",
		Value:       "this_is_the_packer_test_secret_value",
		Description: "this is a secret used in a packer acc test",
	}

	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_secretsmanager-secret-version_datasource_basic_test",
		Setup: func() error {
			return secret.Create()
		},
		Teardown: func() error {
			return secret.Delete()
		},
		Template: testDatasourceBasic,
		Type:     "amazon-secrestmanager-secret-version",
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
			secretStringLog := fmt.Sprintf("null.basic-example: secret secret_string: %s", fmt.Sprintf("{%s:%s}", secret.Key, secret.Value))
			versionIdLog := fmt.Sprintf("null.basic-example: secret version_id: %s", aws.StringValue(secret.Info.VersionId))
			secretValueLog := fmt.Sprintf("null.basic-example: secret value: %s", secret.Value)

			if matched, _ := regexp.MatchString(arnLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected arn %q", logsString)
			}
			if matched, _ := regexp.MatchString(secretStringLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected secret_string %q", logsString)
			}
			if matched, _ := regexp.MatchString(versionIdLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected version_id %q", logsString)
			}
			if matched, _ := regexp.MatchString(secretValueLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected value %q", logsString)
			}
			return nil
		},
	}
	acctest.TestDatasource(t, testCase)
}

const testDatasourceBasic = `
data "amazon-secretsmanager-secret-version" "test" {
  secret_id = "packer_datasource_secret_version_test_secret"
}

locals {
  arn           = data.amazon-secretsmanager-secret-version.test.arn
  secret_string = data.amazon-secretsmanager-secret-version.test.secret_string
  version_id    = data.amazon-secretsmanager-secret-version.test.version_id
  secret_value  = jsondecode(data.amazon-secretsmanager-secret-version.test.secret_string)["packer_test_key"]
}

source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo secret arn: ${local.arn}",
      "echo secret secret_string: ${local.secret_string}",
      "echo secret version_id: ${local.version_id}",
 	  "echo secret value: ${local.secret_value}"
    ]
  }
}
`
