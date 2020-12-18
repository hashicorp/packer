package yandex

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
)

const InstanceMetadataAddr = "169.254.169.254"

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func TestBuilderAcc_instanceSA(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheckInstanceSA(t) },
		Builder:  &Builder{},
		Template: testBuilderAccInstanceSA,
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

func testAccPreCheckInstanceSA(t *testing.T) {
	client := resty.New()

	_, err := client.R().SetHeader("Metadata-Flavor", "Google").Get(tokenUrl())
	if err != nil {
		t.Fatalf("error get Service Account token assignment: %s", err)
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

const testBuilderAccInstanceSA = `
{
	"builders": [{
		"type": "test",
        "source_image_family": "ubuntu-1804-lts",
		"use_ipv4_nat": "true",
		"ssh_username": "ubuntu"
	}]
}
`

func tokenUrl() string {
	return fmt.Sprintf("http://%s/computeMetadata/v1/instance/service-accounts/default/token", InstanceMetadataAddr)
}
