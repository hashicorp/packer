package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"net/http"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/outscale/osc-sdk-go/osc"
)

type oscDescriber interface {
	ReadVms(ctx context.Context, localVarOptionals *osc.ReadVmsOpts) (osc.ReadVmsResponse, *http.Response, error)
}

var (
	// modified in tests
	sshHostSleepDuration = time.Second
)

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address based on the vm DNS name.
func SSHHost(e oscDescriber, sshInterface string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			var host string
			i := state.Get("vm").(osc.Vm)

			if sshInterface != "" {
				switch sshInterface {
				case "public_ip":
					if i.PublicIp != "" {
						host = i.PublicIp
					}
				case "public_dns":
					if i.PublicDnsName != "" {
						host = i.PublicDnsName
					}
				case "private_ip":
					if i.PrivateIp != "" {
						host = i.PrivateIp
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

			r, _, err := e.ReadVms(context.Background(), &osc.ReadVmsOpts{
				ReadVmsRequest: optional.NewInterface(osc.ReadVmsRequest{
					Filters: osc.FiltersVm{
						VmIds: []string{i.VmId},
					},
				}),
			})
			if err != nil {
				return "", err
			}

			if len(r.Vms) == 0 {
				return "", fmt.Errorf("vm not found: %s", i.VmId)
			}

			state.Put("vm", r.Vms[0])
			time.Sleep(sshHostSleepDuration)
		}

		return "", errors.New("couldn't determine address for vm")
	}
}

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address based on the vm DNS name.
func OscSSHHost(e oscDescriber, sshInterface string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			var host string
			i := state.Get("vm").(osc.Vm)

			if sshInterface != "" {
				switch sshInterface {
				case "public_ip":
					if i.PublicIp != "" {
						host = i.PublicIp
					}
				case "public_dns":
					if i.PublicDnsName != "" {
						host = i.PublicDnsName
					}
				case "private_ip":
					if i.PrivateIp != "" {
						host = i.PrivateIp
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

			r, _, err := e.ReadVms(context.Background(), &osc.ReadVmsOpts{
				ReadVmsRequest: optional.NewInterface(osc.ReadVmsRequest{
					Filters: osc.FiltersVm{
						VmIds: []string{i.VmId},
					},
				}),
			})

			if err != nil {
				return "", err
			}

			if len(r.Vms) == 0 {
				return "", fmt.Errorf("vm not found: %s", i.VmId)
			}

			state.Put("vm", r.Vms[0])
			time.Sleep(sshHostSleepDuration)
		}

		return "", errors.New("couldn't determine address for vm")
	}
}
