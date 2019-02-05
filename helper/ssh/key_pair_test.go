package ssh

import (
	"bytes"
	"crypto/rand"
	"errors"
	"strconv"
	"testing"

	"github.com/hashicorp/packer/common/uuid"
	gossh "golang.org/x/crypto/ssh"
)

// expected contains the data that the key pair should contain.
type expected struct {
	kind KeyPairType
	bits int
	desc string
	name string
	data []byte
}

func (o expected) matches(kp KeyPair) error {
	if o.kind.String() == "" {
		return errors.New("expected kind's value cannot be empty")
	}

	if o.bits <= 0 {
		return errors.New("expected bits' value cannot be less than or equal to 0")
	}

	if o.desc == "" {
		return errors.New("expected description's value cannot be empty")
	}

	if len(o.data) == 0 {
		return errors.New("expected random data value cannot be nothing")
	}

	if kp.Type() != o.kind {
		return errors.New("key pair type should be " + o.kind.String() +
			" - got '" + kp.Type().String() + "'")
	}

	if kp.Bits() != o.bits {
		return errors.New("key pair bits should be " + strconv.Itoa(o.bits) +
			" - got " + strconv.Itoa(kp.Bits()))
	}

	if len(o.name) > 0 && kp.Name() != o.name {
		return errors.New("key pair name should be '" + o.name +
			"' - got '" + kp.Name() + "'")
	}

	if kp.Description() != o.desc {
		return errors.New("key pair description should be '" +
			o.desc + "' - got '" + kp.Description() + "'")
	}

	err := o.verifyPublicKeyAuthorizedKeysFormat(kp)
	if err != nil {
		return err
	}

	err = o.verifyKeyPair(kp)
	if err != nil {
		return err
	}

	return nil
}

func (o expected) verifyPublicKeyAuthorizedKeysFormat(kp KeyPair) error {
	newLines := []NewLineOption{
		UnixNewLine,
		NoNewLine,
		WindowsNewLine,
	}

	for _, nl := range newLines {
		publicKeyAk := kp.PublicKeyAuthorizedKeysLine(nl)

		if len(publicKeyAk) < 2 {
			return errors.New("expected public key in authorized keys format to be at least 2 bytes")
		}

		switch nl {
		case NoNewLine:
			if publicKeyAk[len(publicKeyAk) - 1] == '\n' {
				return errors.New("public key in authorized keys format has trailing new line when none was specified")
			}
		case UnixNewLine:
			if publicKeyAk[len(publicKeyAk) - 1] != '\n' {
				return errors.New("public key in authorized keys format does not have unix new line when unix was specified")
			}
			if string(publicKeyAk[len(publicKeyAk) - 2:]) == WindowsNewLine.String() {
				return errors.New("public key in authorized keys format has windows new line when unix was specified")
			}
		case WindowsNewLine:
			if string(publicKeyAk[len(publicKeyAk) - 2:]) != WindowsNewLine.String() {
				return errors.New("public key in authorized keys format does not have windows new line when windows was specified")
			}
		}

		if len(o.name) > 0 {
			if len(publicKeyAk) < len(o.name) {
				return errors.New("public key in authorized keys format is shorter than the key pair's name")
			}

			suffix := []byte{' '}
			suffix = append(suffix, o.name...)
			suffix = append(suffix, nl.Bytes()...)
			if !bytes.HasSuffix(publicKeyAk, suffix) {
				return errors.New("public key in authorized keys format with name does not have name in suffix - got '" +
					string(publicKeyAk) + "'")
			}
		}
	}

	return nil
}

func (o expected) verifyKeyPair(kp KeyPair) error {
	signer, err := gossh.ParsePrivateKey(kp.PrivateKeyPemBlock())
	if err != nil {
		return errors.New("failed to parse private key during verification - " + err.Error())
	}

	signature, err := signer.Sign(rand.Reader, o.data)
	if err != nil {
		return errors.New("failed to sign test data during verification - " + err.Error())
	}

	err = signer.PublicKey().Verify(o.data, signature)
	if err != nil {
		return errors.New("failed to verify test data - " + err.Error())
	}

	return nil
}

func TestDefaultKeyPairBuilder_Build_Default(t *testing.T) {
	kp, err := NewKeyPairBuilder().Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_EcdsaDefault(t *testing.T) {
	kp, err := NewKeyPairBuilder().
		SetType(Ecdsa).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_EcdsaSupportedCurves(t *testing.T) {
	supportedBits := []int{
		521,
		384,
		256,
	}

	for _, bits := range supportedBits {
		kp, err := NewKeyPairBuilder().
			SetType(Ecdsa).
			SetBits(bits).
			Build()
		if err != nil {
			t.Fatal(err.Error())
		}

		err = expected{
			kind: Ecdsa,
			bits: bits,
			desc: strconv.Itoa(bits) + " bit ECDSA",
			data: []byte(uuid.TimeOrderedUUID()),
		}.matches(kp)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestDefaultKeyPairBuilder_Build_RsaDefault(t *testing.T) {
	kp, err := NewKeyPairBuilder().
		SetType(Rsa).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Rsa,
		bits: 4096,
		desc: "4096 bit RSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_NamedEcdsa(t *testing.T) {
	name := uuid.TimeOrderedUUID()

	kp, err := NewKeyPairBuilder().
		SetType(Ecdsa).
		SetName(name).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
		name: name,
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_NamedRsa(t *testing.T) {
	name := uuid.TimeOrderedUUID()

	kp, err := NewKeyPairBuilder().
		SetType(Rsa).
		SetName(name).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Rsa,
		bits: 4096,
		desc: "4096 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
		name: name,
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}
