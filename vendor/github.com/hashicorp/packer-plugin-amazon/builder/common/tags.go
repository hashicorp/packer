package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type TagMap map[string]string
type EC2Tags []*ec2.Tag

func (t EC2Tags) Report(ui packersdk.Ui) {
	for _, tag := range t {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"",
			aws.StringValue(tag.Key), aws.StringValue(tag.Value)))
	}
}

func (t TagMap) EC2Tags(ictx interpolate.Context, region string, state multistep.StateBag) (EC2Tags, error) {
	var ec2Tags []*ec2.Tag
	generatedData := packerbuilderdata.GeneratedData{State: state}
	ictx.Data = extractBuildInfo(region, state, &generatedData)

	for key, value := range t {
		interpolatedKey, err := interpolate.Render(key, &ictx)
		if err != nil {
			return nil, fmt.Errorf("Error processing tag: %s:%s - %s", key, value, err)
		}
		interpolatedValue, err := interpolate.Render(value, &ictx)
		if err != nil {
			return nil, fmt.Errorf("Error processing tag: %s:%s - %s", key, value, err)
		}
		ec2Tags = append(ec2Tags, &ec2.Tag{
			Key:   aws.String(interpolatedKey),
			Value: aws.String(interpolatedValue),
		})
	}
	return ec2Tags, nil
}
