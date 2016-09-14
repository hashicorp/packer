package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
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
func SSHHost(e ec2Describer, private bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			var host string
			i := state.Get("instance").(*ec2.Instance)
			if i.VpcId != nil && *i.VpcId != "" {
				if i.PublicIpAddress != nil && *i.PublicIpAddress != "" && !private {
					host = *i.PublicIpAddress
				} else if i.PrivateIpAddress != nil && *i.PrivateIpAddress != "" {
					host = *i.PrivateIpAddress
				}
			} else if private {
				host = *i.PrivateIpAddress
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

		return "", errors.New("couldn't determine IP address for instance")
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string)

		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		}, nil
	}
}
