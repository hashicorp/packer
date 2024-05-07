// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-openapi/runtime"
	hcpPackerService "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/client/packer_service"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockPackerClientService represents a basic mock of the Cloud Packer Service.
// Upon calling a service method a boolean is set to true to indicate that a method has been called.
// To skip the setting of these booleans set TrackCalledServiceMethods to false; defaults to true in NewMockPackerClientService().
type MockPackerClientService struct {
	CreateBucketCalled, UpdateBucketCalled, BucketAlreadyExist                   bool
	CreateVersionCalled, GetVersionCalled, VersionAlreadyExist, VersionCompleted bool
	CreateBuildCalled, UpdateBuildCalled, ListBuildsCalled, BuildAlreadyDone     bool
	TrackCalledServiceMethods                                                    bool

	// Mock Creates
	CreateBucketResp  *hcpPackerModels.HashicorpCloudPacker20230101CreateBucketResponse
	CreateVersionResp *hcpPackerModels.HashicorpCloudPacker20230101CreateVersionResponse
	CreateBuildResp   *hcpPackerModels.HashicorpCloudPacker20230101CreateBuildResponse

	// Mock Gets
	GetVersionResp *hcpPackerModels.HashicorpCloudPacker20230101GetVersionResponse

	ExistingBuilds      []string
	ExistingBuildLabels map[string]string

	hcpPackerService.ClientService
}

// NewMockPackerClientService returns a basic mock of the Cloud Packer Service.
// Upon calling a service method a boolean is set to true to indicate that a method has been called.
// To skip the setting of these booleans set TrackCalledServiceMethods to false. By default, it is true.
func NewMockPackerClientService() *MockPackerClientService {
	m := MockPackerClientService{
		ExistingBuilds:            make([]string, 0),
		ExistingBuildLabels:       make(map[string]string),
		TrackCalledServiceMethods: true,
	}

	return &m
}

