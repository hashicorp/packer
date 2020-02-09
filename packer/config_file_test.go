package packer

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
)

// Depending on the platform, find a valid username to use
func platform_user() string {
	// XXX: We make an assumption here that there's an Administrator user
	//      on the windows platform, whereas the correct way is to use
	//		the api or to scrape `net user`.
	if runtime.GOOS == "windows" {
		return "Administrator"
	}
	return "root"
}

func homedir_current() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}

func homedir_user(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}

// Begin the actual tests and stuff
func TestExpandUser_Empty(t *testing.T) {
	var path, expected string

	// Try an invalid user
	_, err := ExpandUser("~invalid-user-that-should-not-exist")
	if err == nil {
		t.Fatalf("expected failure")
	}

	// Try an empty string
	expected = ""
	if path, err = ExpandUser(""); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try an absolute path
	expected = "/etc/shadow"
	if path, err = ExpandUser("/etc/shadow"); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try a relative path
	expected = "tmp/foo"
	if path, err = ExpandUser("tmp/foo"); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}
}

func TestExpandUser_Current(t *testing.T) {
	var path, expected string

	// Grab the current user's home directory to verify ExpandUser works
	homedir, err := homedir_current()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Try just a tilde
	expected = homedir
	if path, err = ExpandUser("~"); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try as a directory
	expected = filepath.Join(homedir, "")
	if path, err = ExpandUser("~/"); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try as a file
	expected = filepath.Join(homedir, "foo")
	if path, err = ExpandUser("~/foo"); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}
}

func TestExpandUser_User(t *testing.T) {
	var path, expected string

	username := platform_user()

	// Grab the current user's home directory to verify ExpandUser works
	homedir, err := homedir_user(username)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Try just a tilde
	expected = homedir
	if path, err = ExpandUser(fmt.Sprintf("~%s", username)); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try as a directory
	expected = filepath.Join(homedir, "")
	if path, err = ExpandUser(fmt.Sprintf("~%s/", username)); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}

	// Try as a file
	expected = filepath.Join(homedir, "foo")
	if path, err = ExpandUser(fmt.Sprintf("~%s/foo", username)); err != nil {
		t.Fatalf("err: %s", err)
	}

	if path != expected {
		t.Fatalf("err: %v != %v", path, expected)
	}
}
