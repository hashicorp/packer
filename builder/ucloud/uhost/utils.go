package uhost

import (
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"strings"
)

func checkStringIn(val string, availables []string) error {
	for _, choice := range availables {
		if val == choice {
			return nil
		}
	}

	return fmt.Errorf("should be one of %q, got %q", strings.Join(availables, ","), val)
}

func checkIntIn(val int, availables []int) error {
	for _, choice := range availables {
		if val == choice {
			return nil
		}
	}

	return fmt.Errorf("should be one of %v, got %d", availables, val)
}

func isStringIn(val string, availables []string) bool {
	for _, choice := range availables {
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
			if v.Type == "Private" {
				privateIp = v.IP
			} else {
				publicIp = v.IP
			}
		}
		if usePrivateIp {
			return privateIp, nil
		} else {
			return publicIp, nil
		}
	}
}

func halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}
