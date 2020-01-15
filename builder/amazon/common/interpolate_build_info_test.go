package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
)

func testImage() *ec2.Image {
	return &ec2.Image{
		ImageId:         aws.String("ami-abcd1234"),
		Name:            aws.String("ami_test_name"),
		OwnerId:         aws.String("ami_test_owner_id"),
		ImageOwnerAlias: aws.String("ami_test_owner_alias"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("key-1"),
				Value: aws.String("value-1"),
			},
			{
				Key:   aws.String("key-2"),
				Value: aws.String("value-2"),
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
		BuildRegion:        "foo",
		SourceAMI:          "ami-abcd1234",
		SourceAMIName:      "ami_test_name",
		SourceAMIOwner:     "ami_test_owner_id",
		SourceAMIOwnerName: "ami_test_owner_alias",
		SourceAMITags: map[string]string{
			"key-1": "value-1",
			"key-2": "value-2",
		},
	}
	if !reflect.DeepEqual(*buildInfo, expected) {
		t.Fatalf("Unexpected BuildInfoTemplate: expected %#v got %#v\n", expected, *buildInfo)
	}
}

func TestInterpolateBuildInfo_extractBuildInfo_GeneratedDataWithSourceImageName(t *testing.T) {
	state := testState()
	state.Put("source_image", testImage())
	extractBuildInfo("foo", state)

	generatedData := state.Get("generated_data").(map[string]interface{})

	if generatedData["SourceAMIName"] != "ami_test_name" {
		t.Fatalf("Unexpected state SourceAMIName: expected %#v got %#v\n", "ami_test_name", generatedData["SourceAMIName"])
	}
}
