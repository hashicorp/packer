//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type DatasourceOutput,Config
package secretsmanager

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/hcl/v2/hcldec"
	awscommon "github.com/hashicorp/packer-plugin-amazon/builder/common"
	"github.com/hashicorp/packer-plugin-amazon/builder/common/awserrors"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Datasource struct {
	config Config
}

type Config struct {
	// Specifies the secret containing the version that you want to retrieve.
	// You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret.
	Name string `mapstructure:"name" required:"true"`
	// Optional key for JSON secrets that contain more than one value. When set, the `value` output will
	// contain the value for the provided key.
	Key string `mapstructure:"key"`
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

	if d.config.Name == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a 'name' must be provided"))
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
	// When a [key](#key) is provided, this will be the value for that key. If a key is not provided,
	// `value` will contain the first value found in the secret string.
	Value string `mapstructure:"value"`
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
		SecretId: aws.String(d.config.Name),
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
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q Version %q not found", d.config.Name, version)
		}
		if awserrors.Matches(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q Version %q not found", d.config.Name, version)
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error reading Secrets Manager Secret Version: %s", err)
	}

	value, err := getSecretValue(aws.StringValue(secret.SecretString), d.config.Key)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error to get secret value: %q", err.Error())
	}

	versionId := aws.StringValue(secret.VersionId)
	output := DatasourceOutput{
		Value:        value,
		SecretString: aws.StringValue(secret.SecretString),
		SecretBinary: fmt.Sprintf("%s", secret.SecretBinary),
		VersionId:    versionId,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func getSecretValue(secretString string, key string) (string, error) {
	var secretValue map[string]interface{}
	blob := []byte(secretString)

	//For those plaintext secrets just return the value
	if json.Valid(blob) != true {
		return secretString, nil
	}

	err := json.Unmarshal(blob, &secretValue)
	if err != nil {
		return "", err
	}

	if key == "" {
		for _, v := range secretValue {
			return getStringSecretValue(v)
		}
	}

	if v, ok := secretValue[key]; ok {
		return getStringSecretValue(v)
	}

	return "", nil
}

func getStringSecretValue(v interface{}) (string, error) {
	switch valueType := v.(type) {
	case string:
		return valueType, nil
	case float64:
		return strconv.FormatFloat(valueType, 'f', 0, 64), nil
	default:
		return "", fmt.Errorf("Unsupported secret value type: %T", valueType)
	}
}
