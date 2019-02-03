package common

import (
	"crypto/rand"
	"errors"
	"strconv"
	"testing"

	"golang.org/x/crypto/ssh"
)

// expected contains the data that the key pair should contain.
type expected struct {
	kind sshKeyPairType
	bits int
	desc string
}

func (o expected) matches(kp sshKeyPair) error {
	if o.kind.String() == "" {
		return errors.New("expected kind's value cannot be empty")
	}

	if o.bits <= 0 {
		return errors.New("expected bits' value cannot be less than or equal to 0")
	}

	if o.desc == "" {
		return errors.New("expected description's value cannot be empty")
	}

	if kp.Type() != o.kind {
		return errors.New("expected key pair type to be " +
			o.kind.String() + " - got '" + kp.Type().String() + "'")
	}

	if kp.Bits() != o.bits {
		return errors.New("expected key pair to be " +
			strconv.Itoa(o.bits) + " bits - got " + strconv.Itoa(kp.Bits()))
	}

	expDescription := kp.Type().String() + " " + strconv.Itoa(o.bits)
	if kp.Description() != expDescription {
		return errors.New("expected key pair description to be '" +
			expDescription + "' - got '" + kp.Description() + "'")
	}

	err := verifySshKeyPair(kp)
	if err != nil {
		return err
	}

	return nil
}

func TestDefaultSshKeyPairBuilder_Build_Default(t *testing.T) {
	kp, err := newSshKeyPairBuilder().Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: ecdsaSsh,
		bits: 521,
		desc: "ecdsa 521",
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultSshKeyPairBuilder_Build_EcdsaDefault(t *testing.T) {
	kp, err := newSshKeyPairBuilder().SetType(ecdsaSsh).Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: ecdsaSsh,
		bits: 521,
		desc: "ecdsa 521",
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultSshKeyPairBuilder_Build_RsaDefault(t *testing.T) {
	kp, err := newSshKeyPairBuilder().SetType(rsaSsh).Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: rsaSsh,
		bits: 4096,
		desc: "rsa 4096",
	}.matches(kp)
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
