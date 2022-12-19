package registry

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer/internal/hcp/api"
)

func createInitialTestBucket(t testing.TB) *Bucket {
	t.Helper()
	bucket := NewBucketWithIteration()
	err := bucket.Iteration.Initialize()
	if err != nil {
		t.Errorf("failed to initialize Bucket: %s", err)
		return nil
	}

	mockService := api.NewMockPackerClientService()
	mockService.TrackCalledServiceMethods = false
	bucket.Slug = "TestBucket"
	bucket.client = &api.Client{
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

func TestBucket_CreateInitialBuildForIteration(t *testing.T) {
	bucket := createInitialTestBucket(t)

	componentName := "happycloud.image"
	bucket.RegisterBuildForComponent(componentName)
	bucket.BuildLabels = map[string]string{
		"version":   "1.7.0",
		"based_off": "alpine",
	}
	err := bucket.CreateInitialBuildForIteration(context.TODO(), componentName)
	checkError(t, err)

	// Assert that a build stored on the iteration
	build, err := bucket.Iteration.Build(componentName)
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
			buildName:      "happcloud.image",
			noDiffExpected: true,
		},
		{
			desc:      "bucket build labels",
			buildName: "happcloud.image",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    2,
			noDiffExpected: true,
		},
		{
			desc:      "bucket build labels and build specific label",
			buildName: "happcloud.image",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			buildLabels: map[string]string{
				"source_image": "another-happycloud-image",
			},
			labelsCount:    3,
			noDiffExpected: false,
		},
		{
			desc:      "build specific label",
			buildName: "happcloud.image",
			buildLabels: map[string]string{
				"source_image": "another-happycloud-image",
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

			err := bucket.CreateInitialBuildForIteration(context.TODO(), componentName)
			checkError(t, err)

			// Assert that the build is stored on the iteration
			build, err := bucket.Iteration.Build(componentName)
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

	firstComponent := "happycloud.image"
	bucket.RegisterBuildForComponent(firstComponent)

	secondComponent := "happycloud.image2"
	bucket.RegisterBuildForComponent(secondComponent)

	err := bucket.populateIteration(context.TODO())
	checkError(t, err)

	err = bucket.UpdateLabelsForBuild(firstComponent, map[string]string{
		"source_image": "another-happycloud-image",
	})
	checkError(t, err)

	err = bucket.UpdateLabelsForBuild(secondComponent, map[string]string{
		"source_image": "the-original-happycloud-image",
		"role_name":    "no-role-is-a-good-role",
	})
	checkError(t, err)

	var registeredBuilds []*Build
	expectedComponents := []string{firstComponent, secondComponent}
	for _, componentName := range expectedComponents {
		// Assert that a build stored on the iteration
		build, err := bucket.Iteration.Build(componentName)
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

func TestBucket_PopulateIteration(t *testing.T) {
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
			desc:           "populating iteration with existing incomplete build and no bucket build labels does nothing",
			buildName:      "happcloud.image",
			labelsCount:    0,
			buildCompleted: false,
			noDiffExpected: true,
		},
		{
			desc:      "populating iteration with existing incomplete build should add bucket build labels",
			buildName: "happcloud.image",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    2,
			buildCompleted: false,
			noDiffExpected: true,
		},
		{
			desc:      "populating iteration with existing incomplete build should update bucket build labels",
			buildName: "happcloud.image",
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
			desc:      "populating iteration with completed build should not modify any labels",
			buildName: "happcloud.image",
			bucketBuildLabels: map[string]string{
				"version":   "1.7.0",
				"based_off": "alpine",
			},
			labelsCount:    0,
			buildCompleted: true,
			noDiffExpected: false,
		},
		{
			desc:      "populating iteration with existing build should only modify bucket build labels",
			buildName: "happcloud.image",
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

			mockService := api.NewMockPackerClientService()
			mockService.BucketAlreadyExist = true
			mockService.IterationAlreadyExist = true
			mockService.BuildAlreadyDone = tt.buildCompleted

			bucket := NewBucketWithIteration()
			err := bucket.Iteration.Initialize()
			if err != nil {
				t.Fatalf("failed when calling NewBucketWithIteration: %s", err)
			}

			bucket.Slug = "TestBucket"
			bucket.client = &api.Client{
				Packer: mockService,
			}
			for k, v := range tt.bucketBuildLabels {
				bucket.BuildLabels[k] = v
			}

			componentName := "happycloud.image"
			bucket.RegisterBuildForComponent(componentName)

			mockService.ExistingBuilds = append(mockService.ExistingBuilds, componentName)
			mockService.ExistingBuildLabels = tt.buildLabels

			err = bucket.populateIteration(context.TODO())
			checkError(t, err)

			if mockService.CreateBuildCalled {
				t.Errorf("expected an initial build for %s to already exist, but it called CreateBuild", componentName)
			}
			// Assert that a build stored on the iteration
			build, err := bucket.Iteration.Build(componentName)
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
