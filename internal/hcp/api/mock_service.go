package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockPackerClientService represents a basic mock of the Cloud Packer Service.
// Upon calling a service method a boolean is set to true to indicate that a method has been called.
// To skip the setting of these booleans set TrackCalledServiceMethods to false; defaults to true in NewMockPackerClientService().
type MockPackerClientService struct {
	CreateBucketCalled, UpdateBucketCalled, BucketAlreadyExist                           bool
	CreateIterationCalled, GetIterationCalled, IterationAlreadyExist, IterationCompleted bool
	CreateBuildCalled, UpdateBuildCalled, ListBuildsCalled, BuildAlreadyDone             bool
	TrackCalledServiceMethods                                                            bool

	// Mock Creates
	CreateBucketResp    *models.HashicorpCloudPackerCreateBucketResponse
	CreateIterationResp *models.HashicorpCloudPackerCreateIterationResponse
	CreateBuildResp     *models.HashicorpCloudPackerCreateBuildResponse

	// Mock Gets
	GetIterationResp *models.HashicorpCloudPackerGetIterationResponse

	ExistingBuilds      []string
	ExistingBuildLabels map[string]string

	packerSvc.ClientService
}

// NewMockPackerClientService returns a basic mock of the Cloud Packer Service.
// Upon calling a service method a boolean is set to true to indicate that a method has been called.
// To skip the setting of these booleans set TrackCalledServiceMethods to false. By default it is true.
func NewMockPackerClientService() *MockPackerClientService {
	m := MockPackerClientService{
		ExistingBuilds:            make([]string, 0),
		ExistingBuildLabels:       make(map[string]string),
		TrackCalledServiceMethods: true,
	}

	return &m
}

