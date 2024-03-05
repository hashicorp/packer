// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/hcl2template"
	hcpPackerAPI "github.com/hashicorp/packer/internal/hcp/api"
)

func createInitialTestBucket(t testing.TB) *Bucket {
	t.Helper()
	bucket := NewBucketWithVersion()
	err := bucket.Version.Initialize()
	if err != nil {
		t.Errorf("failed to initialize Bucket: %s", err)
		return nil
	}

	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.TrackCalledServiceMethods = false
	bucket.Name = "TestBucket"
	bucket.client = &hcpPackerAPI.Client{
		Packer: mockService,
	}

	return bucket
}

func checkError(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Errorf("received an error during testing %s", err)
}

func TestBucket_CreateInitialBuildForVersion(t *testing.T) {
	bucket := createInitialTestBucket(t)

	componentName := "happycloud.artifact"
	bucket.RegisterBuildForComponent(componentName)
	bucket.BuildLabels = map[string]string{
		"version":   "1.7.0",
		"based_off": "alpine",
	}
	err := bucket.CreateInitialBuildForVersion(context.TODO(), componentName)
	checkError(t, err)

	// Assert that a build stored on the version
	build, err := bucket.Version.Build(componentName)
	if err != nil {
		t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
	}

	if build.ComponentType != componentName {
		t.Errorf("expected the initial build to have the defined component type")
	}

	if diff := cmp.Diff(build.Labels, bucket.BuildLabels); diff != "" {
		t.Errorf("expected the initial build to have the defined build labels %v", diff)
	}
}

func TestBucket_UpdateLabelsForBuild(t *testing.T) {
	tc := []struct {
		desc              string
		buildName         string
		bucketBuildLabels map[string]string
		buildLabels       map[string]string
		labelsCount       int
		noDiffExpected    bool
	}{
		{
			desc:           "no bucket or build specific labels",
			buildName:      "happcloud.artifact",
			noDiffExpected: true,
		},
		{
			desc:      "bucket build labels",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    2,
			noDiffExpected: true,
		},
		{
			desc:      "bucket build labels and build specific label",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			buildLabels: map[string]string{
				"source_artifact": "another-happycloud-artifact",
			},
			labelsCount:    3,
			noDiffExpected: false,
		},
		{
			desc:      "build specific label",
			buildName: "happcloud.artifact",
			buildLabels: map[string]string{
				"source_artifact": "another-happycloud-artifact",
			},
			labelsCount:    1,
			noDiffExpected: false,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			bucket := createInitialTestBucket(t)

			componentName := tt.buildName
			bucket.RegisterBuildForComponent(componentName)

			for k, v := range tt.bucketBuildLabels {
				bucket.BuildLabels[k] = v
			}

			err := bucket.CreateInitialBuildForVersion(context.TODO(), componentName)
			checkError(t, err)

			// Assert that the build is stored on the version
			build, err := bucket.Version.Build(componentName)
			if err != nil {
				t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
			}

			if build.ComponentType != componentName {
				t.Errorf("expected the build to have the defined component type")
			}

			err = bucket.UpdateLabelsForBuild(componentName, tt.buildLabels)
			checkError(t, err)

			if len(build.Labels) != tt.labelsCount {
				t.Errorf("expected the build to have %d build labels but there is only: %d", tt.labelsCount, len(build.Labels))
			}

			diff := cmp.Diff(build.Labels, bucket.BuildLabels)
			if (diff == "") != tt.noDiffExpected {
				t.Errorf("expected the build to have an additional build label but there is no diff: %q", diff)
			}

		})
	}
}

func TestBucket_UpdateLabelsForBuild_withMultipleBuilds(t *testing.T) {
	bucket := createInitialTestBucket(t)

	firstComponent := "happycloud.artifact"
	bucket.RegisterBuildForComponent(firstComponent)

	secondComponent := "happycloud.artifact2"
	bucket.RegisterBuildForComponent(secondComponent)

	err := bucket.populateVersion(context.TODO())
	checkError(t, err)

	err = bucket.UpdateLabelsForBuild(firstComponent, map[string]string{
		"source_artifact": "another-happycloud-artifact",
	})
	checkError(t, err)

	err = bucket.UpdateLabelsForBuild(secondComponent, map[string]string{
		"source_artifact": "the-original-happycloud-artifact",
		"role_name":       "no-role-is-a-good-role",
	})
	checkError(t, err)

	var registeredBuilds []*Build
	expectedComponents := []string{firstComponent, secondComponent}
	for _, componentName := range expectedComponents {
		// Assert that a build stored on the version
		build, err := bucket.Version.Build(componentName)
		if err != nil {
			t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
		}
		registeredBuilds = append(registeredBuilds, build)

		if build.ComponentType != componentName {
			t.Errorf("expected the initial build to have the defined component type")
		}

		if ok := cmp.Equal(build.Labels, bucket.BuildLabels); ok {
			t.Errorf("expected the build to have an additional build label but they are equal")
		}
	}

	if len(registeredBuilds) != 2 {
		t.Errorf("expected the bucket to have 2 registered builds but got %d", len(registeredBuilds))
	}

	if ok := cmp.Equal(registeredBuilds[0].Labels, registeredBuilds[1].Labels); ok {
		t.Errorf("expected registered builds to have different labels but they are equal")
	}
}

