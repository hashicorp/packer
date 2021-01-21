package acceptance

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/builder/amazon/common/awserrors"
)

type AmazonSecret struct {
	Name        string
	Key         string
	Value       string
	Description string

	Info    *secretsmanager.CreateSecretOutput
	manager *secretsmanager.SecretsManager
}

func (as *AmazonSecret) Create() error {
	if as.manager == nil {
		accessConfig := &awscommon.AccessConfig{}
		session, err := accessConfig.Session()
		if err != nil {
			return fmt.Errorf("Unable to create aws session %s", err.Error())
		}
		as.manager = secretsmanager.New(session)
	}

	newSecret := &secretsmanager.CreateSecretInput{
		Description:  aws.String(as.Description),
		Name:         aws.String(as.Name),
		SecretString: aws.String(fmt.Sprintf(`{%q:%q}`, as.Key, as.Value)),
	}

	secret := new(secretsmanager.CreateSecretOutput)
	var err error
	err = retry.Config{
		Tries: 11,
		ShouldRetry: func(err error) bool {
			if awserrors.Matches(err, "ResourceExistsException", "") {
				_ = as.Delete()
				return true
			}
			if awserrors.Matches(err, "InvalidRequestException", "already scheduled for deletion") {
				return true
			}
			return false
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(context.TODO(), func(_ context.Context) error {
		secret, err = as.manager.CreateSecret(newSecret)
		return err
	})
	as.Info = secret
	return err
}

func (as *AmazonSecret) Delete() error {
	if as.manager == nil {
		accessConfig := &awscommon.AccessConfig{}
		session, err := accessConfig.Session()
		if err != nil {
			return fmt.Errorf("Unable to create aws session %s", err.Error())
		}
		as.manager = secretsmanager.New(session)
	}

	secret := &secretsmanager.DeleteSecretInput{
		ForceDeleteWithoutRecovery: aws.Bool(true),
		SecretId:                   aws.String(as.Name),
	}
	_, err := as.manager.DeleteSecret(secret)
	return err
}
