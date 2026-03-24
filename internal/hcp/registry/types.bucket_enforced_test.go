// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"errors"
	"fmt"
	"testing"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	hcpPackerAPI "github.com/hashicorp/packer/internal/hcp/api"
	"google.golang.org/grpc/codes"
)

func TestBucket_FetchEnforcedBlocks_ReturnsAllBlocks(t *testing.T) {
	hcl2Type := hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2
	jsonType := hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeJSON

	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.GetEnforcedBlocksByBucketResp = &hcpPackerModels.HashicorpCloudPacker20230101GetEnforcedBlocksByBucketResponse{
		EnforcedBlockDetail: []*hcpPackerModels.HashicorpCloudPacker20230101EnforcedBlockDetail{
			{
				ID:   "hcl-id",
				Name: "hcl-block",
				Version: &hcpPackerModels.HashicorpCloudPacker20230101EnforcedBlockVersion{
					ID:           "hcl-v1",
					Version:      "1",
					BlockContent: "provisioner \"shell\" {}",
					TemplateType: &hcl2Type,
				},
			},
			{
				ID:   "json-id",
				Name: "json-block",
				Version: &hcpPackerModels.HashicorpCloudPacker20230101EnforcedBlockVersion{
					ID:           "json-v1",
					Version:      "1",
					BlockContent: "{\"provisioner\":[{\"shell\":{}}]}",
					TemplateType: &jsonType,
				},
			},
			{
				ID:   "unset-id",
				Name: "unset-block",
				Version: &hcpPackerModels.HashicorpCloudPacker20230101EnforcedBlockVersion{
					ID:           "unset-v1",
					Version:      "1",
					BlockContent: "provisioner \"shell\" {}",
				},
			},
		},
	}

	bucket := &Bucket{
		Name: "test-bucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	err := bucket.FetchEnforcedBlocks(context.Background())
	if err != nil {
		t.Fatalf("FetchEnforcedBlocks() unexpected error: %v", err)
	}

	if len(bucket.EnforcedBlocks) != 3 {
		t.Fatalf("FetchEnforcedBlocks() got %d blocks, want 3", len(bucket.EnforcedBlocks))
	}

	if bucket.EnforcedBlocks[0].Name != "hcl-block" {
		t.Fatalf("first block name = %q, want %q", bucket.EnforcedBlocks[0].Name, "hcl-block")
	}

	if bucket.EnforcedBlocks[1].Name != "json-block" {
		t.Fatalf("second block name = %q, want %q", bucket.EnforcedBlocks[1].Name, "json-block")
	}

	if bucket.EnforcedBlocks[2].Name != "unset-block" {
		t.Fatalf("third block name = %q, want %q", bucket.EnforcedBlocks[2].Name, "unset-block")
	}
}

func TestBucket_FetchEnforcedBlocks_ReturnsErrorOnServiceFailure(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.GetEnforcedBlocksByBucketErr = errors.New("service unavailable")

	bucket := &Bucket{
		Name: "test-bucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	err := bucket.FetchEnforcedBlocks(context.Background())
	if err == nil {
		t.Fatal("FetchEnforcedBlocks() expected error, got nil")
	}
}

func TestBucket_FetchEnforcedBlocks_NotFoundIsNonFatal(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.GetEnforcedBlocksByBucketErr = fmt.Errorf("Code:%d %s", codes.NotFound, codes.NotFound.String())

	bucket := &Bucket{
		Name: "test-bucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	err := bucket.FetchEnforcedBlocks(context.Background())
	if err != nil {
		t.Fatalf("FetchEnforcedBlocks() expected nil error for NotFound, got: %v", err)
	}
}
