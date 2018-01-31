package classic

import (
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"identity_domain":   "abc12345",
		"username":          "test@hashicorp.com",
		"password":          "testpassword123",
		"api_endpoint":      "https://api-test.compute.test.oraclecloud.com/",
		"dest_image_list":   "/Config-thing/myuser/myimage",
		"source_image_list": "/oracle/public/whatever",
		"shape":             "oc3",
		"image_name":        "TestImageName",
		"ssh_username":      "opc",
	}
}

func TestConfigAutoFillsSourceList(t *testing.T) {
	tc := testConfig()
	conf, err := NewConfig(tc)
	if err != nil {
		t.Fatalf("Should not have error: %s", err.Error())
	}
	if conf.SSHSourceList != "seciplist:/oracle/public/public-internet" {
		t.Fatalf("conf.SSHSourceList should have been "+
			"\"seciplist:/oracle/public/public-internet\" but is \"%s\"",
			conf.SSHSourceList)
	}
}

func TestConfigValidationCatchesMissing(t *testing.T) {
	required := []string{
		"username",
		"password",
		"api_endpoint",
		"identity_domain",
		"dest_image_list",
		"source_image_list",
		"shape",
	}
	for _, key := range required {
		tc := testConfig()
		delete(tc, key)
		_, err := NewConfig(tc)
		if err == nil {
			t.Fatalf("Test should have failed when config lacked %s!", key)
		}
	}
}

func TestValidationsIgnoresOptional(t *testing.T) {
	tc := testConfig()
	delete(tc, "ssh_username")
	_, err := NewConfig(tc)
	if err != nil {
		t.Fatalf("Shouldn't care if ssh_username is missing: err: %#v", err.Error())
	}
}
