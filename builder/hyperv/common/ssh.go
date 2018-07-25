package common

import (
	commonssh "github.com/hashicorp/packer/common/ssh"
	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	gossh "golang.org/x/crypto/ssh"
)

func CommHost(state multistep.StateBag) (string, error) {
	vmName := state.Get("vmName").(string)
	driver := state.Get("driver").(Driver)

	mac, err := driver.Mac(vmName)
	if err != nil {
		return "", err
	}

	ip, err := driver.IpAddress(mac)
	if err != nil {
		return "", err
	}

	return ip, nil
}

func SSHConfigFunc(config *SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		auth := []gossh.AuthMethod{
			gossh.Password(config.Comm.SSHPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		}

		if config.Comm.SSHPrivateKey != "" {
			signer, err := commonssh.FileSigner(config.Comm.SSHPrivateKey)
			if err != nil {
				return nil, err
			}

			auth = append(auth, gossh.PublicKeys(signer))
		}

		return &gossh.ClientConfig{
			User:            config.Comm.SSHUsername,
			Auth:            auth,
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		}, nil
	}
}
