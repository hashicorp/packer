package common

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func CheckStringIn(val string, available []string) error {
	for _, choice := range available {
		if val == choice {
			return nil
		}
	}

	return fmt.Errorf("should be one of %q, got %q", strings.Join(available, ","), val)
}

func CheckIntIn(val int, available []int) error {
	for _, choice := range available {
		if val == choice {
			return nil
		}
	}

	return fmt.Errorf("should be one of %v, got %d", available, val)
}

func IsStringIn(val string, available []string) bool {
	for _, choice := range available {
		if val == choice {
			return true
		}
	}

	return false
}

// SSHHost returns a function that can be given to the SSH communicator
func SSHHost(usePrivateIp bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {

		instance := state.Get("instance").(*uhost.UHostInstanceSet)
		var privateIp, publicIp string

		for _, v := range instance.IPSet {
			if v.Type == IpTypePrivate {
				privateIp = v.IP
			} else {
				publicIp = v.IP
			}
		}

		if usePrivateIp {
			return privateIp, nil
		}

		return publicIp, nil
	}
}

func Halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}
