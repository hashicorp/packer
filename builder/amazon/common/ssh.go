package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
func SSHHost(e ec2Describer, sshInterface string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		const tries = 2
		// <= with current structure to check result of describing `tries` times
		for j := 0; j <= tries; j++ {
			var host string
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

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the private key
// or password.
func SSHConfig(useAgent bool, username, password string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		if useAgent {
			authSock := os.Getenv("SSH_AUTH_SOCK")
			if authSock == "" {
				return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
			}

			sshAgent, err := net.Dial("unix", authSock)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
			}

			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}, nil
		}

		privateKey, hasKey := state.GetOk("privateKey")
		if hasKey {

			signer, err := ssh.ParsePrivateKey([]byte(privateKey.(string)))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}
			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}, nil

		} else {
			return &ssh.ClientConfig{
				User:            username,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Auth: []ssh.AuthMethod{
					ssh.Password(password),
					ssh.KeyboardInteractive(
						packerssh.PasswordKeyboardInteractive(password)),
				}}, nil
		}
	}
}