func (svc *MockPackerClientService) PackerServiceCreateBucket(params *packerSvc.PackerServiceCreateBucketParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceCreateBucketOK, error) {

	if svc.BucketAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()))
	}

	if params.Body.BucketSlug == "" {
		return nil, errors.New("No bucket slug was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateBucketCalled = true
	}
	payload := &models.HashicorpCloudPackerCreateBucketResponse{
		Bucket: &models.HashicorpCloudPackerBucket{
			ID: "bucket-id",
		},
	}
	payload.Bucket.Slug = params.Body.BucketSlug

	ok := &packerSvc.PackerServiceCreateBucketOK{
		Payload: payload,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBucket(params *packerSvc.PackerServiceUpdateBucketParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceUpdateBucketOK, error) {
	if svc.TrackCalledServiceMethods {
		svc.UpdateBucketCalled = true
	}

	return packerSvc.NewPackerServiceUpdateBucketOK(), nil
}

func (svc *MockPackerClientService) PackerServiceCreateIteration(params *packerSvc.PackerServiceCreateIterationParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceCreateIterationOK, error) {
	if svc.IterationAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()))
	}

	if params.Body.Fingerprint == "" {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateIterationCalled = true
	}
	payload := &models.HashicorpCloudPackerCreateIterationResponse{
		Iteration: &models.HashicorpCloudPackerIteration{
			ID:           "iteration-id",
			TemplateType: params.Body.TemplateType,
		},
	}

	payload.Iteration.BucketSlug = params.BucketSlug
	payload.Iteration.Fingerprint = params.Body.Fingerprint

	ok := &packerSvc.PackerServiceCreateIterationOK{
		Payload: payload,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceGetIteration(params *packerSvc.PackerServiceGetIterationParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceGetIterationOK, error) {
	if !svc.IterationAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.Aborted, codes.Aborted.String()))
	}

	if params.BucketSlug == "" {
		return nil, errors.New("No valid BucketSlug was passed in")
	}

	if params.Fingerprint == nil {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.GetIterationCalled = true
	}

	payload := &models.HashicorpCloudPackerGetIterationResponse{
		Iteration: &models.HashicorpCloudPackerIteration{
			ID:           "iteration-id",
			Builds:       make([]*models.HashicorpCloudPackerBuild, 0),
			TemplateType: models.HashicorpCloudPackerIterationTemplateTypeTEMPLATETYPEUNSET.Pointer(),
		},
	}

	payload.Iteration.BucketSlug = params.BucketSlug
	payload.Iteration.Fingerprint = *params.Fingerprint
	ok := &packerSvc.PackerServiceGetIterationOK{
		Payload: payload,
	}

	if svc.IterationCompleted {
		ok.Payload.Iteration.Complete = true
		ok.Payload.Iteration.IncrementalVersion = 1
		ok.Payload.Iteration.Builds = append(ok.Payload.Iteration.Builds, &models.HashicorpCloudPackerBuild{
			ID:            "build-id",
			ComponentType: svc.ExistingBuilds[0],
			Status:        models.HashicorpCloudPackerBuildStatusDONE.Pointer(),
			Images: []*models.HashicorpCloudPackerImage{
				{ImageID: "image-id", Region: "somewhere"},
			},
			Labels: make(map[string]string),
		})
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceCreateBuild(params *packerSvc.PackerServiceCreateBuildParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceCreateBuildOK, error) {
	if params.BucketSlug == "" {
		return nil, errors.New("No valid BucketSlug was passed in")
	}

	if params.Body.Fingerprint == "" {
		return nil, errors.New("No valid Fingerprint was passed in")
	}

	if params.Body.Build.ComponentType == "" {
		return nil, errors.New("No build componentType was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateBuildCalled = true
	}

	payload := &models.HashicorpCloudPackerCreateBuildResponse{
		Build: &models.HashicorpCloudPackerBuild{
			PackerRunUUID: "test-uuid",
			Status:        models.HashicorpCloudPackerBuildStatusUNSET.Pointer(),
		},
	}

	payload.Build.ComponentType = params.Body.Build.ComponentType
	payload.Build.IterationID = params.IterationID

	ok := packerSvc.NewPackerServiceCreateBuildOK()
	ok.Payload = payload

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBuild(params *packerSvc.PackerServiceUpdateBuildParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceUpdateBuildOK, error) {
	if params.BuildID == "" {
		return nil, errors.New("No valid BuildID was passed in")
	}

	if params.Body.Updates == nil {
		return nil, errors.New("No valid Updates were passed in")
	}

	if params.Body.Updates.Status == nil || *params.Body.Updates.Status == "" {
		return nil, errors.New("No build status was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.UpdateBuildCalled = true
	}

	ok := packerSvc.NewPackerServiceUpdateBuildOK()
	ok.Payload = &models.HashicorpCloudPackerUpdateBuildResponse{
		Build: &models.HashicorpCloudPackerBuild{
			ID: params.BuildID,
		},
	}
	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceListBuilds(params *packerSvc.PackerServiceListBuildsParams, _ runtime.ClientAuthInfoWriter, opts ...packer_service.ClientOption) (*packerSvc.PackerServiceListBuildsOK, error) {

	status := models.HashicorpCloudPackerBuildStatusUNSET
	images := make([]*models.HashicorpCloudPackerImage, 0)
	labels := make(map[string]string)
	if svc.BuildAlreadyDone {
		status = models.HashicorpCloudPackerBuildStatusDONE
		images = append(images, &models.HashicorpCloudPackerImage{ImageID: "image-id", Region: "somewhere"})
	}

	for k, v := range svc.ExistingBuildLabels {
		labels[k] = v
	}

	builds := make([]*models.HashicorpCloudPackerBuild, 0, len(svc.ExistingBuilds))
	for i, name := range svc.ExistingBuilds {
		builds = append(builds, &models.HashicorpCloudPackerBuild{
			ID:            name + "--" + strconv.Itoa(i),
			ComponentType: name,
			CloudProvider: "mockProvider",
			Status:        status.Pointer(),
			Images:        images,
			Labels:        labels,
		})
	}

	ok := packerSvc.NewPackerServiceListBuildsOK()
	ok.Payload = &models.HashicorpCloudPackerListBuildsResponse{
		Builds: builds,
	}

	return ok, nil
}
