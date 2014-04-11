package ssh

import (
	"code.google.com/p/gosshold/ssh"
	"reflect"
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

func TestPasswordKeyboardInteractive_Impl(t *testing.T) {
	var raw interface{}
	raw = PasswordKeyboardInteractive("foo")
	if _, ok := raw.(ssh.ClientKeyboardInteractive); !ok {
		t.Fatal("PasswordKeyboardInteractive must implement ClientKeyboardInteractive")
	}
}

func TestPasswordKeybardInteractive_Challenge(t *testing.T) {
	p := PasswordKeyboardInteractive("foo")
	result, err := p.Challenge("foo", "bar", []string{"one", "two"}, nil)
	if err != nil {
		t.Fatalf("err not nil: %s", err)
	}

	if !reflect.DeepEqual(result, []string{"foo", "foo"}) {
		t.Fatalf("invalid password: %#v", result)
	}
}
