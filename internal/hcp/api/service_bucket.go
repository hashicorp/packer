package api

import (
	"context"

	hcpPackerService "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"google.golang.org/grpc/codes"
)

func (c *Client) CreateBucket(
	ctx context.Context, bucketName, bucketDescription string, bucketLabels map[string]string,
) (*hcpPackerService.PackerServiceCreateBucketOK, error) {

	createBktParams := hcpPackerService.NewPackerServiceCreateBucketParams()
	createBktParams.LocationOrganizationID = c.OrganizationID
	createBktParams.LocationProjectID = c.ProjectID
	createBktParams.Body = &hcpPackerModels.HashicorpCloudPacker20230101CreateBucketBody{
		Name:        bucketName,
		Description: bucketDescription,
		Labels:      bucketLabels,
	}

	return c.Packer.PackerServiceCreateBucket(createBktParams, nil)
}

func (c *Client) DeleteBucket(
	ctx context.Context, bucketName string,
) (*hcpPackerService.PackerServiceDeleteBucketOK, error) {

	deleteBktParams := hcpPackerService.NewPackerServiceDeleteBucketParamsWithContext(ctx)
	deleteBktParams.LocationOrganizationID = c.OrganizationID
	deleteBktParams.LocationProjectID = c.ProjectID
	deleteBktParams.BucketName = bucketName

	return c.Packer.PackerServiceDeleteBucket(deleteBktParams, nil)
}

// UpsertBucket tries to create a bucket on a HCP Packer Registry. If the bucket exists it will
// handle the error and update the bucket with the provided details.
func (c *Client) UpsertBucket(
	ctx context.Context, bucketName, bucketDescription string, bucketLabels map[string]string,
) error {

	// Create bucket if exist we continue as is, eventually we want to treat this like an upsert.
	_, err := c.CreateBucket(ctx, bucketName, bucketDescription, bucketLabels)
	if err != nil && !CheckErrorCode(err, codes.AlreadyExists) {
		return err
	}

	if err == nil {
		return nil
	}

	params := hcpPackerService.NewPackerServiceUpdateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101UpdateBucketBody{
		Description: bucketDescription,
		Labels:      bucketLabels,
	}
	_, err = c.Packer.PackerServiceUpdateBucket(params, nil)

	return err
}
