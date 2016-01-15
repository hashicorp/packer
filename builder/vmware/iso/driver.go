package iso

import (
	"fmt"

	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
)

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver(config *Config) (vmwcommon.Driver, error) {
	drivers := []vmwcommon.Driver{}

	if config.RemoteType == "" {
		return vmwcommon.NewDriver(&config.DriverConfig, &config.SSHConfig)
	}

	drivers = []vmwcommon.Driver{
		&ESX5Driver{
			Host:           config.RemoteHost,
			Port:           config.RemotePort,
			Username:       config.RemoteUser,
			Password:       config.RemotePassword,
			PrivateKey:     config.RemotePrivateKey,
			Datastore:      config.RemoteDatastore,
			CacheDatastore: config.RemoteCacheDatastore,
			CacheDirectory: config.RemoteCacheDirectory,
		},
	}

	errs := ""
	for _, driver := range drivers {
		err := driver.Verify()
		if err == nil {
			return driver, nil
		}
		errs += "* " + err.Error() + "\n"
	}

	return nil, fmt.Errorf(
		"Unable to initialize any driver for this platform. The errors\n"+
			"from each driver are shown below. Please fix at least one driver\n"+
			"to continue:\n%s", errs)
}
