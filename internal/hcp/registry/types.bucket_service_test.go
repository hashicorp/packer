package registry

import (
	"context"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/packer/internal/hcp/api"
)

func TestInitialize_NewBucketNewIteration(t *testing.T) {
	mockService := api.NewMockPackerClientService()

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.CreateBucketCalled {
		t.Errorf("expected a call to CreateBucket but it didn't happen")
	}

	if !mockService.CreateIterationCalled {
		t.Errorf("expected a call to CreateIteration but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("Didn't expect a call to CreateBuild")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to created but it didn't")
	}

	err = b.populateIteration(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	if !mockService.CreateBuildCalled {
		t.Errorf("Expected a call to CreateBuild but it didn't happen")
	}

	if ok := b.Iteration.HasBuild("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}
}

func TestInitialize_UnsetTemplateTypeError(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeTEMPLATETYPEUNSET)
	if err == nil {
		t.Fatalf("unexpected success")
	}

	t.Logf("iteration creating failed as expected: %s", err)
}

func TestInitialize_ExistingBucketNewIteration(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if !mockService.UpdateBucketCalled {
		t.Errorf("expected call to UpdateBucket but it didn't happen")
	}

	if !mockService.CreateIterationCalled {
		t.Errorf("expected a call to CreateIteration but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("Didn't expect a call to CreateBuild")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to created but it didn't")
	}

	err = b.populateIteration(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	if !mockService.CreateBuildCalled {
		t.Errorf("Expected a call to CreateBuild but it didn't happen")
	}

	if ok := b.Iteration.HasBuild("happycloud.image"); !ok {
		t.Errorf("expected a basic build entry to be created but it didn't")
	}

}

func TestInitialize_ExistingBucketExistingIteration(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateIteration(context.TODO())
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

	err = b.populateIteration(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Iteration.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}
}

func TestInitialize_ExistingBucketCompleteIteration(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true
	mockService.IterationCompleted = true
	mockService.BuildAlreadyDone = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err == nil {
		t.Errorf("unexpected failure: %v", err)
	}

	if mockService.CreateIterationCalled {
		t.Errorf("unexpected call to CreateIteration")
	}

	if !mockService.GetIterationCalled {
		t.Errorf("expected a call to GetIteration but it didn't happen")
	}

	if mockService.CreateBuildCalled {
		t.Errorf("unexpected call to CreateBuild")
	}

	if b.Iteration.ID != "iteration-id" {
		t.Errorf("expected an iteration to be returned but it wasn't")
	}
}

func TestUpdateBuildStatus(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateIteration(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Iteration.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusUNSET {
		t.Errorf("expected the existing build to be in the default state")
	}

	err = b.UpdateBuildStatus(context.TODO(), "happycloud.image", models.HashicorpCloudPackerBuildStatusRUNNING)
	if err != nil {
		t.Errorf("unexpected failure for PublishBuildStatus: %v", err)
	}

	existingBuild, err = b.Iteration.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}

func TestUpdateBuildStatus_DONENoImages(t *testing.T) {
	mockService := api.NewMockPackerClientService()
	mockService.BucketAlreadyExist = true
	mockService.IterationAlreadyExist = true

	b := &Bucket{
		Slug: "TestBucket",
		client: &api.Client{
			Packer: mockService,
		},
	}

	b.Iteration = NewIteration()
	err := b.Iteration.Initialize()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	b.Iteration.expectedBuilds = append(b.Iteration.expectedBuilds, "happycloud.image")
	mockService.ExistingBuilds = append(mockService.ExistingBuilds, "happycloud.image")

	err = b.Initialize(context.TODO(), models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	err = b.populateIteration(context.TODO())
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	existingBuild, err := b.Iteration.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
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

	existingBuild, err = b.Iteration.Build("happycloud.image")
	if err != nil {
		t.Errorf("expected the existing build loaded from an existing bucket to be valid: %v", err)
	}

	if existingBuild.Status != models.HashicorpCloudPackerBuildStatusRUNNING {
		t.Errorf("expected the existing build to be in the running state")
	}
}

//func (b *Bucket) PublishBuildStatus(ctx context.Context, name string, status models.HashicorpCloudPackerBuildStatus) error {}
