package common

import (
	"fmt"

	"github.com/mitchellh/multistep"
)

func WinRMAddressFunc(config *WinRMConfig, driver Driver) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if config.WinRMHost != "" {
			return fmt.Sprintf("%s:%d", config.WinRMHost, config.WinRMPort), nil
		}

		ipAddress, err := driver.IPAddress(state)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s:%d", ipAddress, config.WinRMPort), nil
	}
}
