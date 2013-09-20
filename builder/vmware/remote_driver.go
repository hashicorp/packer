package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
)

type RemoteDriver interface {
	Driver
	SSHAddress() func(multistep.StateBag) (string, error)
	Download() func(*common.DownloadConfig, multistep.StateBag) (string, error, bool)
}

func NewRemoteDriver(config *config) (Driver, error) {
	var driver Driver

	switch config.RemoteType {
	case "esx5":
		driver = &ESX5Driver{
			config: config,
		}
	default:
		return nil, fmt.Errorf("Unknown product type: '%s'", config.RemoteType)
	}

	return driver, driver.Verify()
}
