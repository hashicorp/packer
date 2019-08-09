package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/outscale/osc-go/oapi"
)

type oapiDescriber interface {
	POST_ReadVms(oapi.ReadVmsRequest) (*oapi.POST_ReadVmsResponses, error)
}

var (
	// modified in tests
	sshHostSleepDuration = time.Second
)

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address based on the vm DNS name.
func SSHHost(e oapiDescriber, sshInterface string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			var host string
			i := state.Get("vm").(oapi.Vm)
			if sshInterface != "" {
				switch sshInterface {
				case "public_ip":
					if i.PublicIp != "" {
						host = i.PublicIp
					}
				case "private_ip":
					if i.PrivateIp != "" {
						host = i.PrivateIp
					}
				case "public_dns":
					if i.PublicDnsName != "" {
						host = i.PublicDnsName
					}
				case "private_dns":
					if i.PrivateDnsName != "" {
						host = i.PrivateDnsName
					}
				default:
					panic(fmt.Sprintf("Unknown interface type: %s", sshInterface))
				}
			} else if i.NetId != "" {
				if i.PublicIp != "" {
					host = i.PublicIp
				} else if i.PrivateIp != "" {
					host = i.PrivateIp
				}
			} else if i.PublicDnsName != "" {
				host = i.PublicDnsName
			}

			if host != "" {
				return host, nil
			}

			r, err := e.POST_ReadVms(oapi.ReadVmsRequest{
				Filters: oapi.FiltersVm{
					VmIds: []string{i.VmId},
				},
			})
			if err != nil {
				return "", err
			}

			if len(r.OK.Vms) == 0 {
				return "", fmt.Errorf("vm not found: %s", i.VmId)
			}

			state.Put("vm", r.OK.Vms[0])
			time.Sleep(sshHostSleepDuration)
		}

		return "", errors.New("couldn't determine address for vm")
	}
}
