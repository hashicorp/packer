package packer_registry

import (
	"context"
	"errors"
	"fmt"

	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
)

// CreateBucket creates a bucket on a HCP Packer Registry.
func CreateBucket(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBucketRequest) (string, error) {

	params := packerSvc.NewCreateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.Body = input

	resp, err := client.Packer.CreateBucket(params, nil)
	if err != nil {
		return "", err
	}

	return resp.Payload.Bucket.ID, nil
}

// UpsertBucket tries to create a bucket on a HCP Packer Registry.
// If the bucket exists it will handle the error and update the bucket with the provided details.
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
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = input.BucketSlug
	params.Body = &models.HashicorpCloudPackerUpdateBucketRequest{
		Description: input.Description,
		Labels:      input.Labels,
	}
	_, err = client.Packer.UpdateBucket(params, nil)

	return err
}

// CreateIteration creates an Iteration for some Bucket on a HCP Packer Registry.
func CreateIteration(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateIterationRequest) (*models.HashicorpCloudPackerIteration, error) {
	params := packerSvc.NewCreateIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = input.BucketSlug
	params.Body = input

	it, err := client.Packer.CreateIteration(params, nil)
	if err != nil {
		return nil, err
	}

	return it.Payload.Iteration, nil
}

// GetIteration queries the HCP Packer registry for an existing bucket iteration.
func GetIteration(ctx context.Context, client *Client, bucketslug string, fingerprint string) (*models.HashicorpCloudPackerIteration, error) {
	params := packerSvc.NewGetIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketslug

	// The identifier can be either fingerprint, iterationid, or incremental version
	// for now, we only care about fingerprint so we're hardcoding it.
	params.Fingerprint = &fingerprint

	it, err := client.Packer.GetIteration(params, nil)
	if err != nil {
		return nil, err
	}

	return it.Payload.Iteration, nil
}

// CreateBuild create a build entry to track for the IterationID and BucketSlug defined within input.
func CreateBuild(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBuildRequest) (string, error) {
	params := packerSvc.NewCreateBuildParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = input.BucketSlug
	params.BuildIterationID = input.Build.IterationID
	params.Body = input

	resp, err := client.Packer.CreateBuild(params, nil)
	if err != nil {
		return "", err
	}

	return resp.Payload.Build.ID, nil
}

// ListBuilds queries an Iteration on HCP Packer registry for all of it's associated builds.
// Currently all builds are returned regardless of status.
func ListBuilds(ctx context.Context, client *Client, bucketSlug string, iterationID string) ([]*models.HashicorpCloudPackerBuild, error) {
	params := packerSvc.NewListBuildsParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.IterationID = iterationID

	resp, err := client.Packer.ListBuilds(params, nil)
	if err != nil {
		return []*models.HashicorpCloudPackerBuild{}, err
	}

	return resp.Payload.Builds, nil
}

// UpdateBuild updates a single iteration build entry with the incoming input data.
func UpdateBuild(ctx context.Context, client *Client, input *models.HashicorpCloudPackerUpdateBuildRequest) (string, error) {
	params := packerSvc.NewUpdateBuildParamsWithContext(ctx)
	params.BuildID = input.BuildID
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.Body = input

	resp, err := client.Packer.UpdateBuild(params, nil)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("Not sure why response is nil")
	}

	return resp.Payload.Build.ID, nil
}

// GetChannel loads the iterationId associated with a current channel. If
// the channel does not exist in HCP Packer, GetChannel returns an error.
func GetIterationFromChannel(ctx context.Context, client *Client, bucketSlug string, channelName string) (*models.HashicorpCloudPackerIteration, error) {
	params := packerSvc.NewGetChannelParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelName

	resp, err := client.Packer.GetChannel(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Channel != nil {
		if resp.Payload.Channel.Pointer != nil {
			// The channel payload contains a pointer, which points to the iteration.
			// Reach into the pointer to get the desired iteration.
			return resp.Payload.Channel.Pointer.Iteration, nil
		}
		return nil, fmt.Errorf("there is no iteration associated with the channel %s",
			channelName)
	}

	return nil, fmt.Errorf("there is no channel with the name %s associated with the bucket %s",
		channelName, bucketSlug)
}