func TestBucket_PopulateVersion(t *testing.T) {
	tc := []struct {
		desc              string
		buildName         string
		bucketBuildLabels map[string]string
		buildLabels       map[string]string
		labelsCount       int
		buildCompleted    bool
		noDiffExpected    bool
	}{
		{
			desc:           "populating version with existing incomplete build and no bucket build labels does nothing",
			buildName:      "happcloud.artifact",
			labelsCount:    0,
			buildCompleted: false,
			noDiffExpected: true,
		},
		{
			desc:      "populating version with existing incomplete build should add bucket build labels",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    2,
			buildCompleted: false,
			noDiffExpected: true,
		},
		{
			desc:      "populating version with existing incomplete build should update bucket build labels",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.3",
				"based_off": "alpine-3.14",
			},
			buildLabels: map[string]string{
				"version":   "packer.version",
				"based_off": "var.distro",
			},
			labelsCount:    2,
			buildCompleted: false,
			noDiffExpected: true,
		},
		{
			desc:      "populating version with completed build should not modify any labels",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    0,
			buildCompleted: true,
			noDiffExpected: false,
		},
		{
			desc:      "populating version with existing build should only modify bucket build labels",
			buildName: "happcloud.artifact",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.3",
				"based_off": "alpine-3.14",
			},
			buildLabels: map[string]string{
				"arch": "linux/386",
			},
			labelsCount:    3,
			buildCompleted: false,
			noDiffExpected: false,
		},
	}

	for i, tt := range tc {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {

			t.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "test-run-"+strconv.Itoa(i))

			mockService := hcpPackerAPI.NewMockPackerClientService()
			mockService.BucketAlreadyExist = true
			mockService.VersionAlreadyExist = true
			mockService.BuildAlreadyDone = tt.buildCompleted

			bucket := NewBucketWithVersion()
			err := bucket.Version.Initialize()
			if err != nil {
				t.Fatalf("failed when calling NewBucketWithVersion: %s", err)
			}

			bucket.Name = "TestBucket"
			bucket.client = &hcpPackerAPI.Client{
				Packer: mockService,
			}
			for k, v := range tt.bucketBuildLabels {
				bucket.BuildLabels[k] = v
			}

			componentName := "happycloud.artifact"
			bucket.RegisterBuildForComponent(componentName)

			mockService.ExistingBuilds = append(mockService.ExistingBuilds, componentName)
			mockService.ExistingBuildLabels = tt.buildLabels

			err = bucket.populateVersion(context.TODO())
			checkError(t, err)

			if mockService.CreateBuildCalled {
				t.Errorf("expected an initial build for %s to already exist, but it called CreateBuild", componentName)
			}
			// Assert that a build stored on the version
			build, err := bucket.Version.Build(componentName)
			if err != nil {
				t.Errorf("expected an existing build for %s to be stored, but it failed", componentName)
			}

			if build.ComponentType != componentName {
				t.Errorf("expected the build to have the defined component type")
			}

			if len(build.Labels) != tt.labelsCount {
				t.Errorf("expected the build to have %d build labels but there is only: %d", tt.labelsCount, len(build.Labels))
			}

			diff := cmp.Diff(build.Labels, bucket.BuildLabels)
			if (diff == "") != tt.noDiffExpected {
				t.Errorf("expected the build to have bucket build labels but there is no diff: %q", diff)
			}
		})
	}
}

