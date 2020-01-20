package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/builder"
	"github.com/hashicorp/packer/helper/multistep"
)

type BuildInfoTemplate struct {
	BuildRegion        string
	SourceAMI          string
	SourceAMIName      string
	SourceAMIOwner     string
	SourceAMIOwnerName string
	SourceAMITags      map[string]string
}

func extractBuildInfo(region string, state multistep.StateBag, generatedData *builder.GeneratedData) *BuildInfoTemplate {
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

	buildInfoTemplate := &BuildInfoTemplate{
		BuildRegion:        region,
		SourceAMI:          aws.StringValue(sourceAMI.ImageId),
		SourceAMIName:      aws.StringValue(sourceAMI.Name),
		SourceAMIOwner:     aws.StringValue(sourceAMI.OwnerId),
		SourceAMIOwnerName: aws.StringValue(sourceAMI.ImageOwnerAlias),
		SourceAMITags:      sourceAMITags,
	}
	generatedData.Put("SourceAMIName", buildInfoTemplate.SourceAMIName)
	return buildInfoTemplate
}
