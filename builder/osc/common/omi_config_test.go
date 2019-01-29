package common

import (
	"fmt"
	"reflect"
	"testing"
)

func testOMIConfig() *OMIConfig {
	return &OMIConfig{
		OMIName: "foo",
	}
}

func getFakeAccessConfig(region string) *AccessConfig {
	c := testAccessConfig()
	c.RawRegion = region
	return c
}

func TestOMIConfigPrepare_name(t *testing.T) {
	c := testOMIConfig()
	accessConf := testAccessConfig()
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.OMIName = ""
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestOMIConfigPrepare_regions(t *testing.T) {
	c := testOMIConfig()
	c.OMIRegions = nil

	var errs []error
	var err error
	accessConf := testAccessConfig()
	mockConn := &mockOAPIClient{}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}

	c.OMIRegions, err = listOAPIRegions(mockConn)
	if err != nil {
		t.Fatalf("shouldn't have err: %s", err.Error())
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("shouldn't have err: %#v", errs)
	}
	errs = errs[:0]

	c.OMIRegions = []string{"us-east-1", "us-west-1", "us-east-1"}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatalf("bad: %s", errs[0])
	}

	expected := []string{"us-east-1", "us-west-1"}
	if !reflect.DeepEqual(c.OMIRegions, expected) {
		t.Fatalf("bad: %#v", c.OMIRegions)
	}

	c.OMIRegions = []string{"custom"}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("shouldn't have error")
	}

	c.OMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "456-789-0123",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal(fmt.Sprintf("shouldn't have error: %s", errs[0]))
	}

	c.OMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have passed; we are able to use default KMS key if not sharing")
	}

	c.SnapshotUsers = []string{"user-foo", "user-bar"}
	c.OMIRegions = []string{"us-east-1", "us-east-2", "us-west-1"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have an error b/c can't use default KMS key if sharing")
	}

	c.OMIRegions = []string{"us-east-1", "us-west-1"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
		"us-east-2": "456-789-0123",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have error b/c theres a region in the key map that isn't in omi_regions")
	}

	c.OMIRegions = []string{"us-east-1", "us-west-1", "us-east-2"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "789-012-3456",
	}

	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("should have error b/c theres a region in in omi_regions that isn't in the key map")
	}

	c.SnapshotUsers = []string{"foo", "bar"}
	c.OMIKmsKeyId = "123-abc-456"
	c.OMIEncryptBootVolume = true
	c.OMIRegions = []string{"us-east-1", "us-west-1"}
	c.OMIRegionKMSKeyIDs = map[string]string{
		"us-east-1": "123-456-7890",
		"us-west-1": "",
	}
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should have error b/c theres a region in in omi_regions that isn't in the key map")
	}

	// allow rawregion to exist in omi_regions list.
	accessConf = getFakeAccessConfig("us-east-1")
	c.OMIRegions = []string{"us-east-1", "us-west-1", "us-east-2"}
	c.OMIRegionKMSKeyIDs = nil
	if errs = c.prepareRegions(accessConf); len(errs) > 0 {
		t.Fatal("should allow user to have the raw region in omi_regions")
	}

}

func TestOMIConfigPrepare_Share_EncryptedBoot(t *testing.T) {
	c := testOMIConfig()
	c.OMIUsers = []string{"testAccountID"}
	c.OMIEncryptBootVolume = true

	accessConf := testAccessConfig()

	c.OMIKmsKeyId = ""
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to share omi with encrypted boot volume")
	}

	c.OMIKmsKeyId = "89c3fb9a-de87-4f2a-aedc-fddc5138193c"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to share omi with encrypted boot volume")
	}
}

func TestOMIConfigPrepare_ValidateKmsKey(t *testing.T) {
	c := testOMIConfig()
	c.OMIEncryptBootVolume = true

	accessConf := testAccessConfig()

	validCases := []string{
		"abcd1234-e567-890f-a12b-a123b4cd56ef",
		"alias/foo/bar",
		"arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef",
		"arn:aws:kms:us-east-1:012345678910:alias/foo/bar",
	}
	for _, validCase := range validCases {
		c.OMIKmsKeyId = validCase
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
		c.OMIKmsKeyId = invalidCase
		if err := c.Prepare(accessConf, nil); err == nil {
			t.Fatalf("%s should have failed KMS key validation", invalidCase)
		}
	}

}

func TestOMINameValidation(t *testing.T) {
	c := testOMIConfig()

	accessConf := testAccessConfig()

	c.OMIName = "aa"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an omi name with less than 3 characters")
	}

	var longOmiName string
	for i := 0; i < 129; i++ {
		longOmiName += "a"
	}
	c.OMIName = longOmiName
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an omi name with great than 128 characters")
	}

	c.OMIName = "+aaa"
	if err := c.Prepare(accessConf, nil); err == nil {
		t.Fatal("shouldn't be able to have an omi name with invalid characters")
	}

	c.OMIName = "fooBAR1()[] ./-'@_"
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatal("should be able to use all of the allowed OMI characters")
	}

	c.OMIName = `xyz-base-2017-04-05-1934`
	if err := c.Prepare(accessConf, nil); err != nil {
		t.Fatalf("expected `xyz-base-2017-04-05-1934` to pass validation.")
	}

}
