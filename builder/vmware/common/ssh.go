package common

import (
	"errors"
	"fmt"
	"log"

	commonssh "github.com/hashicorp/packer/common/ssh"
	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	gossh "golang.org/x/crypto/ssh"
)

func CommHost(config *SSHConfig) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		driver := state.Get("driver").(Driver)

		if config.Comm.SSHHost != "" {
			return config.Comm.SSHHost, nil
		}

		ipAddress, err := driver.GuestIP(state)
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			return "", fmt.Errorf("IP lookup failed: %s", err)
		}

		if ipAddress == "" {
			log.Println("IP is blank, no IP yet.")
			return "", errors.New("IP is blank")
		}

		log.Printf("Detected IP: %s", ipAddress)
		return ipAddress, nil
	}
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
