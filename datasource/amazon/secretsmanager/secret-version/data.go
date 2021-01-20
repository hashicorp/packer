//go:generate struct-markdown
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
	// Specifies the secret containing the version that you want to retrieve.
	// You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret.
	SecretId string `mapstructure:"secret_id" required:"true"`
	// Specifies the unique identifier of the version of the secret that you want to retrieve.
	// Overrides version_stage.
	VersionId string `mapstructure:"version_id"`
	// Specifies the secret version that you want to retrieve by the staging label attached to the version.
	// Defaults to AWSCURRENT.
	VersionStage           string `mapstructure:"version_stage"`
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

	if d.config.VersionStage == "" {
		d.config.VersionStage = "AWSCURRENT"
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	// The Amazon Resource Name (ARN) of the secret.
	Arn string `mapstructure:"arn"`
	// The decrypted part of the protected secret information that
	// was originally provided as a string.
	SecretString string `mapstructure:"secret_string"`
	// The decrypted part of the protected secret information that
	// was originally provided as a binary. Base64 encoded.
	SecretBinary string `mapstructure:"secret_binary"`
	// The unique identifier of this version of the secret.
	VersionId string `mapstructure:"version_id"`
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
	} else {
		input.VersionStage = aws.String(d.config.VersionStage)
		version = d.config.VersionStage
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

	versionId := aws.StringValue(secret.VersionId)
	output := DatasourceOutput{
		Arn:          aws.StringValue(secret.ARN),
		SecretString: aws.StringValue(secret.SecretString),
		SecretBinary: fmt.Sprintf("%s", secret.SecretBinary),
		VersionId:    versionId,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
