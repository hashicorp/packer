package ssh

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.google.com/p/go.crypto/ssh"
)

// FileSigner returns an ssh.Signer for a key file.
func FileSigner(path string) (ssh.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}
