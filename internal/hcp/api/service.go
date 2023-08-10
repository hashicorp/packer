// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
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
	createBktParams.Body = packer_service.PackerServiceCreateBucketBody{
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
	if err != nil && !CheckErrorCode(err, codes.AlreadyExists) {
		return err
	}

	if err == nil {
		return nil
	}

	params := packer_service.NewPackerServiceUpdateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Body = packer_service.PackerServiceUpdateBucketBody{
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
	templateType models.HashicorpCloudPackerIterationTemplateType,
) (*packer_service.PackerServiceCreateIterationOK, error) {

	params := packer_service.NewPackerServiceCreateIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Body = packer_service.PackerServiceCreateIterationBody{
		Fingerprint:  fingerprint,
		TemplateType: templateType.Pointer(),
	}

	return client.Packer.PackerServiceCreateIteration(params, nil)
}

type GetIterationOption func(*packer_service.PackerServiceGetIterationParams)

var (
	GetIteration_byID = func(id string) GetIterationOption {
		return func(params *packer_service.PackerServiceGetIterationParams) {
			params.IterationID = &id
		}
	}
	GetIteration_byFingerprint = func(fingerprint string) GetIterationOption {
		return func(params *packer_service.PackerServiceGetIterationParams) {
			params.Fingerprint = &fingerprint
		}
	}
)

func (client *Client) GetIteration(ctx context.Context, bucketSlug string, opts ...GetIterationOption) (*models.HashicorpCloudPackerIteration, error) {
	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = client.OrganizationID
	getItParams.LocationProjectID = client.ProjectID
	getItParams.BucketSlug = bucketSlug

	for _, opt := range opts {
		opt(getItParams)
	}

	resp, err := client.Packer.PackerServiceGetIteration(getItParams, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Iteration != nil {
		return resp.Payload.Iteration, nil
	}

	return nil, fmt.Errorf("something went wrong retrieving the iteration for bucket %s", bucketSlug)
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
	params.Body = packer_service.PackerServiceCreateBuildBody{
		Fingerprint: fingerprint,
		Build: &models.HashicorpCloudPackerBuildCreateBody{
			ComponentType: componentType,
			PackerRunUUID: runUUID,
			Status:        status.Pointer(),
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
	sourceIterationID string,
	sourceChannelID string,
	labels map[string]string,
	status models.HashicorpCloudPackerBuildStatus,
	images []*models.HashicorpCloudPackerImageCreateBody,
) (string, error) {

	params := packer_service.NewPackerServiceUpdateBuildParamsWithContext(ctx)
	params.BuildID = buildID
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID

	params.Body = packer_service.PackerServiceUpdateBuildBody{
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			Images:            images,
			PackerRunUUID:     runUUID,
			Labels:            labels,
			Status:            status.Pointer(),
			CloudProvider:     cloudProvider,
			SourceImageID:     sourceImageID,
			SourceIterationID: sourceIterationID,
			SourceChannelID:   sourceChannelID,
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

// GetChannel loads the named channel that is associated to the bucket slug . If the
// channel does not exist in HCP Packer, GetChannel returns an error.
func (client *Client) GetChannel(ctx context.Context, bucketSlug string, channelName string) (*models.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceGetChannelParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelName

	resp, err := client.Packer.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Channel == nil {
		return nil, fmt.Errorf("there is no channel with the name %s associated with the bucket %s",
			channelName, bucketSlug)
	}

	return resp.Payload.Channel, nil
}
