package sshkey

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/ssh"
)

func TestGeneratePair_parseable(t *testing.T) {
	tests := []struct {
		t Algorithm
	}{
		{DSA},
		{RSA},
		{ECDSA},
		{ED25519},
	}
	for _, tt := range tests {
		t.Run(tt.t.String(), func(t *testing.T) {
			got, err := GeneratePair(tt.t, nil, 0)
			if err != nil {
				t.Errorf("GeneratePair() error = %v", err)
				return
			}

			privateKey, err := ssh.ParsePrivateKey(got.Private)
			if err != nil {
				t.Fatal(err)
			}
			publicKey, _, _, _, err := ssh.ParseAuthorizedKey(got.Public)
			if err != nil {
				t.Fatalf("%v: %s", err, got.Public)
			}
			if diff := cmp.Diff(privateKey.PublicKey().Marshal(), publicKey.Marshal()); diff != "" {
				t.Fatalf("wrong public key: %s", diff)
			}
		})
	}
}
