package packer_registry

import (
	"context"
	"errors"

	"github.com/go-openapi/runtime"
	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
)

// CreateBucket creates a bucket on a HCP Packer Registry.
func CreateBucket(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBucketRequest) (string, error) {

	params := packerSvc.NewCreateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.Body = input

	resp, err := client.Packer.CreateBucket(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return "", err
	}

	return resp.Payload.Bucket.ID, nil
}

// UpsertBucket tries to create a bucket on a HCP Packer Registry. If the bucket exists it will handle the error
// and update the bucket with the provided details.
func UpsertBucket(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBucketRequest) error {

	// Create bucket if exist we continue as is, eventually we want to treat this like an upsert
	_, err := CreateBucket(ctx, client, input)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return err
	}

	if err == nil {
		return nil
	}

	params := packerSvc.NewUpdateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	_, err = client.Packer.UpdateBucket(params, nil, func(*runtime.ClientOperation) {})

	return err
}

/*
CreateIteration creates an Iteration for some Bucket on a HCP Packer Registry for the given input
and returns the ID associated with the persisted Bucket iteration.

input: *models.HashicorpCloudPackerCreateIterationRequest{BucketSlug: "bucket name"
*/
func CreateIteration(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateIterationRequest) (string, error) {
	// Create/find iteration
	params := packerSvc.NewCreateIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	params.Body = input

	it, err := client.Packer.CreateIteration(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return "", err
	}

	return it.Payload.Iteration.ID, nil
}

func CreateBuild(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBuildRequest) (string, error) {
	params := packerSvc.NewCreateBuildParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	params.BuildIterationID = input.Build.IterationID
	params.Body = input

	resp, err := client.Packer.CreateBuild(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return "", err
	}

	return resp.Payload.Build.ID, nil
}

func UpdateBuild(ctx context.Context, client *Client, input *models.HashicorpCloudPackerUpdateBuildRequest) (string, error) {
	params := packerSvc.NewUpdateBuildParamsWithContext(ctx)
	params.BuildID = input.BuildID
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.Body = input

	resp, err := client.Packer.UpdateBuild(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("Not sure why response is nil")
	}

	return resp.Payload.Build.ID, nil
}
