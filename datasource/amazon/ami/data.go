//go:generate mapstructure-to-hcl2 -type DatasourceOutput,Config
package ami

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
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
	if errs := packersdk.MultiErrorAppend(errs, d.config.AccessConfig.Prepare()...); len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type DatasourceOutput struct {
	ID           string            `mapstructure:"id"`
	Name         string            `mapstructure:"name"`
	CreationDate string            `mapstructure:"creation_date"`
	Owner        string            `mapstructure:"owner"`
	OwnerName    string            `mapstructure:"owner_name"`
	Tags         map[string]string `mapstructure:"tags"`
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
	return hcl2.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
