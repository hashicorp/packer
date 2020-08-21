package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
	"github.com/outscale/osc-sdk-go/osc"
)

type TagMap map[string]string
type OAPITags []oapi.ResourceTag
type OSCTags []osc.ResourceTag

func (t OAPITags) Report(ui packer.Ui) {
	for _, tag := range t {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"",
			tag.Key, tag.Value))
	}
}

func (t OSCTags) Report(ui packer.Ui) {
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

func CreateTags(conn *oapi.Client, resourceID string, ui packer.Ui, tags OAPITags) error {
	tags.Report(ui)

	_, err := conn.POST_CreateTags(oapi.CreateTagsRequest{
		ResourceIds: []string{resourceID},
		Tags:        tags,
	})

	return err
}

func CreateOSCTags(conn *osc.APIClient, resourceID string, ui packer.Ui, tags OSCTags) error {
	tags.Report(ui)

	_, _, err := conn.TagApi.CreateTags(context.Background(), &osc.CreateTagsOpts{
		CreateTagsRequest: optional.NewInterface(osc.CreateTagsRequest{
			ResourceIds: []string{resourceID},
			Tags:        tags,
		}),
	})

	return err
}
