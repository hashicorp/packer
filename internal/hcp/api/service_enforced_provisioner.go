// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"

	hcpPackerService "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// CreateEnforcedBlock creates a new enforced block in the HCP Packer registry.
// The block content contains raw HCL provisioner configuration that will be
// enforced on all builds for buckets linked to this enforced block.
func (c *Client) CreateEnforcedBlock(
	ctx context.Context,
	name string,
	blockContent string,
	version string,
	templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
	description string,
	labels map[string]string,
) (*hcpPackerModels.HashicorpCloudPacker20230101CreateEnforcedBlockResponse, error) {

	params := hcpPackerService.NewPackerServiceCreateEnforcedBlockParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101CreateEnforcedBlockBody{
		Name:                  name,
		BlockContent:          blockContent,
		Version:               version,
		TemplateType:          &templateType,
		AdditionalDescription: description,
		Labels:                labels,
	}

	resp, err := c.Packer.PackerServiceCreateEnforcedBlock(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetEnforcedBlock retrieves a single enforced block by its ID.
func (c *Client) GetEnforcedBlock(
	ctx context.Context,
	enforcedBlockID string,
) (*hcpPackerModels.HashicorpCloudPacker20230101GetEnforcedBlockResponse, error) {

	params := hcpPackerService.NewPackerServiceGetEnforcedBlockParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.EnforcedBlockID = enforcedBlockID

	resp, err := c.Packer.PackerServiceGetEnforcedBlock(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// ListEnforcedBlocks lists all enforced blocks in the current project.
func (c *Client) ListEnforcedBlocks(
	ctx context.Context,
) (*hcpPackerModels.HashicorpCloudPacker20230101ListEnforcedBlocksResponse, error) {

	params := hcpPackerService.NewPackerServiceListEnforcedBlocksParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID

	resp, err := c.Packer.PackerServiceListEnforcedBlocks(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// CreateEnforcedBlockVersion creates a new version of an existing enforced block.
// This allows updating the block content while keeping a version history.
func (c *Client) CreateEnforcedBlockVersion(
	ctx context.Context,
	enforcedBlockID string,
	blockContent string,
	version string,
	templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
	description string,
) (*hcpPackerModels.HashicorpCloudPacker20230101CreateEnforcedBlockVersionResponse, error) {

	params := hcpPackerService.NewPackerServiceCreateEnforcedBlockVersionParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.EnforcedBlockID = enforcedBlockID
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101CreateEnforcedBlockVersionBody{
		BlockContent:          blockContent,
		Version:               version,
		TemplateType:          &templateType,
		AdditionalDescription: description,
	}

	resp, err := c.Packer.PackerServiceCreateEnforcedBlockVersion(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetEnforcedBlockVersions retrieves all versions of an enforced block.
func (c *Client) GetEnforcedBlockVersions(
	ctx context.Context,
	enforcedBlockID string,
) (*hcpPackerModels.HashicorpCloudPacker20230101GetEnforcedBlockVersionsResponse, error) {

	params := hcpPackerService.NewPackerServiceGetEnforcedBlockVersionsParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.EnforcedBlockID = enforcedBlockID

	resp, err := c.Packer.PackerServiceGetEnforcedBlockVersions(params, nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

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
