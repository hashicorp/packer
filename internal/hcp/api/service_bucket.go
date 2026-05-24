package api

import (
	"context"
	"reflect"

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

// UpsertBucket will create or update a bucket. It calls GetBucket first, if the bucket is not found it creates that bucket
// If GetBucket succeeded we then call UpdateBucket description and bucket labels in case they've changed.
// GetBucket is used instead of CreateBucket since users with bucket level access to specific existing buckets can not create new buckets.
func (c *Client) UpsertBucket(
	ctx context.Context, bucketName, bucketDescription string, bucketLabels map[string]string,
) error {

	getParams := hcpPackerService.NewPackerServiceGetBucketParamsWithContext(ctx)
	getParams.LocationOrganizationID = c.OrganizationID
	getParams.LocationProjectID = c.ProjectID
	getParams.BucketName = bucketName

	resp, err := c.Packer.PackerServiceGetBucket(getParams, nil)
	if err != nil {
		if CheckErrorCode(err, codes.NotFound) {
			_, err = c.CreateBucket(ctx, bucketName, bucketDescription, bucketLabels)
		}
		return err
	}

	if resp != nil && resp.Payload != nil && bucketMetadataMatches(resp.Payload.Bucket, bucketDescription, bucketLabels) {
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

func bucketMetadataMatches(
	bucket *hcpPackerModels.HashicorpCloudPacker20230101Bucket,
	description string,
	labels map[string]string,
) bool {
	if bucket == nil {
		return false
	}

	if bucket.Description != description {
		return false
	}

	if len(bucket.Labels) == 0 && len(labels) == 0 {
		return true
	}

	return reflect.DeepEqual(bucket.Labels, labels)
}
