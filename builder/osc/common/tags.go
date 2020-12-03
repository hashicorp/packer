package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/outscale/osc-sdk-go/osc"
)

type TagMap map[string]string
type OSCTags []osc.ResourceTag

func (t OSCTags) Report(ui packersdk.Ui) {
	for _, tag := range t {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"",
			tag.Key, tag.Value))
	}
}

func (t TagMap) IsSet() bool {
	return len(t) > 0
}

func (t TagMap) OSCTags(ctx interpolate.Context, region string, state multistep.StateBag) (OSCTags, error) {
	var oscTags []osc.ResourceTag
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
		oscTags = append(oscTags, osc.ResourceTag{
			Key:   interpolatedKey,
			Value: interpolatedValue,
		})
	}
	return oscTags, nil
}

func CreateOSCTags(conn *osc.APIClient, resourceID string, ui packersdk.Ui, tags OSCTags) error {
	tags.Report(ui)

	_, _, err := conn.TagApi.CreateTags(context.Background(), &osc.CreateTagsOpts{
		CreateTagsRequest: optional.NewInterface(osc.CreateTagsRequest{
			ResourceIds: []string{resourceID},
			Tags:        tags,
		}),
	})

	return err
}
