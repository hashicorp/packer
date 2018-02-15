package oci

import (
	"bytes"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
)

// TODO(apryde): It would be good not to have to write a key file to disk to
// load the config.
func baseTestConfig() *Config {
	_, keyFile, err := client.BaseTestConfig()
	if err != nil {
		panic(err)
	}

	cfg, err := NewConfig(map[string]interface{}{
		"availability_domain": "aaaa:PHX-AD-3",

		// Image
		"base_image_ocid": "ocd1...",
		"shape":           "VM.Standard1.1",
		"image_name":      "HelloWorld",

		// Networking
		"subnet_ocid": "ocd1...",

		// AccessConfig
		"user_ocid":    "ocid1...",
		"tenancy_ocid": "ocid1...",
		"fingerprint":  "00:00...",
		"key_file":     keyFile.Name(),

		// Comm
		"ssh_username": "opc",
	})

	// Once we have a config object they key file isn't re-read so we can
	// remove it now.
	os.Remove(keyFile.Name())

	if err != nil {
		panic(err)
	}
	return cfg
}

func testState() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("config", baseTestConfig())
	state.Put("driver", &driverMock{})
	state.Put("hook", &packer.MockHook{})
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
