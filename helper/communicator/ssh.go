package communicator

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

// SSHFileSigner returns an ssh.Signer for a key file.
func SSHFileSigner(path string) (ssh.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// We parse the private key on our own first so that we can
	// show a nicer error if the private key has a password.
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf(
			"Failed to read key '%s': no key found", path)
	}
	if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
		return nil, fmt.Errorf(
			"Failed to read key '%s': password protected keys are\n"+
				"not supported. Please decrypt the key prior to use.", path)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}
