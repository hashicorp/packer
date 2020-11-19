package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreateTags struct {
	Tags map[string]string
}

func (s *stepCreateTags) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)
	imageId := state.Get("alicloudimage").(string)
	snapshotIds := state.Get("alicloudsnapshots").([]string)

	if len(s.Tags) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Adding tags(%s) to image: %s", s.Tags, imageId))

	var tags []ecs.AddTagsTag
	for key, value := range s.Tags {
		var tag ecs.AddTagsTag
		tag.Key = key
		tag.Value = value
		tags = append(tags, tag)
	}

	addTagsRequest := ecs.CreateAddTagsRequest()
	addTagsRequest.RegionId = config.AlicloudRegion
	addTagsRequest.ResourceId = imageId
	addTagsRequest.ResourceType = TagResourceImage
	addTagsRequest.Tag = &tags

	if _, err := client.AddTags(addTagsRequest); err != nil {
		return halt(state, err, "Error Adding tags to image")
	}

	for _, snapshotId := range snapshotIds {
		ui.Say(fmt.Sprintf("Adding tags(%s) to snapshot: %s", s.Tags, snapshotId))
		addTagsRequest := ecs.CreateAddTagsRequest()

		addTagsRequest.RegionId = config.AlicloudRegion
		addTagsRequest.ResourceId = snapshotId
		addTagsRequest.ResourceType = TagResourceSnapshot
		addTagsRequest.Tag = &tags

		if _, err := client.AddTags(addTagsRequest); err != nil {
			return halt(state, err, "Error Adding tags to snapshot")
		}
	}

	return multistep.ActionContinue
}
func (s *stepCreateTags) Cleanup(state multistep.StateBag) {
	// Nothing need to do, tags will be cleaned when the resource is cleaned
}
