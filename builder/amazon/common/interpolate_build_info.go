package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
)

type BuildInfoTemplate struct {
	BuildRegion   string
	SourceAMI     string
	SourceAMIName string
	SourceAMITags map[string]string
}

func extractBuildInfo(region string, state multistep.StateBag) *BuildInfoTemplate {
	rawSourceAMI, hasSourceAMI := state.GetOk("source_image")
	if !hasSourceAMI {
		return &BuildInfoTemplate{
			BuildRegion: region,
		}
	}

	sourceAMI := rawSourceAMI.(*ec2.Image)
	sourceAMITags := make(map[string]string, len(sourceAMI.Tags))
	for _, tag := range sourceAMI.Tags {
		sourceAMITags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	return &BuildInfoTemplate{
		BuildRegion:   region,
		SourceAMI:     aws.StringValue(sourceAMI.ImageId),
		SourceAMIName: aws.StringValue(sourceAMI.Name),
		SourceAMITags: sourceAMITags,
	}
}
