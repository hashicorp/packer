package common

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
)

type ec2Describer interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

var (
	// modified in tests
	sshHostSleepDuration = time.Second
)

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address based on the instance DNS name.
func SSHHost(e ec2Describer, sshInterface string, host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}

		if sshInterface == "session_manager" {
			return "127.0.0.1", nil
		}

		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			i := state.Get("instance").(*ec2.Instance)
			if sshInterface != "" {
				switch sshInterface {
				case "public_ip":
					if i.PublicIpAddress != nil {
						host = *i.PublicIpAddress
					}
				case "private_ip":
					if i.PrivateIpAddress != nil {
						host = *i.PrivateIpAddress
					}
				case "public_dns":
					if i.PublicDnsName != nil {
						host = *i.PublicDnsName
					}
				case "private_dns":
					if i.PrivateDnsName != nil {
						host = *i.PrivateDnsName
					}
				default:
					panic(fmt.Sprintf("Unknown interface type: %s", sshInterface))
				}
			} else if i.VpcId != nil && *i.VpcId != "" {
				if i.PublicIpAddress != nil && *i.PublicIpAddress != "" {
					host = *i.PublicIpAddress
				} else if i.PrivateIpAddress != nil && *i.PrivateIpAddress != "" {
					host = *i.PrivateIpAddress
				}
			} else if i.PublicDnsName != nil && *i.PublicDnsName != "" {
				host = *i.PublicDnsName
			}

			if host != "" {
				return host, nil
			}

			r, err := e.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{i.InstanceId},
			})
			if err != nil {
				return "", err
			}

			if len(r.Reservations) == 0 || len(r.Reservations[0].Instances) == 0 {
				return "", fmt.Errorf("instance not found: %s", *i.InstanceId)
			}

			state.Put("instance", r.Reservations[0].Instances[0])
			time.Sleep(sshHostSleepDuration)
		}

		return "", errors.New("couldn't determine address for instance")
	}
}

// Port returns a function that can be given to the communicator
// for determining the port to use when connecting to an instance.
func Port(sshInterface string, port int) func(multistep.StateBag) (int, error) {
	return func(state multistep.StateBag) (int, error) {
		if sshInterface != "session_manager" {
			return port, nil
		}

		port, ok := state.GetOk("sessionPort")
		if !ok {
			return 0, fmt.Errorf("no local port defined for session-manager")
		}
		return port.(int), nil

	}
}
