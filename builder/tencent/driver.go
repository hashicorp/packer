package tencent

import "github.com/hashicorp/packer/helper/multistep"

type (
	// An interface for communicating with any cloud provider
	Driver interface {
		// CWCreateImage creates an image based on the configuration given in the Config
		// and returns a failure bool (true for failure), the error, and the instance info.
		CWCreateImage(config Config) (bool, CVMError, CVMInstanceInfo)

		// CWCreateCustomImage creates an image based on the given configuration and returns
		// a failure bool (true for failure), the error, and information returned by the Tencent API
		CWCreateCustomImage(config Config, instanceId string) (bool, CVMError, CVMCreateCustomImage)

		// Waits for a new image, and returns a boolean and a string (ImageID)
		CWWaitForCustomImageReady(config Config) (bool, string)

		CWGetImageState(config Config, instanceId string) (error, string)

		CWGetInstanceIP(config Config, instanceId string) (error, string)

		CWStopImage(config Config, instanceId string) error

		CWRunImage(config Config, instanceId string) error

		CWWaitForImageState(c Config, instanceId string, state string) error

		// This actually creates a keypair and associates it with a VM instance
		CWCreateKeyPair(c Config, instanceId string, state multistep.StateBag) (error, CVMKeyPair)
		CWWaitKeyPairAttached(c Config, instanceId, KeyId string) error
	}
)
