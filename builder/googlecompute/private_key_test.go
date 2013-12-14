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
	// Encrypt the file
	b, err := x509.EncryptPEMBlock(rand.Reader,
		"RSA PRIVATE KEY",
		[]byte("what"),
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
	if _, err := processPrivateKeyFile(path, "bad"); err == nil {
		t.Fatal("should error")
	}

	if _, err := processPrivateKeyFile(path, "password"); err != nil {
		t.Fatalf("bad: %s", err)
	}
}
