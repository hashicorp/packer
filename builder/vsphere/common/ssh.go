package common

import (
	"errors"
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/communicator/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func CommHost(state multistep.StateBag) (string, error) {
	driver := state.Get("driver").(Driver)

	//Removed to allow address change during the build
	/*if ipAddress, ok := state.GetOk("vm_address"); ok {
		return ipAddress.(string), nil
	}*/
	log.Println("Lookup up IP information...")
	ipAddress, err := driver.GuestIP()
	if err != nil {
		log.Printf("IP lookup failed: %s", err)
		return "", fmt.Errorf("IP lookup failed: %s", err)

	}
	if ipAddress == "" {
		log.Println("IP is blank, no IP yet.")
		return "", errors.New("IP is blank")
	}

	log.Printf("Detected IP: %s", ipAddress)
	state.Put("vm_address", ipAddress)
	return ipAddress, nil
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
			User: config.Comm.SSHUsername,
			Auth: auth,
		}, nil
	}
}
