package common

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/outscale/osc-go/oapi"
)

func testImage() oapi.Image {
	return oapi.Image{
		ImageId:   "ami-abcd1234",
		ImageName: "ami_test_name",
		Tags: []oapi.ResourceTag{
			{
				Key:   "key-1",
				Value: "value-1",
			},
			{
				Key:   "key-2",
				Value: "value-2",
			},
		},
	}
}

func testState() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	return state
}

func TestInterpolateBuildInfo_extractBuildInfo_noSourceImage(t *testing.T) {
	state := testState()
	buildInfo := extractBuildInfo("foo", state)

	expected := BuildInfoTemplate{
		BuildRegion: "foo",
	}
	if !reflect.DeepEqual(*buildInfo, expected) {
		t.Fatalf("Unexpected BuildInfoTemplate: expected %#v got %#v\n", expected, *buildInfo)
	}
}

func TestInterpolateBuildInfo_extractBuildInfo_withSourceImage(t *testing.T) {
	state := testState()
	state.Put("source_image", testImage())
	buildInfo := extractBuildInfo("foo", state)

	expected := BuildInfoTemplate{
		BuildRegion:   "foo",
		SourceOMI:     "ami-abcd1234",
		SourceOMIName: "ami_test_name",
		SourceOMITags: map[string]string{
			"key-1": "value-1",
			"key-2": "value-2",
		},
	}
	if !reflect.DeepEqual(*buildInfo, expected) {
		t.Fatalf("Unexpected BuildInfoTemplate: expected %#v got %#v\n", expected, *buildInfo)
	}
}
