package common

import (
	commonssh "github.com/hashicorp/packer/common/ssh"
	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	"golang.org/x/crypto/ssh"
)

// CommHost returns the VM's IP address which should be used to access it by SSH.
func CommHost(state multistep.StateBag) (string, error) {
	vmName := state.Get("vmName").(string)
	driver := state.Get("driver").(Driver)

	mac, err := driver.MAC(vmName)
	if err != nil {
		return "", err
	}

	ip, err := driver.IPAddress(mac)
	if err != nil {
		return "", err
	}

	return ip, nil
}

// SSHConfigFunc returns SSH credentials to access the VM by SSH.
func SSHConfigFunc(config SSHConfig) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		auth := []ssh.AuthMethod{
			ssh.Password(config.Comm.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		}

		if config.Comm.SSHPrivateKey != "" {
			signer, err := commonssh.FileSigner(config.Comm.SSHPrivateKey)
			if err != nil {
				return nil, err
			}

			auth = append(auth, ssh.PublicKeys(signer))
		}

		return &ssh.ClientConfig{
			User:            config.Comm.SSHUsername,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}, nil
	}
}
