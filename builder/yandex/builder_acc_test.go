package yandex

import (
	"os"
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("YC_TOKEN"); v == "" {
		t.Fatal("YC_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("YC_FOLDER_ID"); v == "" {
		t.Fatal("YC_FOLDER_ID must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
        "source_image_family": "ubuntu-1804-lts",
		"use_ipv4_nat": "true",
		"ssh_username": "ubuntu"
	}]
}
`
