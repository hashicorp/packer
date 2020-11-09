package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/multistep"
)

func testImage() *ec2.Image {
	return &ec2.Image{
		ImageId:         aws.String("ami-abcd1234"),
		CreationDate:    aws.String("ami_test_creation_date"),
		Name:            aws.String("ami_test_name"),
		OwnerId:         aws.String("ami_test_owner_id"),
		ImageOwnerAlias: aws.String("ami_test_owner_alias"),
		RootDeviceType:  aws.String("ebs"),
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

func testGeneratedData(state multistep.StateBag) packerbuilderdata.GeneratedData {
	generatedData := packerbuilderdata.GeneratedData{State: state}
	return generatedData
}

func TestInterpolateBuildInfo_extractBuildInfo_noSourceImage(t *testing.T) {
	state := testState()
	generatedData := testGeneratedData(state)
	buildInfo := extractBuildInfo("foo", state, &generatedData)

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
	generatedData := testGeneratedData(state)
	buildInfo := extractBuildInfo("foo", state, &generatedData)

	expected := BuildInfoTemplate{
		BuildRegion:           "foo",
		SourceAMI:             "ami-abcd1234",
		SourceAMICreationDate: "ami_test_creation_date",
		SourceAMIName:         "ami_test_name",
		SourceAMIOwner:        "ami_test_owner_id",
		SourceAMIOwnerName:    "ami_test_owner_alias",
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
	generatedData := testGeneratedData(state)
	extractBuildInfo("foo", state, &generatedData)

	generatedDataState := state.Get("generated_data").(map[string]interface{})

	if generatedDataState["SourceAMIName"] != "ami_test_name" {
		t.Fatalf("Unexpected state SourceAMIName: expected %#v got %#v\n", "ami_test_name", generatedDataState["SourceAMIName"])
	}
}
