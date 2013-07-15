package virtualbox

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/packer/communicator/ssh"
)

func sshAddress(state map[string]interface{}) (string, error) {
	sshHostPort := state["sshHostPort"].(uint)
	return fmt.Sprintf("127.0.0.1:%d", sshHostPort), nil
}

func sshConfig(state map[string]interface{}) (*gossh.ClientConfig, error) {
	config := state["config"].(*config)

	return &gossh.ClientConfig{
		User: config.SSHUser,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
			gossh.ClientAuthKeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		},
	}, nil
}
