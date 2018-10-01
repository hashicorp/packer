package googlecompute

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"
)

func testPrivateKeyFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	defer tf.Close()

	b := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("what"),
	}

	if err := pem.Encode(tf, b); err != nil {
		t.Fatalf("err: %s", err)
	}

	return tf.Name()
}

func TestProcesssPrivateKeyFile(t *testing.T) {
	path := testPrivateKeyFile(t)
	defer os.Remove(path)

	data, err := processPrivateKeyFile(path, "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(data) <= 0 {
		t.Fatalf("bad: %#v", data)
	}
}

func TestProcessPrivateKeyFile_encrypted(t *testing.T) {
	data := []byte("what")
	// Encrypt the file
	b, err := x509.EncryptPEMBlock(rand.Reader,
		"RSA PRIVATE KEY",
		data,
		[]byte("password"),
		x509.PEMCipherAES128)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	defer os.Remove(tf.Name())

	err = pem.Encode(tf, b)
	tf.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	path := tf.Name()

	// Should have an error with a bad password
	if b, err := processPrivateKeyFile(path, "bad"); err == nil {
		if string(b) == string(data) {
			t.Fatal("should error & be different")
		}
		t.Logf(`Decrypt was successfull but the body was wrong.`)
		// Because of deficiencies
		// in the encrypted-PEM format, it's not always possible to detect an incorrect
		// password. In these cases no error will be returned but the decrypted DER
		// bytes will be random noise.
		// https://github.com/golang/go/blob/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/src/crypto/x509/pem_decrypt.go#L112-L114
	}

	if _, err := processPrivateKeyFile(path, "password"); err != nil {
		t.Fatalf("bad: %s", err)
	}
}
