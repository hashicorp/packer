package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

// FileSigner returns an ssh.Signer for a key file.
func FileSigner(path string) (ssh.Signer, error) {
	return PassphraseFileSigner(path, "")
}

// PassphraseFileSigner returns an ssh.Signer for a key file.
func PassphraseFileSigner(path, passphrase string) (ssh.Signer, error) {
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
		if passphrase == "" {
			return nil, fmt.Errorf(
				"Failed to read key '%s': password protected keys require\n"+
					"ssh_private_key_passphrase to be set.", path)
		}
		return parseEncryptedKey(block, path, passphrase)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}

// parseEncryptedKey returns an ssh.Signer from an encrypted key file.
func parseEncryptedKey(block *pem.Block, path, passphrase string) (ssh.Signer, error) {
	der, err := x509.DecryptPEMBlock(block, []byte(passphrase))
	if err != nil {
		return nil, fmt.Errorf("Failed to decrypt key '%s': %s", path, err)
	}
	var key interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(der)
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(der)
	default:
		err = fmt.Errorf("unsupported key type %q", block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("Error decrypting private key: %s", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("Error setting up encrypted SSH config: %s", err)
	}
	return signer, nil
}
