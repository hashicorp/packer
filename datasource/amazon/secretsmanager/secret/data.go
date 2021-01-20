//go:generate mapstructure-to-hcl2 -type DatasourceOutput,Config
package secret

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packerjson "github.com/hashicorp/packer-plugin-sdk/json"
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
	Arn                    string `mapstructure:"arn"`
	Name                   string `mapstructure:"name"`
	awscommon.AccessConfig `mapstructure:",squash"`

	secretId string
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

	if d.config.Arn == "" && d.config.Name == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("provide either a 'name' or an 'arn'"))
	}
	if d.config.Arn != "" && d.config.Name != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("provide only a 'name' or an 'arn'"))
	}

	if d.config.Arn != "" {
		d.config.secretId = d.config.Arn
	} else if d.config.Name != "" {
		d.config.secretId = d.config.Name
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	Arn         string            `mapstructure:"arn"`
	Name        string            `mapstructure:"name"`
	Description string            `mapstructure:"description"`
	KmsKeyId    string            `mapstructure:"kms_key_id"`
	Id          string            `mapstructure:"id"`
	Tags        map[string]string `mapstructure:"tags"`
	Policy      string            `mapstructure:"policy"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	session, err := d.config.Session()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(d.config.secretId),
	}

	secretsApi := secretsmanager.New(session)
	secret, err := secretsApi.DescribeSecret(input)
	if err != nil {
		if awserrors.Matches(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q not found", d.config.secretId)
		}
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error reading Secrets Manager Secret: %s", err)
	}

	if secret.ARN == nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("Secrets Manager Secret %q not found", d.config.secretId)
	}

	output := DatasourceOutput{
		Arn:         aws.StringValue(secret.ARN),
		Name:        aws.StringValue(secret.Name),
		Description: aws.StringValue(secret.Description),
		KmsKeyId:    aws.StringValue(secret.KmsKeyId),
		Id:          aws.StringValue(secret.ARN),
		Tags:        SecretsmanagerTagsMap(secret.Tags),
	}

	policyInput := &secretsmanager.GetResourcePolicyInput{
		SecretId: secret.ARN,
	}
	pOut, err := secretsApi.GetResourcePolicy(policyInput)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("error reading Secrets Manager Secret policy: %s", err)
	}

	if pOut != nil && pOut.ResourcePolicy != nil {
		policy, err := packerjson.NormalizeJsonString(aws.StringValue(pOut.ResourcePolicy))
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		output.Policy = policy
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func SecretsmanagerTagsMap(tags []*secretsmanager.Tag) map[string]string {
	mapOfTags := map[string]string{}

	awsTagKeyPrefix := `aws:`
	for _, tag := range tags {
		k := aws.StringValue(tag.Key)
		if !strings.HasPrefix(k, awsTagKeyPrefix) {
			mapOfTags[k] = aws.StringValue(tag.Value)
		}
	}

	return mapOfTags
}
