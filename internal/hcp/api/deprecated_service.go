// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"
	"fmt"

	hcpPackerDeprecatedAPI "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	hcpPackerDeprecatedModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
)

type GetIterationOption func(*hcpPackerDeprecatedAPI.PackerServiceGetIterationParams)

var (
	GetIteration_byID = func(id string) GetIterationOption {
		return func(params *hcpPackerDeprecatedAPI.PackerServiceGetIterationParams) {
			params.IterationID = &id
		}
	}
	GetIteration_byFingerprint = func(fingerprint string) GetIterationOption {
		return func(params *hcpPackerDeprecatedAPI.PackerServiceGetIterationParams) {
			params.Fingerprint = &fingerprint
		}
	}
)

func (client *DeprecatedClient) GetIteration(
	ctx context.Context, bucketSlug string, opts ...GetIterationOption,
) (*hcpPackerDeprecatedModels.HashicorpCloudPackerIteration, error) {
	getItParams := hcpPackerDeprecatedAPI.NewPackerServiceGetIterationParams()
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

	return nil, fmt.Errorf(
		"something went wrong retrieving the iteration for bucket %s", bucketSlug,
	)
}

// GetChannel loads the named channel that is associated to the bucket slug . If the
// channel does not exist in HCP Packer, GetChannel returns an error.
func (client *DeprecatedClient) GetChannel(
	ctx context.Context, bucketSlug string, channelName string,
) (*hcpPackerDeprecatedModels.HashicorpCloudPackerChannel, error) {
	params := hcpPackerDeprecatedAPI.NewPackerServiceGetChannelParamsWithContext(ctx)
	params.LocationOrganizationID = client.OrganizationID
	params.LocationProjectID = client.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelName

	resp, err := client.Packer.PackerServiceGetChannel(params, nil)
	if err != nil {
		return nil, err
	}

	if resp.Payload.Channel == nil {
		return nil, fmt.Errorf(
			"there is no channel with the name %s associated with the bucket %s",
			channelName, bucketSlug,
		)
	}

	return resp.Payload.Channel, nil
}
