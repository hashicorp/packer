// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"testing"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	hcpPackerAPI "github.com/hashicorp/packer/internal/hcp/api"
)

func TestInitialize_NewBucketNewVersion(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.CreateBucketCalled {
		t.Errorf("expected a call to CreateBucket but it didn't happen")
	}

	if !mockService.CreateVersionCalled {
		t.Errorf("expected a call to CreateVersion but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("Didn't expect a call to CreateBuild")
	}

	if b.Version.ID != "version-id" {
		t.Errorf("expected a version to created but it didn't")
	}

	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	if !mockService.CreateBuildCalled {
		t.Errorf("Expected a call to CreateBuild but it didn't happen")
	}

	if ok := b.Version.HasBuild("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}
}

func TestInitialize_UnsetTemplateTypeError(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeTEMPLATETYPEUNSET)
	if err == nil {
		t.Fatalf("unexpected success")
	}

	t.Logf("version creating failed as expected: %s", err)
}

func TestInitialize_ExistingBucketNewVersion(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.UpdateBucketCalled {
		t.Errorf("expected call to UpdateBucket but it didn't happen")
	}

	if !mockService.CreateVersionCalled {
		t.Errorf("expected a call to CreateVersion but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("Didn't expect a call to CreateBuild")
	}

	if b.Version.ID != "version-id" {
		t.Errorf("expected a version to created but it didn't")
	}

	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	if !mockService.CreateBuildCalled {
		t.Errorf("Expected a call to CreateBuild but it didn't happen")
	}

	if ok := b.Version.HasBuild("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

}

func TestInitialize_ExistingBucketExistingVersion(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.VersionAlreadyExist = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if mockService.CreateBucketCalled {
		t.Errorf("unexpected call to CreateBucket")
	}

	if !mockService.UpdateBucketCalled {
		t.Errorf("expected call to UpdateBucket but it didn't happen")
	}

	if mockService.CreateVersionCalled {
		t.Errorf("unexpected a call to CreateVersion")
	}

	if !mockService.GetVersionCalled {
		t.Errorf("expected a call to GetVersion but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("unexpected a call to CreateBuild")
	}

	if b.Version.ID != "version-id" {
		t.Errorf("expected a version to created but it didn't")
	}

	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Version.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}
}

func TestInitialize_ExistingBucketCompleteVersion(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.VersionAlreadyExist = true
	mockService.VersionCompleted = true
	mockService.BuildAlreadyDone = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err == nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if mockService.CreateVersionCalled {
		t.Errorf("unexpected call to CreateVersion")
	}

	if !mockService.GetVersionCalled {
		t.Errorf("expected a call to GetVersion but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("unexpected call to CreateBuild")
	}

	if b.Version.ID != "version-id" {
		t.Errorf("expected a version to be returned but it wasn't")
	}
}

func TestUpdateBuildStatus(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.VersionAlreadyExist = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Version.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}

	err = b.UpdateBuildStatus(context.TODO(), "happycloud.image", hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING)
	if err != nil {
		t.Errorf("unexpected failure for PublishBuildStatus: %v", err)
	}

	existingBuild, err = b.Version.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}

func TestUpdateBuildStatus_DONENoImages(t *testing.T) {
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.VersionAlreadyExist = true

	b := &Bucket{
		Name: "TestBucket",
		client: &hcpPackerAPI.Client{
			Packer: mockService,
		},
	}

	b.Version = NewVersion()
	err := b.Version.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Version.expectedBuilds = append(b.Version.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateVersion(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Version.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}

	//nolint:errcheck
	_ = b.UpdateBuildStatus(context.TODO(), "happycloud.image", hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING)

	err = b.UpdateBuildStatus(context.TODO(), "happycloud.image", hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE)
	if err == nil {
		t.Errorf("expected failure for PublishBuildStatus when setting status to DONE with no images")
	}

	existingBuild, err = b.Version.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}
