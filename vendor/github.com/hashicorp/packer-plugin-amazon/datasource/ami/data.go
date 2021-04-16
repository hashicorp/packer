//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type DatasourceOutput,Config
package ami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl/v2/hcldec"
	awscommon "github.com/hashicorp/packer-plugin-amazon/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Datasource struct {
	config Config
}

type Config struct {
	awscommon.AccessConfig     `mapstructure:",squash"`
	awscommon.AmiFilterOptions `mapstructure:",squash"`
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

	if d.config.Empty() {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The `filters` must be specified"))
	}
	if d.config.NoOwner() {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("For security reasons, you must declare an owner."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	// The ID of the AMI.
	ID string `mapstructure:"id"`
	// The name of the AMI.
	Name string `mapstructure:"name"`
	// The date of creation of the AMI.
	CreationDate string `mapstructure:"creation_date"`
	// The AWS account ID of the owner.
	Owner string `mapstructure:"owner"`
	// The owner alias.
	OwnerName string `mapstructure:"owner_name"`
	// The key/value combination of the tags assigned to the AMI.
	Tags map[string]string `mapstructure:"tags"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	session, err := d.config.Session()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	image, err := d.config.AmiFilterOptions.GetFilteredImage(&ec2.DescribeImagesInput{}, ec2.New(session))
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	imageTags := make(map[string]string, len(image.Tags))
	for _, tag := range image.Tags {
		imageTags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	output := DatasourceOutput{
		ID:           aws.StringValue(image.ImageId),
		Name:         aws.StringValue(image.Name),
		CreationDate: aws.StringValue(image.CreationDate),
		Owner:        aws.StringValue(image.OwnerId),
		OwnerName:    aws.StringValue(image.ImageOwnerAlias),
		Tags:         imageTags,
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
