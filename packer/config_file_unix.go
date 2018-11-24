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
	var u *user.User

	/// First prefer the HOME environmental variable
	if home, ok := os.LookupEnv("HOME"); ok {
		log.Printf("Detected home directory from env var: %s", home)
		return home, nil
	}

	/// Fall back to the passwd database if not found which follows
	/// the same semantics as bourne shell
	var err error

	// Check username specified in the environment first
	if username, ok := os.LookupEnv("USER"); ok {
		u, err = user.Lookup(username)

	} else {
		// Otherwise we assume the current user
		u, err = user.Current()
	}

	// Fail if we were unable to read the record
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}
