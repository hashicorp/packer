// +build darwin freebsd linux netbsd openbsd solaris

package packer

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func configFile() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ".packerconfig"), nil
}

func configDir() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ".packer.d"), nil
}

func homeDir() (string, error) {

	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		log.Printf("Detected home directory from env var: %s", home)
		return home, nil
	}

	// Fall back to the passwd database if not found which follows
	// the same semantics as bourne shell
	u, err := user.Current()

	// Get homedir from specified username
	// if it is set and different than what we have
	if username := os.Getenv("USER"); username != "" && err == nil && u.Username != username {
		u, err = user.Lookup(username)
	}

	// Fail if we were unable to read the record
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}
