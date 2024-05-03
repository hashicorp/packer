package api

import (
	"context"
	"fmt"

	hcpPackerAPI "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// GetChannel loads the named channel that is associated to the bucket name. If the
// channel does not exist in HCP Packer, GetChannel returns an error.
func (c *Client) GetChannel(
	ctx context.Context, bucketName, channelName string,
) (*hcpPackerModels.HashicorpCloudPacker20230101Channel, error) {
	params := hcpPackerAPI.NewPackerServiceGetChannelParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.ChannelName = channelName

	resp, err := c.Packer.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Channel == nil {
		return nil, fmt.Errorf(
			"there is no channel with the name %s associated with the bucket %s",
			channelName, bucketName,
		)
	}

	return resp.Payload.Channel, nil
}
