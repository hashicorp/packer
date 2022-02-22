package registry

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func createInitialBucket(t testing.TB) *Bucket {
	oldEnv := os.Getenv("HCP_PACKER_BUILD_FINGERPRINT")
	os.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "no-fingerprint-here")
	defer func() {
		os.Setenv("HCP_PACKER_BUILD_FINGERPRINT", oldEnv)
	}()

	t.Helper()
	subject, err := NewBucketWithIteration(IterationOptions{})
	if err != nil {
		t.Fatalf("failed when calling NewBucketWithIteration: %s", err)
	}

	subject.Slug = "TestBucket"
	subject.BuildLabels = map[string]string{
		"version":   "1.7.0",
		"based_off": "alpine",
	}
	subject.client = &Client{
		Packer: NewMockPackerClientService(),
	}
	return subject
}

func checkError(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		return
	}

	t.Errorf("received an error during testing %s", err)
}

func TestBucket_CreateInitialBuildForIteration(t *testing.T) {
	subject := createInitialBucket(t)

	componentName := "happycloud.image"
	subject.RegisterBuildForComponent(componentName)
	err := subject.CreateInitialBuildForIteration(context.TODO(), componentName)
	checkError(t, err)

	// Assert that a build stored on the iteration
	iBuild, ok := subject.Iteration.builds.Load(componentName)
	if !ok {
		t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
	}

	build, ok := iBuild.(*Build)
	if !ok {
		t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
	}

	if build.ComponentType != componentName {
		t.Errorf("expected the initial build to have the defined component type")
	}

	if diff := cmp.Diff(build.Labels, subject.BuildLabels); diff != "" {
		t.Errorf("expected the initial build to have the defined build labels %v", diff)
	}
}

func TestBucket_UpdateLabelsForBuild(t *testing.T) {
	tc := []struct {
		desc           string
		components     []string
		labels         map[string]string
		labelsCount    int
		noDiffExpected bool
	}{
		{
			desc:           "only global build labels",
			components:     []string{"happcloud.image"},
			labelsCount:    2,
			noDiffExpected: true,
		},
		{
			desc:       "global build labels and one additional build specific label",
			components: []string{"happcloud.image"},
			labels: map[string]string{
				"source_image": "another-happycloud-image",
			},
			labelsCount:    3,
			noDiffExpected: false,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			subject := createInitialBucket(t)

			for _, componentName := range tt.components {
				subject.RegisterBuildForComponent(componentName)
				err := subject.CreateInitialBuildForIteration(context.TODO(), componentName)
				checkError(t, err)

				err = subject.UpdateLabelsForBuild(componentName, tt.labels)
				checkError(t, err)

				// Assert that the build is stored on the iteration
				iBuild, ok := subject.Iteration.builds.Load(componentName)
				if !ok {
					t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
				}

				build, ok := iBuild.(*Build)
				if !ok {
					t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
				}

				if build.ComponentType != componentName {
					t.Errorf("expected the build to have the defined component type")
				}

				if len(build.Labels) != tt.labelsCount {
					t.Errorf("expected the build to have %d build labels but there is only: %d", tt.labelsCount, len(build.Labels))
				}

				diff := cmp.Diff(build.Labels, subject.BuildLabels)
				if (diff == "") != tt.noDiffExpected {
					t.Errorf("expected the build to have an additional build label but there is no diff: %q", diff)
				}

			}
		})
	}
}

func TestBucket_UpdateLabelsForBuild_withMultipleBuilds(t *testing.T) {
	subject := createInitialBucket(t)

	firstComponent := "happycloud.image"
	subject.RegisterBuildForComponent(firstComponent)
	err := subject.CreateInitialBuildForIteration(context.TODO(), firstComponent)
	checkError(t, err)

	err = subject.UpdateLabelsForBuild(firstComponent, map[string]string{
		"source_image": "another-happycloud-image",
	})
	checkError(t, err)

	secondComponent := "happycloud.image2"
	subject.RegisterBuildForComponent(secondComponent)
	err = subject.CreateInitialBuildForIteration(context.TODO(), secondComponent)
	checkError(t, err)

	err = subject.UpdateLabelsForBuild(secondComponent, map[string]string{
		"source_image": "the-original-happycloud-image",
		"role_name":    "no-role-is-a-good-role",
	})
	checkError(t, err)

	var registeredBuilds []*Build
	expectedComponents := []string{firstComponent, secondComponent}
	for _, componentName := range expectedComponents {
		// Assert that a build stored on the iteration
		iBuild, ok := subject.Iteration.builds.Load(componentName)
		if !ok {
			t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
		}

		build, ok := iBuild.(*Build)
		if !ok {
			t.Errorf("expected an initial build for %s to be created, but it failed", componentName)
		}
		registeredBuilds = append(registeredBuilds, build)

		if build.ComponentType != componentName {
			t.Errorf("expected the initial build to have the defined component type")
		}

		if ok := cmp.Equal(build.Labels, subject.BuildLabels); ok {
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
