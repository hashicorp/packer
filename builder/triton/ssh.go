package triton

import (
	"fmt"

	"github.com/mitchellh/multistep"
	packerssh "github.com/mitchellh/packer/communicator/ssh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func commHost(state multistep.StateBag) (string, error) {
	driver := state.Get("driver").(Driver)
	machineID := state.Get("machine").(string)

	machine, err := driver.GetMachine(machineID)
	if err != nil {
		return "", err
	}

	return machine, nil
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the private key
// or password.
func sshConfig(useAgent bool, username, privateKeyPath, password string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {

		if useAgent {
			log.Println("Configuring SSH agent.")

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
			}, nil
		}

		hasKey := privateKeyPath != ""

		if hasKey {
			log.Printf("Configuring SSH private key '%s'.", privateKeyPath)

			privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
			if err != nil {
				return nil, fmt.Errorf("Unable to read SSH private key: %s", err)
			}

			signer, err := ssh.ParsePrivateKey(privateKeyBytes)
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}

			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
			}, nil
		} else {
			log.Println("Configuring SSH keyboard interactive.")

			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.Password(password),
					ssh.KeyboardInteractive(
						packerssh.PasswordKeyboardInteractive(password)),
				}}, nil
		}
	}
}