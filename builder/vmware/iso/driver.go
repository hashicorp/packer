package iso

import vmwcommon "github.com/mitchellh/packer/builder/vmware/common"

// NewDriver returns a new driver implementation for this operating
// system, or an error if the driver couldn't be initialized.
func NewDriver(config *Config) (vmwcommon.Driver, error) {
	return vmwcommon.NewDriver(&config.DriverConfig, &config.SSHConfig, &config.CommConfig, config.VMName)
}
