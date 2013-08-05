package common

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/communicator/ssh"
	"time"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the instance DNS name.
func SSHAddress(e *ec2.EC2, port int) func(map[string]interface{}) (string, error) {
	return func(state map[string]interface{}) (string, error) {
		for j := 0; j < 2; j++ {
			var host string
			i := state["instance"].(*ec2.Instance)
			if i.DNSName != "" {
				host = i.DNSName
			} else if i.VpcId != "" {
				host = i.PrivateIpAddress
			}

			if host != "" {
				return fmt.Sprintf("%s:%d", host, port), nil
			}

			r, err := e.Instances([]string{i.InstanceId}, ec2.NewFilter())
			if err != nil {
				return "", err
			}

			if len(r.Reservations) == 0 || len(r.Reservations[0].Instances) == 0 {
				return "", fmt.Errorf("instance not found: %s", i.InstanceId)
			}

			state["instance"] = &r.Reservations[0].Instances[0]
			time.Sleep(1 * time.Second)
		}

		return "", errors.New("couldn't determine IP address for instance")
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(map[string]interface{}) (*gossh.ClientConfig, error) {
	return func(state map[string]interface{}) (*gossh.ClientConfig, error) {
		privateKey := state["privateKey"].(string)

		keyring := new(ssh.SimpleKeychain)
		if err := keyring.AddPEMKey(privateKey); err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &gossh.ClientConfig{
			User: username,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}, nil
	}
}
