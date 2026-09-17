// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	hcpPackerService "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

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

// ResolveEnforcedProvisioners calls the vNext resolver endpoint
// (POST .../buckets/{bucket_name}:resolve-enforced-provisioners). The resolver
// is the source of truth for the ordered effective provisioner set, the bucket
// policy mode (mandatory/advisory), and the resolution audit context.
//
// ifNoneMatch, when non-empty, is sent as the If-None-Match cache-validation
// header. The freshness token of record for subsequent calls is the body's
// audit_context.etag (the generated SDK does not model the ETag response header
// or a 304 response).
func (c *Client) ResolveEnforcedProvisioners(
	ctx context.Context,
	bucketName string,
	ifNoneMatch string,
	buildCorrelationID string,
	cliVersion string,
) (*hcpPackerModels.HashicorpCloudPacker20230101ResolveEnforcedProvisionersResponse, error) {

	params := hcpPackerService.NewPackerServiceResolveEnforcedProvisionersParamsWithContext(ctx)
	params.LocationOrganizationID = c.OrganizationID
	params.LocationProjectID = c.ProjectID
	params.BucketName = bucketName
	params.Body = &hcpPackerModels.HashicorpCloudPacker20230101ResolveEnforcedProvisionersBody{
		BuildCorrelationID: buildCorrelationID,
		CliVersion:         cliVersion,
	}

	resp, err := c.Packer.PackerServiceResolveEnforcedProvisioners(params, nil, withIfNoneMatch(ifNoneMatch))
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// withIfNoneMatch returns a ClientOption that sets the If-None-Match request
// header on the resolver call, wrapping the params writer so the header is added
// after the generated body/path params are written. A no-op when etag is empty.
func withIfNoneMatch(etag string) hcpPackerService.ClientOption {
	return func(op *runtime.ClientOperation) {
		if etag == "" {
			return
		}
		inner := op.Params
		op.Params = runtime.ClientRequestWriterFunc(func(req runtime.ClientRequest, reg strfmt.Registry) error {
			if inner != nil {
				if err := inner.WriteToRequest(req, reg); err != nil {
					return err
				}
			}
			return req.SetHeaderParam("If-None-Match", etag)
		})
	}
}