func (svc *MockPackerClientService) PackerServiceCreateBucket(
	params *hcpPackerService.PackerServiceCreateBucketParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceCreateBucketOK, error) {

	if svc.BucketAlreadyExist {
		return nil, status.Error(
			codes.AlreadyExists,
			fmt.Sprintf("Code:%d %s", codes.AlreadyExists, codes.AlreadyExists.String()),
		)
	}

	if params.Body.Name == "" {
		return nil, errors.New("no bucket name was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateBucketCalled = true
	}
	payload := &hcpPackerModels.HashicorpCloudPacker20230101CreateBucketResponse{
		Bucket: &hcpPackerModels.HashicorpCloudPacker20230101Bucket{
			ID: "bucket-id",
		},
	}
	payload.Bucket.Name = params.Body.Name

	ok := &hcpPackerService.PackerServiceCreateBucketOK{
		Payload: payload,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBucket(
	params *hcpPackerService.PackerServiceUpdateBucketParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceUpdateBucketOK, error) {
	if svc.TrackCalledServiceMethods {
		svc.UpdateBucketCalled = true
	}

	return hcpPackerService.NewPackerServiceUpdateBucketOK(), nil
}

func (svc *MockPackerClientService) PackerServiceCreateVersion(
	params *hcpPackerService.PackerServiceCreateVersionParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceCreateVersionOK,
	error) {
	if svc.VersionAlreadyExist {
		return nil, status.Error(
			codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.AlreadyExists,
				codes.AlreadyExists.String()),
		)
	}

	if params.Body.Fingerprint == "" {
		return nil, errors.New("no valid Fingerprint was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateVersionCalled = true
	}
	payload := &hcpPackerModels.HashicorpCloudPacker20230101CreateVersionResponse{
		Version: &hcpPackerModels.HashicorpCloudPacker20230101Version{
			BucketName:   params.BucketName,
			Fingerprint:  params.Body.Fingerprint,
			ID:           "version-id",
			Name:         "v0",
			Status:       hcpPackerModels.HashicorpCloudPacker20230101VersionStatusVERSIONRUNNING.Pointer(),
			TemplateType: params.Body.TemplateType,
		},
	}

	ok := &hcpPackerService.PackerServiceCreateVersionOK{
		Payload: payload,
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceGetVersion(
	params *hcpPackerService.PackerServiceGetVersionParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceGetVersionOK, error) {
	if !svc.VersionAlreadyExist {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Code:%d %s", codes.Aborted, codes.Aborted.String()))
	}

	if params.BucketName == "" {
		return nil, errors.New("no valid BucketName was passed in")
	}

	if params.Fingerprint == "" {
		return nil, errors.New("no valid Fingerprint was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.GetVersionCalled = true
	}

	payload := &hcpPackerModels.HashicorpCloudPacker20230101GetVersionResponse{
		Version: &hcpPackerModels.HashicorpCloudPacker20230101Version{
			ID:           "version-id",
			Builds:       make([]*hcpPackerModels.HashicorpCloudPacker20230101Build, 0),
			TemplateType: hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeTEMPLATETYPEUNSET.Pointer(),
		},
	}

	payload.Version.BucketName = params.BucketName
	payload.Version.Fingerprint = params.Fingerprint
	ok := &hcpPackerService.PackerServiceGetVersionOK{
		Payload: payload,
	}

	if svc.VersionCompleted {
		ok.Payload.Version.Name = "v1"
		ok.Payload.Version.Builds = append(ok.Payload.Version.Builds, &hcpPackerModels.HashicorpCloudPacker20230101Build{
			ID:            "build-id",
			ComponentType: svc.ExistingBuilds[0],
			Status:        hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE.Pointer(),
			Artifacts: []*hcpPackerModels.HashicorpCloudPacker20230101Artifact{
				{ExternalIdentifier: "image-id", Region: "somewhere"},
			},
			Labels: make(map[string]string),
		})
	} else {
		ok.Payload.Version.Name = "v0"
	}

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceCreateBuild(
	params *hcpPackerService.PackerServiceCreateBuildParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceCreateBuildOK, error) {
	if params.BucketName == "" {
		return nil, errors.New("no valid BucketName was passed in")
	}

	if params.Fingerprint == "" {
		return nil, errors.New("no valid Fingerprint was passed in")
	}

	if params.Body.ComponentType == "" {
		return nil, errors.New("no build componentType was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.CreateBuildCalled = true
	}

	payload := &hcpPackerModels.HashicorpCloudPacker20230101CreateBuildResponse{
		Build: &hcpPackerModels.HashicorpCloudPacker20230101Build{
			PackerRunUUID: "test-uuid",
			Status:        hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET.Pointer(),
		},
	}

	payload.Build.ComponentType = params.Body.ComponentType

	ok := hcpPackerService.NewPackerServiceCreateBuildOK()
	ok.Payload = payload

	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceUpdateBuild(
	params *hcpPackerService.PackerServiceUpdateBuildParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceUpdateBuildOK, error) {
	if params.BuildID == "" {
		return nil, errors.New("no valid BuildID was passed in")
	}

	if params.Body == nil {
		return nil, errors.New("no valid Updates were passed in")
	}

	if params.Body.Status == nil || *params.Body.Status == "" {
		return nil, errors.New("no build status was passed in")
	}

	if svc.TrackCalledServiceMethods {
		svc.UpdateBuildCalled = true
	}

	ok := hcpPackerService.NewPackerServiceUpdateBuildOK()
	ok.Payload = &hcpPackerModels.HashicorpCloudPacker20230101UpdateBuildResponse{
		Build: &hcpPackerModels.HashicorpCloudPacker20230101Build{
			ID: params.BuildID,
		},
	}
	return ok, nil
}

func (svc *MockPackerClientService) PackerServiceListBuilds(
	params *hcpPackerService.PackerServiceListBuildsParams, _ runtime.ClientAuthInfoWriter,
	opts ...hcpPackerService.ClientOption,
) (*hcpPackerService.PackerServiceListBuildsOK, error) {

	status := hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDUNSET
	artifacts := make([]*hcpPackerModels.HashicorpCloudPacker20230101Artifact, 0)
	labels := make(map[string]string)
	if svc.BuildAlreadyDone {
		status = hcpPackerModels.HashicorpCloudPacker20230101BuildStatusBUILDDONE
		artifacts = append(artifacts, &hcpPackerModels.HashicorpCloudPacker20230101Artifact{ExternalIdentifier: "image-id", Region: "somewhere"})
	}

	for k, v := range svc.ExistingBuildLabels {
		labels[k] = v
	}

	builds := make([]*hcpPackerModels.HashicorpCloudPacker20230101Build, 0, len(svc.ExistingBuilds))
	for i, name := range svc.ExistingBuilds {
		builds = append(builds, &hcpPackerModels.HashicorpCloudPacker20230101Build{
			ID:            name + "--" + strconv.Itoa(i),
			ComponentType: name,
			Platform:      "mockPlatform",
			Status:        &status,
			Artifacts:     artifacts,
			Labels:        labels,
		})
	}

	ok := hcpPackerService.NewPackerServiceListBuildsOK()
	ok.Payload = &hcpPackerModels.HashicorpCloudPacker20230101ListBuildsResponse{
		Builds: builds,
	}

	return ok, nil
}
