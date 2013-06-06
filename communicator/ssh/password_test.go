package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"testing"
)

func TestPassword_Impl(t *testing.T) {
	var raw interface{}
	raw = Password("foo")
	if _, ok := raw.(ssh.ClientPassword); !ok {
		t.Fatal("Password must implement ClientPassword")
	}
}

func TestPasswordPassword(t *testing.T) {
	p := Password("foo")
	result, err := p.Password("user")
	if err != nil {
		t.Fatalf("err not nil: %s", err)
	}

	if result != "foo" {
		t.Fatalf("invalid password: %s", result)
	}
}
