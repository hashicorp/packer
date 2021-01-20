package secret_version

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/builder/amazon/common/awserrors"
)

func TestAmazonSecretsManagerSecretVersion(t *testing.T) {
	secretName := "packer_datasource_secret_version_test_secret"
	secretKey := "packer_test_key"
	secretValue := "this_is_the_packer_test_secret_value"
	secretString := fmt.Sprintf(`{%q:%q}`, secretKey, secretValue)
	secret := new(secretsmanager.CreateSecretOutput)

	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_secretsmanager-secret-version_datasource_basic_test",
		Setup: func() error {
			// Create a secret
			accessConfig := &awscommon.AccessConfig{}
			session, err := accessConfig.Session()
			if err != nil {
				return fmt.Errorf("Unable to create aws session %s", err.Error())
			}

			api := secretsmanager.New(session)
			newSecret := &secretsmanager.CreateSecretInput{
				Description:  aws.String("this is a secret used in a packer acc test"),
				Name:         aws.String(secretName),
				SecretString: aws.String(secretString),
			}

			err = retry.Config{
				Tries: 11,
				ShouldRetry: func(error) bool {
					if awserrors.Matches(err, "ResourceExistsException", "") {
						oldSecret := &secretsmanager.DeleteSecretInput{
							ForceDeleteWithoutRecovery: aws.Bool(true),
							SecretId:                   aws.String(secretName),
						}
						_, _ = api.DeleteSecret(oldSecret)
						return true
					}
					if awserrors.Matches(err, "InvalidRequestException", "already scheduled for deletion") {
						return true
					}
					return false
				},
				RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
			}.Run(context.TODO(), func(_ context.Context) error {
				secret, err = api.CreateSecret(newSecret)
				return err
			})
			return err
		},
		Teardown: func() error {
			// Remove the created secret
			accessConfig := &awscommon.AccessConfig{}
			session, err := accessConfig.Session()
			if err != nil {
				return fmt.Errorf("Unable to create aws session %s", err.Error())
			}

			api := secretsmanager.New(session)
			secret := &secretsmanager.DeleteSecretInput{
				ForceDeleteWithoutRecovery: aws.Bool(true),
				SecretId:                   aws.String(secretName),
			}
			_, err = api.DeleteSecret(secret)
			return err
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

			arnLog := fmt.Sprintf("null.basic-example: secret arn: %s", aws.StringValue(secret.ARN))
			idLog := fmt.Sprintf("null.basic-example: secret id: %s", secretName)
			secretStringLog := fmt.Sprintf("null.basic-example: secret secret_string: %s", fmt.Sprintf("{%s:%s}", secretKey, secretValue))
			versionIdLog := fmt.Sprintf("null.basic-example: secret version_id: %s", aws.StringValue(secret.VersionId))
			secretValueLog := fmt.Sprintf("null.basic-example: secret value: %s", secretValue)

			if matched, _ := regexp.MatchString(arnLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected arn %q", logsString)
			}
			if matched, _ := regexp.MatchString(idLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected id %q", logsString)
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
  id            = data.amazon-secretsmanager-secret-version.test.id
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
      "echo secret id: ${local.id}",
      "echo secret secret_string: ${local.secret_string}",
      "echo secret version_id: ${local.version_id}",
 	  "echo secret value: ${local.secret_value}"
    ]
  }
}
`
