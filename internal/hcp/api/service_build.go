package api

import (
	"context"
	"fmt"

	hcpPackerAPI "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

func (c *Client) CreateBuild(
	ctx context.Context, bucketName, runUUID, fingerprint, componentType string,
	buildStatus hcpPackerModels.HashicorpCloudPacker20230101BuildStatus,
) (*hcpPackerAPI.PackerServiceCreateBuildOK, error) {

	params := hcpPackerAPI.NewPackerServiceCreateBuildParamsWithContext(ctx)

	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Fingerprint = fingerprint
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101CreateBuildBody{
		ComponentType: componentType,
		PackerRunUUID: runUUID,
		Status:        &buildStatus,
	}

	return c.Packer.PackerServiceCreateBuild(params, nil)
}

// ListBuilds queries a Version on HCP Packer registry for all of it's associated builds.
// Currently, all builds are returned regardless of status.
func (c *Client) ListBuilds(
	ctx context.Context, bucketName, fingerprint string,
) ([]*hcpPackerModels.HashicorpCloudPacker20230101Build, error) {

	params := hcpPackerAPI.NewPackerServiceListBuildsParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Fingerprint = fingerprint

	resp, err := c.Packer.PackerServiceListBuilds(params, nil)
	if err != nil {
		return []*hcpPackerModels.HashicorpCloudPacker20230101Build{}, err
	}

	return resp.Payload.Builds, nil
}

// UpdateBuild updates a single build in a version with the incoming input data.
func (c *Client) UpdateBuild(
	ctx context.Context,
	bucketName, fingerprint string,
	buildID, runUUID, platform, sourceExternalIdentifier string,
	parentVersionID string,
	parentChannelID string,
	buildLabels map[string]string,
	buildStatus hcpPackerModels.HashicorpCloudPacker20230101BuildStatus,
	artifacts []*hcpPackerModels.HashicorpCloudPacker20230101ArtifactCreateBody,
	metadata *hcpPackerModels.HashicorpCloudPacker20230101BuildMetadata,
) (string, error) {

	params := hcpPackerAPI.NewPackerServiceUpdateBuildParamsWithContext(ctx)
	params.BuildID = buildID
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Fingerprint = fingerprint

	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101UpdateBuildBody{
		Artifacts:                artifacts,
		Labels:                   buildLabels,
		PackerRunUUID:            runUUID,
		ParentChannelID:          parentChannelID,
		ParentVersionID:          parentVersionID,
		Platform:                 platform,
		SourceExternalIdentifier: sourceExternalIdentifier,
		Status:                   &buildStatus,
		Metadata:                 metadata,
	}

	resp, err := c.Packer.PackerServiceUpdateBuild(params, nil)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf(
			"something went wrong retrieving the build %s from bucket %s", buildID, bucketName,
		)
	}

	return resp.Payload.Build.ID, nil
}
