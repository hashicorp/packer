package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func testAMIConfig() *AMIConfig {
	return &AMIConfig{
		AMIName: "foo",
	}
}

func getFakeAccessConfig(region string) *AccessConfig {
	c := testAccessConfig()
	c.RawRegion = region
	return c
}

func TestAMIConfigPrepare_name(t *testing.T) {
	c := testAMIConfig()
	accessConf := testAccessConfig()
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.AMIName = ""
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("should have error")
	}
}

type mockEC2Client struct {
	ec2iface.EC2API
}

func (m *mockEC2Client) DescribeRegions(*ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	return &ec2.DescribeRegionsOutput{
		Regions: []*ec2.Region{
			{RegionName: aws.String("us-east-1")},
			{RegionName: aws.String("us-east-2")},
			{RegionName: aws.String("us-west-1")},
		},
	}, nil
}

func TestAMIConfigPrepare_regions(t *testing.T) {
	c := testAMIConfig()
	c.AMIRegions = nil

	var errs []error
	var err error
	accessConf := testAccessConfig()
	mockConn := &mockEC2Client{}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}

	c.AMIRegions, err = listEC2Regions(mockConn)
	if err != nil {
		t.Fatalf("shouldn't have err: %s", err.Error())
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}
	errs = errs[:0]

	c.AMIRegions = []string{"us-east-1", "us-west-1", "us-east-1"}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("bad: %s", errs[0])
	}

	expected := []string{"us-east-1", "us-west-1"}
	if !reflect.DeepEqual(c.AMIRegions, expected) {
		t.Fatalf("bad: %#v", c.AMIRegions)
	}

	c.AMIRegions = []string{"custom"}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("shouldn't have error")
	}

	c.AMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "456-789-0123",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal(fmt.Sprintf("shouldn't have error: %s", errs[0]))
	}

	c.AMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have passed; we are able to use default KMS key if not sharing")
	}

	c.SnapshotUsers = []string{"user-foo", "user-bar"}
	c.AMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have an error b/c can't use default KMS key if sharing")
	}

	c.AMIRegions = []string{"us-east-1", "us-west-1"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "456-789-0123",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have error b/c theres a region in the key map that isn't in ami_regions")
	}

	c.AMIRegions = []string{"us-east-1", "us-west-1", "us-east-2"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
	}

	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("should have error b/c theres a region in in ami_regions that isn't in the key map")
	}

	c.SnapshotUsers = []string{"foo", "bar"}
	c.AMIKmsKeyId = "123-abc-456"
	c.AMIEncryptBootVolume = true
	c.AMIRegions = []string{"us-east-1", "us-west-1"}
	c.AMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have error b/c theres a region in in ami_regions that isn't in the key map")
	}

	// allow rawregion to exist in ami_regions list.
	accessConf = getFakeAccessConfig("us-east-1")
	c.AMIRegions = []string{"us-east-1", "us-west-1", "us-east-2"}
	c.AMIRegionKMSKeyIDs = nil
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should allow user to have the raw region in ami_regions")
	}

}

func TestAMIConfigPrepare_Share_EncryptedBoot(t *testing.T) {
	c := testAMIConfig()
	c.AMIUsers = []string{"testAccountID"}
	c.AMIEncryptBootVolume = true

	accessConf := testAccessConfig()

	c.AMIKmsKeyId = ""
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to share ami with encrypted boot volume")
	}

	c.AMIKmsKeyId = "89c3fb9a-de87-4f2a-aedc-fddc5138193c"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to share ami with encrypted boot volume")
	}
}

func TestAMIConfigPrepare_ValidateKmsKey(t *testing.T) {
	c := testAMIConfig()
	c.AMIEncryptBootVolume = true

	accessConf := testAccessConfig()

	validCases := []string{
		"abcd1234-e567-890f-a12b-a123b4cd56ef",
		"alias/foo/bar",
		"arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
		"arn:aws:kms:us-east-1:012345678910:alias/foo/bar",
	}
	for _, validCase := range validCases {
		c.AMIKmsKeyId = validCase
		if err := c.Prepare(accessConf, nil); err != nil {
			t.Fatalf("%s should not have failed KMS key validation", validCase)
		}
	}

	invalidCases := []string{
		"ABCD1234-e567-890f-a12b-a123b4cd56ef",
		"ghij1234-e567-890f-a12b-a123b4cd56ef",
		"ghij1234+e567_890f-a12b-a123b4cd56ef",
		"foo/bar",
		"arn:aws:kms:us-east-1:012345678910:foo/bar",
	}
	for _, invalidCase := range invalidCases {
		c.AMIKmsKeyId = invalidCase
		if err := c.Prepare(accessConf, nil); err == nil {
			t.Fatalf("%s should have failed KMS key validation", invalidCase)
		}
	}

}

func TestAMINameValidation(t *testing.T) {
	c := testAMIConfig()

	accessConf := testAccessConfig()

	c.AMIName = "aa"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an ami name with less than 3 characters")
	}

	var longAmiName string
	for i := 0; i < 129; i++ {
		longAmiName += "a"
	}
	c.AMIName = longAmiName
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an ami name with great than 128 characters")
	}

	c.AMIName = "+aaa"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an ami name with invalid characters")
	}

	c.AMIName = "fooBAR1()[] ./-'@_"
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatal("should be able to use all of the allowed AMI characters")
	}

	c.AMIName = `xyz-base-2017-04-05-1934`
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatalf("expected `xyz-base-2017-04-05-1934` to pass validation.")
	}

}
