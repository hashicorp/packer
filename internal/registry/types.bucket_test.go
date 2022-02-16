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

func TestBucket_CreateInitialBuildForIteration(t *testing.T) {
	subject := createInitialBucket(t)

	componentName := "happycloud.image"
	subject.RegisterBuildForComponent(componentName)
	subject.CreateInitialBuildForIteration(context.TODO(), componentName)

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
	subject := createInitialBucket(t)

	componentName := "happycloud.image"
	subject.RegisterBuildForComponent(componentName)
	subject.CreateInitialBuildForIteration(context.TODO(), componentName)
	subject.UpdateLabelsForBuild(componentName, map[string]string{
		"source_image": "another-happycloud-image",
	})

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

	if diff := cmp.Diff(build.Labels, subject.BuildLabels); diff == "" {
		t.Errorf("expected the initial build to have an additional build label but thee is no diff: %q", diff)
	}
}

func TestBucket_UpdateLabelsForBuild_withMultipleBuilds(t *testing.T) {
	subject := createInitialBucket(t)

	firstComponent := "happycloud.image"
	subject.RegisterBuildForComponent(firstComponent)
	subject.CreateInitialBuildForIteration(context.TODO(), firstComponent)
	subject.UpdateLabelsForBuild(firstComponent, map[string]string{
		"source_image": "another-happycloud-image",
	})

	secondComponent := "happycloud.image2"
	subject.RegisterBuildForComponent(secondComponent)
	subject.CreateInitialBuildForIteration(context.TODO(), secondComponent)
	subject.UpdateLabelsForBuild(secondComponent, map[string]string{
		"source_image": "the-original-happycloud-image",
	})

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

		if build.ComponentType != componentName {
			t.Errorf("expected the initial build to have the defined component type")
		}

		t.Logf("Comparing component build labels: %v \n against global build labels: %v", build.Labels, subject.BuildLabels)
		if ok := cmp.Equal(build.Labels, subject.BuildLabels); ok {

			t.Errorf("expected the initial build to have an additional build label but they are equal")
		}
	}
}
