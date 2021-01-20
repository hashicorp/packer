//go:generate mapstructure-to-hcl2 -type DatasourceOutput,Config
package secret_version

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/builder/amazon/common/awserrors"
	"github.com/zclconf/go-cty/cty"
)

type Datasource struct {
	config Config
}

type Config struct {
	SecretId               string `mapstructure:"secret_id" required:"true"`
	VersionId              string `mapstructure:"version_id"`
	VersionState           string `mapstructure:"version_stage"`
	awscommon.AccessConfig `mapstructure:",squash"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, d.config.AccessConfig.Prepare()...)

	if d.config.SecretId == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a 'secret_id' must be provided"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	Arn          string `mapstructure:"arn"`
	Id           string `mapstructure:"id"`
	SecretString string `mapstructure:"secret_string"`
	SecretBinary string `mapstructure:"secret_binary"`
	VersionId    string `mapstructure:"version_id"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	session, err := d.config.Session()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(d.config.SecretId),
	}

	version := ""
	if d.config.VersionId != "" {
		input.VersionId = aws.String(d.config.VersionId)
		version = d.config.VersionId
	} else if d.config.VersionState != "" {
		input.VersionStage = aws.String(d.config.VersionState)
		version = d.config.VersionState
	}

	secretsApi := secretsmanager.New(session)
	secret, err := secretsApi.GetSecretValue(input)
	if err != nil {
		if awserrors.Matches(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q Version %q not found", input.SecretId, version)
		}
		if awserrors.Matches(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q Version %q not found", input.SecretId, version)
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error reading Secrets Manager Secret Version: %s", err)
	}

	output := DatasourceOutput{
		Arn:          aws.StringValue(secret.ARN),
		Id:           d.config.SecretId,
		SecretString: aws.StringValue(secret.SecretString),
		SecretBinary: fmt.Sprintf("%s", secret.SecretBinary),
		VersionId:    aws.StringValue(secret.VersionId),
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
