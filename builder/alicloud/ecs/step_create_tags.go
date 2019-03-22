package ecs

import (
	"context"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateTags struct {
	Tags map[string]string
}

func (s *stepCreateTags) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("alicloudimage").(string)
	snapshotIds := state.Get("alicloudsnapshots").([]string)

	if len(s.Tags) == 0 {
		return multistep.ActionContinue
	}
	ui.Say(fmt.Sprintf("Adding tags(%s) to image: %s", s.Tags, imageId))
	err := client.AddTags(&ecs.AddTagsArgs{
		ResourceId:   imageId,
		ResourceType: ecs.TagResourceImage,
		RegionId:     common.Region(config.AlicloudRegion),
		Tag:          s.Tags,
	})
	if err != nil {
		return halt(state, err, "Error Adding tags to image")
	}

	for _, snapshotId := range snapshotIds {
		ui.Say(fmt.Sprintf("Adding tags(%s) to snapshot: %s", s.Tags, snapshotId))
		err = client.AddTags(&ecs.AddTagsArgs{
			ResourceId:   snapshotId,
			ResourceType: ecs.TagResourceSnapshot,
			RegionId:     common.Region(config.AlicloudRegion),
			Tag:          s.Tags,
		})
		if err != nil {
			return halt(state, err, "Error Adding tags to snapshot")
		}
	}

	return multistep.ActionContinue
}
func (s *stepCreateTags) Cleanup(state multistep.StateBag) {
	// Nothing need to do, tags will be cleaned when the resource is cleaned
}
