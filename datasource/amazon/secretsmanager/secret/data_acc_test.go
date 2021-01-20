package secret

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

func TestAmazonSecretsManagerSecret(t *testing.T) {
	secretName := "packer_datasource_secret_test_secret"
	secretDescription := "this is a secret used in a packer acc test"
	secret := new(secretsmanager.CreateSecretOutput)

	testCase := &acctest.DatasourceTestCase{
		Name: "amazon_secretsmanager-secret_datasource_basic_test",
		Setup: func() error {
			accessConfig := &awscommon.AccessConfig{}
			session, err := accessConfig.Session()
			if err != nil {
				return fmt.Errorf("Unable to create aws session %s", err.Error())
			}

			api := secretsmanager.New(session)
			newSecret := &secretsmanager.CreateSecretInput{
				Description:  aws.String(secretDescription),
				Name:         aws.String(secretName),
				SecretString: aws.String("{packer_test_key:this_is_the_packer_test_secret_value}"),
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

			arnLog := fmt.Sprintf("null.basic-example: secret arn: %s", aws.StringValue(secret.ARN))
			idLog := fmt.Sprintf("null.basic-example: secret id: %s", aws.StringValue(secret.ARN))
			nameLog := fmt.Sprintf("null.basic-example: secret name: %s", aws.StringValue(secret.Name))
			descriptionLog := fmt.Sprintf("null.basic-example: secret description: %s", secretDescription)

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
