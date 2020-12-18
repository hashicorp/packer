// Package SSH provides tooling for generating a temporary SSH keypair, and
// provides tooling for connecting to an instance via a tunnel.
package ssh

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func parseKeyFile(path string) ([]byte, error) {
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
	return keyBytes, nil
}

// FileSigner returns an ssh.Signer for a key file.
func FileSigner(path string) (ssh.Signer, error) {

	keyBytes, err := parseKeyFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}

func ReadCertificate(certificatePath string, keySigner ssh.Signer) (ssh.Signer, error) {

	if certificatePath == "" {
		return keySigner, fmt.Errorf("no certificate file provided")
	}

	// Load the certificate
	cert, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read certificate file: %v", err)
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey(cert)
	if err != nil {
		return nil, fmt.Errorf("unable to parse public key: %v", err)
	}

	certificate, ok := pk.(*ssh.Certificate)

	if !ok {
		return nil, fmt.Errorf("Error loading certificate")
	}

	err = checkValidCert(certificate)

	if err != nil {
		return nil, fmt.Errorf("%s not a valid cert: %v", certificatePath, err)
	}

	certSigner, err := ssh.NewCertSigner(certificate, keySigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert signer: %v", err)
	}

	return certSigner, nil
}

// FileSigner returns an ssh.Signer for a key file.
func FileSignerWithCert(path string, certificatePath string) (ssh.Signer, error) {

	keySigner, err := FileSigner(path)

	if err != nil {
		return nil, err
	}
	return ReadCertificate(certificatePath, keySigner)
}

func checkValidCert(cert *ssh.Certificate) error {
	const CertTimeInfinity = 1<<64 - 1
	unixNow := time.Now().Unix()

	if after := int64(cert.ValidAfter); after < 0 || unixNow < int64(cert.ValidAfter) {
		return fmt.Errorf("ssh: cert is not yet valid")
	}
	if before := int64(cert.ValidBefore); cert.ValidBefore != uint64(CertTimeInfinity) && (unixNow >= before || before < 0) {
		return fmt.Errorf("ssh: cert has expired")
	}
	return nil
}
