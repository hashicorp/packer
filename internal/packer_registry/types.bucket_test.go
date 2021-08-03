package packer_registry

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
)

func TestInitialize_NewBucketNewIteration(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.CreateBucketCalled {
		t.Errorf("expected a call to CreateBucket but it didn't happen")
	}

	if !mockService.CreateIterationCalled {
		t.Errorf("expected a call to CreateIteration but it didn't happen")
	}

	if !mockService.CreateBuildCalled {
		t.Errorf("expected a call to CreateBuild but it didn't happen")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to created but it didn't")
	}

	if _, ok := b.Iteration.builds.Load("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

}

func TestInitialize_ExistingBucketNewIteration(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()
	mockService.BucketAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.UpdateBucketCalled {
		t.Errorf("expected call to UpdateBucket but it didn't happen")
	}

	if !mockService.CreateIterationCalled {
		t.Errorf("expected a call to CreateIteration but it didn't happen")
	}

	if !mockService.CreateBuildCalled {
		t.Errorf("expected a call to CreateBuild but it didn't happen")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to created but it didn't")
	}

	if _, ok := b.Iteration.builds.Load("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

}

func TestInitialize_ExistingBucketExistingIteration(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if mockService.CreateBucketCalled {
		t.Errorf("unexpected call to CreateBucket")
	}

	if !mockService.UpdateBucketCalled {
		t.Errorf("expected call to UpdateBucket but it didn't happen")
	}

	if mockService.CreateIterationCalled {
		t.Errorf("unexpected a call to CreateIteration")
	}

	if !mockService.GetIterationCalled {
		t.Errorf("expected a call to GetIteration but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("unexpected a call to CreateBuild")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to created but it didn't")
	}

	loadedBuild, ok := b.Iteration.builds.Load("happycloud.image")
	if !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

	existingBuild, ok := loadedBuild.(*Build)
	if !ok {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid")
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}
}

func TestInitialize_ExistingBucketCompleteIteration(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true
	mockService.IterationCompleted = true
	mockService.BuildAlreadyDone = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err == nil {
		t.Errorf("Calling initialize on a completed Iteration should fail hard")
	}

	if mockService.CreateIterationCalled {
		t.Errorf("unexpected a call to CreateIteration")
	}

	if !mockService.GetIterationCalled {
		t.Errorf("expected a call to GetIteration but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("unexpected a call to CreateBuild")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to returned but it didn't")
	}
}

func TestUpdateBuildStatus(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	loadedBuild, ok := b.Iteration.builds.Load("happycloud.image")
	if !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

	existingBuild, ok := loadedBuild.(*Build)
	if !ok {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid")
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}

	err = b.UpdateBuildStatus(context.TODO(), "happycloud.image", models.HashicorpCloudPackerBuildStatusRUNNING)
	if err != nil {
		t.Errorf("unexpected failure for PublishBuildStatus: %v", err)
	}

	reloadedBuild, ok := b.Iteration.builds.Load("happycloud.image")
	if !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

	existingBuild, ok = reloadedBuild.(*Build)
	if !ok {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid")
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}

func TestUpdateBuildStatus_DONENoImages(t *testing.T) {
	//nolint:errcheck
	os.Setenv("HCP_PACKER_BUILD_FINGEPRINT", "testnumber")
	defer os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
	mockService := NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &Client{
			Packer: mockService,
		},
	}

	var err error
	b.Iteration, err = NewIteration(IterationOptions{})
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	loadedBuild, ok := b.Iteration.builds.Load("happycloud.image")
	if !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

	existingBuild, ok := loadedBuild.(*Build)
	if !ok {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid")
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}

	//nolint:errcheck
	b.UpdateBuildStatus(context.TODO(), "happycloud.image", models.HashicorpCloudPackerBuildStatusRUNNING)

	err = b.UpdateBuildStatus(context.TODO(), "happycloud.image", models.HashicorpCloudPackerBuildStatusDONE)
	if err == nil {
		t.Errorf("expected failure for PublishBuildStatus when setting status to DONE with no images")
	}

	reloadedBuild, ok := b.Iteration.builds.Load("happycloud.image")
	if !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

	existingBuild, ok = reloadedBuild.(*Build)
	if !ok {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid")
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}

//func (b *Bucket) PublishBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {}
