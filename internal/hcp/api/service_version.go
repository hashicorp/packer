package api

import (
	"context"
	"fmt"

	hcpPackerAPI "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

const incompleteVersionName = "v0"

// IsVersionComplete returns if the given version is completed or not.
//
// The best way to know if the version is completed or not is from the name of the version. All version that are
// incomplete are named "v0".
func (c *Client) IsVersionComplete(version *hcpPackerModels.HashicorpCloudPacker20230101Version) bool {
	return version.Name != incompleteVersionName
}

func (c *Client) CreateVersion(
	ctx context.Context,
	bucketName,
	fingerprint string,
	templateType hcpPackerModels.HashicorpCloudPacker20230101TemplateType,
) (*hcpPackerAPI.PackerServiceCreateVersionOK, error) {

	params := hcpPackerAPI.NewPackerServiceCreateVersionParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101CreateVersionBody{
		Fingerprint:  fingerprint,
		TemplateType: templateType.Pointer(),
	}

	return c.Packer.PackerServiceCreateVersion(params, nil)
}

func (c *Client) GetVersion(
	ctx context.Context, bucketName string, fingerprint string,
) (*hcpPackerModels.HashicorpCloudPacker20230101Version, error) {
	params := hcpPackerAPI.NewPackerServiceGetVersionParams()
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Fingerprint = fingerprint

	resp, err := c.Packer.PackerServiceGetVersion(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Version != nil {
		return resp.Payload.Version, nil
	}

	return nil, fmt.Errorf(
		"something went wrong retrieving the version for bucket %s", bucketName,
	)
}
