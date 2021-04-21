package hyperone

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/hashicorp/packer-plugin-sdk/json"
)

const (
	sshAddress   = "api.hyperone.com:22"
	sshSubsystem = "rbx-auth"
	hostKeyHash  = "3e2aa423d42d7e8b14d50625512c8ac19db767ed"
)

type sshData struct {
	ID string `json:"_id"`
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func fetchTokenBySSH(user string) (string, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			hash := sha1Sum(key)
			if hash != hostKeyHash {
				return fmt.Errorf("invalid host key hash: %s", hash)
			}

			return nil
		},
	}

	client, err := ssh.Dial("tcp", sshAddress, sshConfig)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = session.RequestSubsystem(sshSubsystem)
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	var data sshData
	err = json.Unmarshal(out, &data)
	if err != nil {
		return "", err
	}

	return data.ID, nil
}

func sha1Sum(pubKey ssh.PublicKey) string {
	sum := sha1.Sum(pubKey.Marshal())
	return hex.EncodeToString(sum[:])
}
