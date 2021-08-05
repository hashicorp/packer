package packer_registry

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-openapi/runtime"
	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockPackerClientService struct {
	CreateBucketCalled, UpdateBucketCalled, BucketAlreadyExist                           bool
	CreateIterationCalled, GetIterationCalled, IterationAlreadyExist, IterationCompleted bool
	CreateBuildCalled, UpdateBuildCalled, ListBuildsCalled, BuildAlreadyDone             bool

	// Mock Creates
	CreateBucketResp    *models.HashicorpCloudPackerCreateBucketResponse
	CreateIterationResp *models.HashicorpCloudPackerCreateIterationResponse
	CreateBuildResp     *models.HashicorpCloudPackerCreateBuildResponse

	// Mock Gets
	GetIterationResp *models.HashicorpCloudPackerGetIterationResponse

	ExistingBuilds []string

	packerSvc.ClientService
}

func NewMockPackerClientService() *MockPackerClientService {
	m := MockPackerClientService{
		ExistingBuilds: make([]string, 0),
	}

	m.CreateBucketResp = &models.HashicorpCloudPackerCreateBucketResponse{
		Bucket: &models.HashicorpCloudPackerBucket{
			ID: "bucket-id",
		},
	}

	m.CreateIterationResp = &models.HashicorpCloudPackerCreateIterationResponse{
		Iteration: &models.HashicorpCloudPackerIteration{
			ID: "iteration-id",
		},
	}

	m.CreateBuildResp = &models.HashicorpCloudPackerCreateBuildResponse{
		Build: &models.HashicorpCloudPackerBuild{
			PackerRunUUID: "test-uuid",
			Status:        models.HashicorpCloudPackerBuildStatusUNSET,
		},
	}

	m.GetIterationResp = &models.HashicorpCloudPackerGetIterationResponse{
		Iteration: &models.HashicorpCloudPackerIteration{
			ID:     "iteration-id",
			Builds: make([]*models.HashicorpCloudPackerBuild, 0),
		},
	}

	return &m
}

func (svc *MockPackerClientService) CreateBucket(params *packerSvc.CreateBucketParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.CreateBucketOK, error) {
	if svc.BucketAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()))
	}

	if params.Body.BucketSlug == "" {
		return nil, errors.New("No bucket slug was passed in")
	}

	svc.CreateBucketCalled = true
	svc.CreateBucketResp.Bucket.Slug = params.Body.BucketSlug

	ok := &packerSvc.CreateBucketOK{
		Payload: svc.CreateBucketResp,
	}

	return ok, nil
}

func (svc *MockPackerClientService) UpdateBucket(params *packerSvc.UpdateBucketParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.UpdateBucketOK, error) {
	svc.UpdateBucketCalled = true

	return packerSvc.NewUpdateBucketOK(), nil
}

func (svc *MockPackerClientService) CreateIteration(params *packerSvc.CreateIterationParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.CreateIterationOK, error) {
	if svc.IterationAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()))
	}

	if params.Body.BucketSlug == "" {
		return nil, errors.New("No valid BucketSlug was passed in")
	}

	if params.Body.Fingerprint == "" {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	svc.CreateIterationCalled = true
	svc.CreateIterationResp.Iteration.BucketSlug = params.Body.BucketSlug
	svc.CreateIterationResp.Iteration.Fingerprint = params.Body.Fingerprint

	ok := &packerSvc.CreateIterationOK{
		Payload: svc.CreateIterationResp,
	}

	return ok, nil
}

func (svc *MockPackerClientService) GetIteration(params *packerSvc.GetIterationParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.GetIterationOK, error) {
	if !svc.IterationAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.Aborted, codes.Aborted.String()))
	}

	if params.BucketSlug == "" {
		return nil, errors.New("No valid BucketSlug was passed in")
	}

	if params.Fingerprint == nil {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	svc.GetIterationCalled = true

	//
	ok := &packerSvc.GetIterationOK{
		Payload: svc.GetIterationResp,
	}

	if svc.IterationCompleted {
		ok.Payload.Iteration.Complete = true
		ok.Payload.Iteration.IncrementalVersion = 1
		ok.Payload.Iteration.Builds = append(ok.Payload.Iteration.Builds, &models.HashicorpCloudPackerBuild{
			ID:            "build-id",
			ComponentType: svc.ExistingBuilds[0],
			Status:        models.HashicorpCloudPackerBuildStatusDONE,
			Images: []*models.HashicorpCloudPackerImage{
				{ImageID: "image-id", Region: "somewhere"},
			},
		})
	}

	return ok, nil
}

func (svc *MockPackerClientService) CreateBuild(params *packerSvc.CreateBuildParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.CreateBuildOK, error) {
	if params.Body.BucketSlug == "" {
		return nil, errors.New("No valid BucketSlug was passed in")
	}

	if params.Body.Fingerprint == "" {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	if params.Body.Build.ComponentType == "" {
		return nil, errors.New("No build componentType was passed in")
	}

	svc.CreateBuildCalled = true

	svc.CreateBuildResp.Build.ComponentType = params.Body.Build.ComponentType
	svc.CreateBuildResp.Build.IterationID = params.BuildIterationID

	ok := packerSvc.NewCreateBuildOK()
	ok.Payload = svc.CreateBuildResp

	return ok, nil
}

func (svc *MockPackerClientService) UpdateBuild(params *packerSvc.UpdateBuildParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.UpdateBuildOK, error) {
	if params.Body.BuildID == "" {
		return nil, errors.New("No valid BuildID was passed in")
	}

	if params.Body.Updates == nil {
		return nil, errors.New("No valid Updates were passed in")
	}

	if params.Body.Updates.Status == "" {
		return nil, errors.New("No build status was passed in")
	}

	svc.UpdateBuildCalled = true
	ok := packerSvc.NewUpdateBuildOK()
	ok.Payload = &models.HashicorpCloudPackerUpdateBuildResponse{
		Build: &models.HashicorpCloudPackerBuild{
			ID: params.Body.BuildID,
		},
	}
	return ok, nil
}

func (svc *MockPackerClientService) ListBuilds(params *packerSvc.ListBuildsParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.ListBuildsOK, error) {

	status := models.HashicorpCloudPackerBuildStatusUNSET
	images := make([]*models.HashicorpCloudPackerImage, 0)
	if svc.BuildAlreadyDone {
		status = models.HashicorpCloudPackerBuildStatusDONE
		images = append(images, &models.HashicorpCloudPackerImage{ID: "image-id", Region: "somewhere"})
	}

	builds := make([]*models.HashicorpCloudPackerBuild, 0, len(svc.ExistingBuilds))
	for i, name := range svc.ExistingBuilds {
		builds = append(builds, &models.HashicorpCloudPackerBuild{
			ID:            name + "--" + strconv.Itoa(i),
			ComponentType: name,
			Status:        status,
			Images:        images,
		})
	}

	ok := packerSvc.NewListBuildsOK()
	ok.Payload = &models.HashicorpCloudPackerListBuildsResponse{
		Builds: builds,
	}

	return ok, nil
}
