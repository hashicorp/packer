package common

import (
	"crypto/rand"
	"errors"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestDefaultSshKeyPairBuilder_Build_Default(t *testing.T) {
	kp, err := newSshKeyPairBuilder().Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	if kp.Type() != ecdsaSsh {
		t.Fatal("Expected key pair type to be",
			ecdsaSsh.String(), "- got", kp.Type())
	}

	if kp.Bits() != 521 {
		t.Fatal("Expected key pair to be 521 bits - got", kp.Bits())
	}

	err = verifySshKeyPair(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultSshKeyPairBuilder_Build_EcdsaDefault(t *testing.T) {
	kp, err := newSshKeyPairBuilder().SetType(ecdsaSsh).Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	if kp.Type() != ecdsaSsh {
		t.Fatal("Expected key pair type to be",
			ecdsaSsh.String(), "- got", kp.Type())
	}

	if kp.Bits() != 521 {
		t.Fatal("Expected key pair to be 521 bits - got", kp.Bits())
	}

	err = verifySshKeyPair(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultSshKeyPairBuilder_Build_RsaDefault(t *testing.T) {
	kp, err := newSshKeyPairBuilder().SetType(rsaSsh).Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	if kp.Type() != rsaSsh {
		t.Fatal("Expected default key pair type to be",
			rsaSsh.String(), "- got", kp.Type())
	}

	if kp.Bits() != 4096 {
		t.Fatal("Expected key pair to be", 4096, "bits - got", kp.Bits())
	}

	err = verifySshKeyPair(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func verifySshKeyPair(kp sshKeyPair) error {
	signer, err := ssh.ParsePrivateKey(kp.PrivateKeyPemBlock())
	if err != nil {
		return errors.New("failed to parse private key during verification - " + err.Error())
	}

	data := []byte{'b', 'r', '4', 'n', '3'}

	signature, err := signer.Sign(rand.Reader, data)
	if err != nil {
		return errors.New("failed to sign test data during verification - " + err.Error())
	}

	err = signer.PublicKey().Verify(data, signature)
	if err != nil {
		return errors.New("failed to verify test data - " + err.Error())
	}

	return nil
}
