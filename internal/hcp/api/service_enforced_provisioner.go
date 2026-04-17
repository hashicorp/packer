// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"

	hcpPackerService "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// GetEnforcedBlocksForBucket fetches all enforced blocks linked to a bucket.
// This is the key method used during packer build to auto-inject provisioners.
// The response includes EnforcedBlockDetail entries each with an active version
// containing the raw HCL block_content to be parsed and injected.
func (c *Client) GetEnforcedBlocksForBucket(
	ctx context.Context,
	bucketName string,
) (*hcpPackerModels.HashicorpCloudPacker20230101GetEnforcedBlocksByBucketResponse, error) {

	params := hcpPackerService.NewPackerServiceGetEnforcedBlocksByBucketParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName

	resp, err := c.Packer.PackerServiceGetEnforcedBlocksByBucket(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
