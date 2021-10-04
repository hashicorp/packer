package registry

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

func (svc *MockPackerClientService) PackerServiceCreateBucket(params *packerSvc.PackerServiceCreateBucketParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceCreateBucketOK, error) {
	if svc.BucketAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()))
	}

	if params.Body == nil {
		return nil, errors.New("No body provided.")
	}
	if params.Body.BucketSlug == "" {
		return nil, errors.New("No bucket slug was passed in")
	}

	svc.CreateBucketCalled = true
	// This is set in NewMockPackerClientService()
	svc.CreateBucketResp.Bucket.Slug = params.Body.BucketSlug

	ok := &packerSvc.PackerServiceCreateBucketOK{
		Payload: svc.CreateBucketResp,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBucket(params *packerSvc.PackerServiceUpdateBucketParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceUpdateBucketOK, error) {
	svc.UpdateBucketCalled = true

	return packerSvc.NewPackerServiceUpdateBucketOK(), nil
}

func (svc *MockPackerClientService) PackerServiceCreateIteration(params *packerSvc.PackerServiceCreateIterationParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceCreateIterationOK, error) {
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

	ok := &packerSvc.PackerServiceCreateIterationOK{
		Payload: svc.CreateIterationResp,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceGetIteration(params *packerSvc.PackerServiceGetIterationParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceGetIterationOK, error) {
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
	ok := &packerSvc.PackerServiceGetIterationOK{
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

func (svc *MockPackerClientService) PackerServiceCreateBuild(params *packerSvc.PackerServiceCreateBuildParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceCreateBuildOK, error) {
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
	svc.CreateBuildResp.Build.IterationID = params.IterationID

	ok := packerSvc.NewPackerServiceCreateBuildOK()
	ok.Payload = svc.CreateBuildResp

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBuild(params *packerSvc.PackerServiceUpdateBuildParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceUpdateBuildOK, error) {
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
	ok := packerSvc.NewPackerServiceUpdateBuildOK()
	ok.Payload = &models.HashicorpCloudPackerUpdateBuildResponse{
		Build: &models.HashicorpCloudPackerBuild{
			ID: params.Body.BuildID,
		},
	}
	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceListBuilds(params *packerSvc.PackerServiceListBuildsParams, _ runtime.ClientAuthInfoWriter) (*packerSvc.PackerServiceListBuildsOK, error) {

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

	ok := packerSvc.NewPackerServiceListBuildsOK()
	ok.Payload = &models.HashicorpCloudPackerListBuildsResponse{
		Builds: builds,
	}

	return ok, nil
}
