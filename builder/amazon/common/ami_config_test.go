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
	c.AMISkipRegionValidation = true
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
	c.AMISkipRegionValidation = true

	var errs []error
	var err error
	accessConf := testAccessConfig()
	mockConn := &mockEC2Client{}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}

	c.AMISkipRegionValidation = false
	c.AMIRegions, err = listEC2Regions(mockConn)
	if err != nil {
		t.Fatalf("shouldn't have err: %s", err.Error())
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}

	c.AMIRegions = []string{"foo"}
	if errs = c.prepareRegions(accessConf); len(errs) == 0 {
		t.Fatal("should have error")
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
	c.AMISkipRegionValidation = true
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("shouldn't have error")
	}
	c.AMISkipRegionValidation = false

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

	c.AMISkipRegionValidation = true
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("should have error b/c theres a region in in ami_regions that isn't in the key map")
	}
	c.AMISkipRegionValidation = false

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
	c.AMISkipRegionValidation = true
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

func TestAMINameValidation(t *testing.T) {
	c := testAMIConfig()
	c.AMISkipRegionValidation = true

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
