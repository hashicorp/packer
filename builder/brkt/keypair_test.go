package brkt

import (
	"testing"
)

func TestNewRandomKeypair_Run(t *testing.T) {
	keypair, err := NewRandomKeyPair()

	if err != nil {
		t.Fatalf("generating a KeyPair should not return an err")
	}
	if keypair.PublicKey == "" {
		t.Fatalf("KeyPair#PublicKey should not be empty")
	}
	if keypair.PrivateKey == "" {
		t.Fatalf("KeyPair#PrivateKey should not be empty")
	}
}