func TestReadFromHCLBuildBlock(t *testing.T) {
	tc := []struct {
		desc           string
		buildBlock     *hcl2template.BuildBlock
		expectedBucket *Bucket
	}{
		{
			desc: "configure bucket using only hcp_packer_registry block",
			buildBlock: &hcl2template.BuildBlock{
				HCPPackerRegistry: &hcl2template.HCPPackerRegistryBlock{
					Slug:        "hcp_packer_registry-block-test",
					Description: "description from hcp_packer_registry block",
					BucketLabels: map[string]string{
						"org": "test",
					},
					BuildLabels: map[string]string{
						"version":   "1.7.0",
						"based_off": "alpine",
					},
				},
			},
			expectedBucket: &Bucket{
				Name:        "hcp_packer_registry-block-test",
				Description: "description from hcp_packer_registry block",
				BucketLabels: map[string]string{
					"org": "test",
				},
				BuildLabels: map[string]string{
					"version":   "1.7.0",
					"based_off": "alpine",
				},
			},
		},
	}
	for _, tt := range tc {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			bucket := &Bucket{}
			bucket.ReadFromHCLBuildBlock(tt.buildBlock)

			diff := cmp.Diff(bucket, tt.expectedBucket, cmp.AllowUnexported(Bucket{}))
			if diff != "" {
				t.Errorf("expected the build to to have contents of hcp_packer_registry block but it does not: %v", diff)
			}
		})
	}
}

func TestCompleteBuild(t *testing.T) {
	hcpArtifact := &packer.MockArtifact{
		BuilderIdValue: "builder.test",
		FilesValue:     []string{"file.one"},
		IdValue:        "Test",
		StateValues: map[string]interface{}{
			"builder.test": "OK",
			image.ArtifactStateURI: &image.Image{
				ImageID:        "hcp-test",
				ProviderName:   "none",
				ProviderRegion: "none",
				Labels:         map[string]string{},
				SourceImageID:  "",
			},
		},
		DestroyCalled: false,
		StringValue:   "",
	}
	nonHCPArtifact := &packer.MockArtifact{
		BuilderIdValue: "builder.test",
		FilesValue:     []string{"file.one"},
		IdValue:        "Test",
		StateValues: map[string]interface{}{
			"builder.test": "OK",
		},
		DestroyCalled: false,
		StringValue:   "",
	}

	testCases := []struct {
		name           string
		artifactsToUse []packer.Artifact
		expectError    bool
		wantNotHCPErr  bool
	}{
		{
			"OK - one artifact compatible with HCP",
			[]packer.Artifact{
				hcpArtifact,
			},
			false, false,
		},
		{
			"Fail - no artifacts",
			[]packer.Artifact{},
			true, false,
		},
		{
			"Fail - only non HCP compatible artifacts",
			[]packer.Artifact{
				nonHCPArtifact,
			},
			true, true,
		},
		{
			"OK - one hcp artifact, one non hcp artifact (order matters)",
			[]packer.Artifact{
				hcpArtifact,
				nonHCPArtifact,
			},
			false, false,
		},
		{
			"OK - one non hcp artifact, one hcp artifact (order matters)",
			[]packer.Artifact{
				nonHCPArtifact,
				hcpArtifact,
			},
			false, false,
		},
	}
	mockCli := &hcpPackerAPI.Client{
		Packer: hcpPackerAPI.NewMockPackerClientService(),
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dummyBucket := &Bucket{
				Name:        "test-bucket",
				Description: "test",
				Destination: "none",
				RunningBuilds: map[string]chan struct{}{
					// Need buffer with 1 cap so we can signal end of
					// heartbeats in test, otherwise it'll block
					"test-build": make(chan struct{}, 1),
				},
				Version: &Version{
					ID:          "noneID",
					Fingerprint: "TestFingerprint",
					RunUUID:     "testuuid",
					builds:      sync.Map{},
				},
				client: mockCli,
			}

			dummyBucket.Version.StoreBuild("test-build", &Build{
				ID:            "test-build",
				Platform:      "none",
				ComponentType: "none",
				RunUUID:       "testuuid",
				Artifacts:     make(map[string]image.Image),
				Status:        models.HashicorpCloudPacker20230101BuildStatusBUILDRUNNING,
			})

			_, err := dummyBucket.completeBuild(context.Background(), "test-build", tt.artifactsToUse, nil)
			if err != nil != tt.expectError {
				t.Errorf("expected %t error; got %t", tt.expectError, err != nil)
				t.Logf("error was: %s", err)
			}

			if err != nil && tt.wantNotHCPErr {
				_, ok := err.(*NotAHCPArtifactError)
				if !ok {
					t.Errorf("expected a NotAHCPArtifactError, got a %q", reflect.TypeOf(err).String())
				}
			}
		})
	}
}
