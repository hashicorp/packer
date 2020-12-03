package oci

import (
	"bytes"
	"os"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// TODO(apryde): It would be good not to have to write a key file to disk to
// load the config.
func baseTestConfig() *Config {
	_, keyFile, err := baseTestConfigWithTmpKeyFile()
	if err != nil {
		panic(err)
	}

	var c Config
	err = c.Prepare(map[string]interface{}{
		"availability_domain": "aaaa:US-ASHBURN-AD-1",

		// Image
		"base_image_ocid": "ocid1.image.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"shape":           "VM.Standard1.1",
		"image_name":      "HelloWorld",
		"region":          "us-ashburn-1",

		// Networking
		"subnet_ocid": "ocid1.subnet.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",

		// AccessConfig
		"user_ocid":    "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"tenancy_ocid": "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"fingerprint":  "70:04:5z:b3:19:ab:90:75:a4:1f:50:d4:c7:c3:33:20",
		"key_file":     keyFile.Name(),

		// Comm
		"ssh_username":   "opc",
		"use_private_ip": false,
	})

	// Once we have a config object they key file isn't re-read so we can
	// remove it now.
	os.Remove(keyFile.Name())

	if err != nil {
		panic(err)
	}
	return &c
}

func testState() multistep.StateBag {
	baseTestConfig := baseTestConfig()
	state := new(multistep.BasicStateBag)
	state.Put("config", baseTestConfig)
	state.Put("driver", &driverMock{cfg: baseTestConfig})
	state.Put("hook", &packersdk.MockHook{})
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
