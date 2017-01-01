package triton

import (
	"testing"
)

func TestTargetImageConfig_Prepare(t *testing.T) {
	tic := testTargetImageConfig(t)
	errs := tic.Prepare(nil)
	if errs != nil {
		t.Fatalf("should not error: %#v", tic)
	}

	tic = testTargetImageConfig(t)
	tic.ImageName = ""
	errs = tic.Prepare(nil)
	if errs == nil {
		t.Fatalf("should error: %#v", tic)
	}

	tic = testTargetImageConfig(t)
	tic.ImageVersion = ""
	errs = tic.Prepare(nil)
	if errs == nil {
		t.Fatalf("should error: %#v", tic)
	}
}

func testTargetImageConfig(t *testing.T) TargetImageConfig {
	return TargetImageConfig{
		ImageName:        "test-image",
		ImageVersion:     "test-version",
		ImageDescription: "test-description",
		ImageHomepage:    "test-homepage",
		ImageEULA:        "test-eula",
		ImageACL: []string{
			"test-acl-1",
			"test-acl-2",
		},
		ImageTags: map[string]string{
			"test-tags-key1": "test-tags-value1",
			"test-tags-key2": "test-tags-value2",
			"test-tags-key3": "test-tags-value3",
		},
	}
}
