package ncloud

import (
	"strings"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":                "access_key",
		"secret_key":                "secret_key",
		"server_image_product_code": "SPSW0WINNT000016",
		"server_product_code":       "SPSVRSSD00000011",
		"server_image_name":         "packer-test {{timestamp}}",
		"server_image_description":  "server description",
		"block_storage_size":        100,
		"user_data":                 "#!/bin/sh\nyum install -y httpd\ntouch /var/www/html/index.html\nchkconfig --level 2345 httpd on",
		"region":                    "Korea",
		"access_control_group_configuration_no": "33",
		"communicator":                          "ssh",
		"ssh_username":                          "root",
	}
}

func testConfigForMemberServerImage() map[string]interface{} {
	return map[string]interface{}{
		"access_key":               "access_key",
		"secret_key":               "secret_key",
		"server_product_code":      "SPSVRSSD00000011",
		"member_server_image_no":   "2440",
		"server_image_name":        "packer-test {{timestamp}}",
		"server_image_description": "server description",
		"block_storage_size":       100,
		"user_data":                "#!/bin/sh\nyum install -y httpd\ntouch /var/www/html/index.html\nchkconfig --level 2345 httpd on",
		"region":                   "Korea",
		"access_control_group_configuration_no": "33",
		"communicator":                          "ssh",
		"ssh_username":                          "root",
	}
}

func TestConfigWithServerImageProductCode(t *testing.T) {
	raw := testConfig()

	c, _, _ := NewConfig(raw)

	if c.AccessKey != "access_key" {
		t.Errorf("Expected 'access_key' to be set to '%s', but got '%s'.", raw["access_key"], c.AccessKey)
	}

	if c.SecretKey != "secret_key" {
		t.Errorf("Expected 'secret_key' to be set to '%s', but got '%s'.", raw["secret_key"], c.SecretKey)
	}

	if c.ServerImageProductCode != "SPSW0WINNT000016" {
		t.Errorf("Expected 'server_image_product_code' to be set to '%s', but got '%s'.", raw["server_image_product_code"], c.ServerImageProductCode)
	}

	if c.ServerProductCode != "SPSVRSSD00000011" {
		t.Errorf("Expected 'server_product_code' to be set to '%s', but got '%s'.", raw["server_product_code"], c.ServerProductCode)
	}

	if c.BlockStorageSize != 100 {
		t.Errorf("Expected 'block_storage_size' to be set to '%d', but got '%d'.", raw["block_storage_size"], c.BlockStorageSize)
	}

	if c.ServerImageDescription != "server description" {
		t.Errorf("Expected 'server_image_description_key' to be set to '%s', but got '%s'.", raw["server_image_description"], c.ServerImageDescription)
	}

	if c.Region != "Korea" {
		t.Errorf("Expected 'region' to be set to '%s', but got '%s'.", raw["server_image_description"], c.Region)
	}
}

func TestConfigWithMemberServerImageCode(t *testing.T) {
	raw := testConfigForMemberServerImage()

	c, _, _ := NewConfig(raw)

	if c.AccessKey != "access_key" {
		t.Errorf("Expected 'access_key' to be set to '%s', but got '%s'.", raw["access_key"], c.AccessKey)
	}

	if c.SecretKey != "secret_key" {
		t.Errorf("Expected 'secret_key' to be set to '%s', but got '%s'.", raw["secret_key"], c.SecretKey)
	}

	if c.MemberServerImageNo != "2440" {
		t.Errorf("Expected 'member_server_image_no' to be set to '%s', but got '%s'.", raw["member_server_image_no"], c.MemberServerImageNo)
	}

	if c.ServerProductCode != "SPSVRSSD00000011" {
		t.Errorf("Expected 'server_product_code' to be set to '%s', but got '%s'.", raw["server_product_code"], c.ServerProductCode)
	}

	if c.BlockStorageSize != 100 {
		t.Errorf("Expected 'block_storage_size' to be set to '%d', but got '%d'.", raw["block_storage_size"], c.BlockStorageSize)
	}

	if c.ServerImageDescription != "server description" {
		t.Errorf("Expected 'server_image_description_key' to be set to '%s', but got '%s'.", raw["server_image_description"], c.ServerImageDescription)
	}

	if c.Region != "Korea" {
		t.Errorf("Expected 'region' to be set to '%s', but got '%s'.", raw["server_image_description"], c.Region)
	}
}

func TestEmptyConfig(t *testing.T) {
	raw := new(map[string]interface{})

	_, _, err := NewConfig(raw)

	if err == nil {
		t.Error("Expected Config to require 'access_key', 'secret_key' and some mendatory fields, but it did not")
	}

	if !strings.Contains(err.Error(), "access_key is required") {
		t.Error("Expected Config to require 'access_key', but it did not")
	}

	if !strings.Contains(err.Error(), "secret_key is required") {
		t.Error("Expected Config to require 'secret_key', but it did not")
	}

	if !strings.Contains(err.Error(), "server_image_product_code or member_server_image_no is required") {
		t.Error("Expected Config to require 'server_image_product_code' or 'member_server_image_no', but it did not")
	}
}

func TestExistsBothServerImageProductCodeAndMemberServerImageNoConfig(t *testing.T) {
	raw := map[string]interface{}{
		"access_key":                "access_key",
		"secret_key":                "secret_key",
		"server_image_product_code": "SPSW0WINNT000016",
		"server_product_code":       "SPSVRSSD00000011",
		"member_server_image_no":    "2440",
	}

	_, _, err := NewConfig(raw)

	if !strings.Contains(err.Error(), "Only one of server_image_product_code and member_server_image_no can be set") {
		t.Error("Expected Config to require Only one of 'server_image_product_code' and 'member_server_image_no' can be set, but it did not")
	}
}
