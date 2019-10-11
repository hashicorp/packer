package common

import (
	"errors"
	"fmt"
	"sort"
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

			if len(i.Nics) <= 0 {
				return "", errors.New("couldn't determine address for vm, nics are empty")
			}

			nic := i.Nics[0]

			if sshInterface != "" {
				switch sshInterface {
				case "public_ip":
					if nic.LinkPublicIp.PublicIp != "" {
						host = nic.LinkPublicIp.PublicIp
					}
				case "public_dns":
					if nic.LinkPublicIp.PublicDnsName != "" {
						host = nic.LinkPublicIp.PublicDnsName
					}
				case "private_ip":
					if privateIP, err := getPrivateIP(nic); err != nil {
						host = privateIP.PrivateIp
					}
				case "private_dns":
					if privateIP, err := getPrivateIP(nic); err != nil {
						host = privateIP.PrivateDnsName
					}
				default:
					panic(fmt.Sprintf("Unknown interface type: %s", sshInterface))
				}
			} else if i.NetId != "" {
				if nic.LinkPublicIp.PublicIp != "" {
					host = nic.LinkPublicIp.PublicIp
				} else if privateIP, err := getPrivateIP(nic); err != nil {
					host = privateIP.PrivateIp
				}
			} else if nic.LinkPublicIp.PublicDnsName != "" {
				host = nic.LinkPublicIp.PublicDnsName
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

func getPrivateIP(nic oapi.NicLight) (oapi.PrivateIpLightForVm, error) {
	isPrimary := true

	i := sort.Search(len(nic.PrivateIps), func(i int) bool { return nic.PrivateIps[i].IsPrimary == isPrimary })

	if i < len(nic.PrivateIps) && nic.PrivateIps[i].IsPrimary == isPrimary {
		return nic.PrivateIps[i], nil
	}

	if len(nic.PrivateIps) > 0 {
		return nic.PrivateIps[0], nil
	}

	return oapi.PrivateIpLightForVm{}, fmt.Errorf("couldn't determine private address for vm")

}
