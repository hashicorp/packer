package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
)

func (client *Client) CreateBucket(
	ctx context.Context,
	bucketSlug,
	bucketDescription string,
	bucketLabels map[string]string,
) (*packer_service.PackerServiceCreateBucketOK, error) {

	createBktParams := packer_service.NewPackerServiceCreateBucketParams()
	createBktParams.LocationOrganizationID = client.OrganizationID
	createBktParams.LocationProjectID = client.ProjectID
	createBktParams.Body = &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug:  bucketSlug,
		Description: bucketDescription,
		Labels:      bucketLabels,
	}

	return client.Packer.PackerServiceCreateBucket(createBktParams, nil)
}

func (client *Client) DeleteBucket(
	ctx context.Context,
	bucketSlug string,
) (*packer_service.PackerServiceDeleteBucketOK, error) {

	deleteBktParams := packer_service.NewPackerServiceDeleteBucketParamsWithContext(ctx)
	deleteBktParams.LocationOrganizationID = client.OrganizationID
	deleteBktParams.LocationProjectID = client.ProjectID
	deleteBktParams.BucketSlug = bucketSlug

	return client.Packer.PackerServiceDeleteBucket(deleteBktParams, nil)
}

// UpsertBucket tries to create a bucket on a HCP Packer Registry. If the bucket
// exists it will handle the error and update the bucket with the provided
// details.
func (client *Client) UpsertBucket(
	ctx context.Context,
	bucketSlug,
	bucketDescription string,
	bucketLabels map[string]string,
) error {

	// Create bucket if exist we continue as is, eventually we want to treat
	// this like an upsert
	_, err := client.CreateBucket(ctx, bucketSlug, bucketDescription, bucketLabels)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return err
	}

	if err == nil {
		return nil
	}

	params := packer_service.NewPackerServiceUpdateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Body = &models.HashicorpCloudPackerUpdateBucketRequest{
		BucketSlug:  bucketSlug,
		Description: bucketDescription,
		Labels:      bucketLabels,
	}
	_, err = client.Packer.PackerServiceUpdateBucket(params, nil)

	return err
}

func (client *Client) CreateIteration(
	ctx context.Context,
	bucketSlug,
	fingerprint string,
) (*packer_service.PackerServiceCreateIterationOK, error) {

	params := packer_service.NewPackerServiceCreateIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Body = &models.HashicorpCloudPackerCreateIterationRequest{
		Fingerprint: fingerprint,
		BucketSlug:  bucketSlug,
	}

	return client.Packer.PackerServiceCreateIteration(params, nil)
}

func (client *Client) GetIteration_byID(
	ctx context.Context,
	bucketSlug,
	id string,
) (*packer_service.PackerServiceGetIterationOK, error) {

	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = client.OrganizationID
	getItParams.LocationProjectID = client.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.IterationID = &id

	return client.Packer.PackerServiceGetIteration(getItParams, nil)
}

func (client *Client) GetIteration_byFingerprint(
	ctx context.Context,
	bucketSlug,
	fingerprint string,
) (*packer_service.PackerServiceGetIterationOK, error) {

	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = client.OrganizationID
	getItParams.LocationProjectID = client.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.Fingerprint = &fingerprint

	return client.Packer.PackerServiceGetIteration(getItParams, nil)
}

func (client *Client) CreateBuild(
	ctx context.Context,
	bucketSlug,
	runUUID,
	iterationID,
	fingerprint,
	componentType string,
	status models.HashicorpCloudPackerBuildStatus,
) (*packer_service.PackerServiceCreateBuildOK, error) {

	params := packer_service.NewPackerServiceCreateBuildParamsWithContext(ctx)

	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.IterationID = iterationID
	params.Body = &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug:  bucketSlug,
		IterationID: iterationID,
		Fingerprint: fingerprint,
		Build: &models.HashicorpCloudPackerBuildCreateBody{
			ComponentType: componentType,
			PackerRunUUID: runUUID,
			Status:        status,
		},
	}

	return client.Packer.PackerServiceCreateBuild(params, nil)
}

// ListBuilds queries an Iteration on HCP Packer registry for all of it's
// associated builds. Currently all builds are returned regardless of status.
func (client *Client) ListBuilds(
	ctx context.Context,
	bucketSlug string,
	iterationID string,
) ([]*models.HashicorpCloudPackerBuild, error) {

	params := packer_service.NewPackerServiceListBuildsParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.IterationID = iterationID

	resp, err := client.Packer.PackerServiceListBuilds(params, nil)
	if err != nil {
		return []*models.HashicorpCloudPackerBuild{}, err
	}

	return resp.Payload.Builds, nil
}

// UpdateBuild updates a single iteration build entry with the incoming input
// data.
func (client *Client) UpdateBuild(
	ctx context.Context,
	buildID,
	runUUID,
	cloudProvider,
	sourceImageID string,
	labels map[string]string,
	status models.HashicorpCloudPackerBuildStatus,
	images []*models.HashicorpCloudPackerImageCreateBody,
) (string, error) {

	params := packer_service.NewPackerServiceUpdateBuildParamsWithContext(ctx)
	params.BuildID = buildID
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID

	params.Body = &models.HashicorpCloudPackerUpdateBuildRequest{
		BuildID: buildID,
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			Images:        images,
			PackerRunUUID: runUUID,
			Labels:        labels,
			Status:        status,
			CloudProvider: cloudProvider,
			SourceImageID: sourceImageID,
		},
	}

	resp, err := client.Packer.PackerServiceUpdateBuild(params, nil)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("Not sure why response is nil")
	}

	return resp.Payload.Build.ID, nil
}

// GetChannel loads the iterationId associated with a current channel. If the
// channel does not exist in HCP Packer, GetChannel returns an error.
func (client *Client) GetIterationFromChannel(
	ctx context.Context,
	bucketSlug string,
	channelName string,
) (*models.HashicorpCloudPackerIteration, error) {

	params := packer_service.NewPackerServiceGetChannelParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelName

	resp, err := client.Packer.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Channel != nil {
		if resp.Payload.Channel.Pointer != nil {
			// The channel payload contains a pointer, which points to the
			// iteration. Reach into the pointer to get the desired iteration.
			return resp.Payload.Channel.Pointer.Iteration, nil
		}
		return nil, fmt.Errorf("there is no iteration associated with the channel %s",
			channelName)
	}

	return nil, fmt.Errorf("there is no channel with the name %s associated with the bucket %s",
		channelName, bucketSlug)
}
