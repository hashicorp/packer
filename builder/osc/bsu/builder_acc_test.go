/*
Deregister the test image with
aws oapi deregister-image --image-id $(aws oapi describe-images --output text --filters "Name=name,Values=packer-test-packer-test-dereg" --query 'Images[*].{ID:ImageId}')
*/
package bsu

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/hashicorp/packer/builder/osc/common"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/outscale/osc-go/oapi"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
}

func testOAPIConn() (*oapi.Client, error) {
	access := &common.AccessConfig{RawRegion: "us-east-1"}
	clientConfig, err := access.Config()
	if err != nil {
		return nil, err
	}

	skipClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return oapi.NewClient(clientConfig, skipClient), nil
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "eu-west-2",
		"vm_type": "m3.medium",
		"source_omi": "ami-46260446",
		"ssh_username": "ubuntu",
		"omi_name": "packer-test {{timestamp}}"
	}]
}
`
