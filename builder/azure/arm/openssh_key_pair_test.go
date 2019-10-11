package arm

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestFart(t *testing.T) {

}

func TestAuthorizedKeyShouldParse(t *testing.T) {
	testSubject, err := NewOpenSshKeyPairWithSize(512)
	if err != nil {
		t.Fatalf("Failed to create a new OpenSSH key pair, err=%s.", err)
	}

	authorizedKey := testSubject.AuthorizedKey()

	_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(authorizedKey))
	if err != nil {
		t.Fatalf("Failed to parse the authorized key, err=%s", err)
	}
}

func TestPrivateKeyShouldParse(t *testing.T) {
	testSubject, err := NewOpenSshKeyPairWithSize(512)
	if err != nil {
		t.Fatalf("Failed to create a new OpenSSH key pair, err=%s.", err)
	}

	_, err = ssh.ParsePrivateKey([]byte(testSubject.PrivateKey()))
	if err != nil {
		t.Fatalf("Failed to parse the private key, err=%s\n", err)
	}
}
