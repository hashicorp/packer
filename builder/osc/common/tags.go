package common

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

type TagMap map[string]string
type OAPITags []oapi.ResourceTag

func (t OAPITags) Report(ui packer.Ui) {
	for _, tag := range t {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"",
			tag.Key, tag.Value))
	}
}

func (t TagMap) IsSet() bool {
	return len(t) > 0
}

func (t TagMap) OAPITags(ctx interpolate.Context, region string, state multistep.StateBag) (OAPITags, error) {
	var oapiTags []oapi.ResourceTag
	ctx.Data = extractBuildInfo(region, state)

	for key, value := range t {
		interpolatedKey, err := interpolate.Render(key, &ctx)
		if err != nil {
			return nil, fmt.Errorf("Error processing tag: %s:%s - %s", key, value, err)
		}
		interpolatedValue, err := interpolate.Render(value, &ctx)
		if err != nil {
			return nil, fmt.Errorf("Error processing tag: %s:%s - %s", key, value, err)
		}
		oapiTags = append(oapiTags, oapi.ResourceTag{
			Key:   interpolatedKey,
			Value: interpolatedValue,
		})
	}
	return oapiTags, nil
}

func CreateTags(conn *oapi.Client, resourceID string, ui packer.Ui, tags OAPITags) error {
	tags.Report(ui)

	_, err := conn.POST_CreateTags(oapi.CreateTagsRequest{
		ResourceIds: []string{resourceID},
		Tags:        tags,
	})

	return err
}
