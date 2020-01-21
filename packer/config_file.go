package packer

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ConfigFile returns the default path to the configuration file. On
// Unix-like systems this is the ".packerconfig" file in the home directory.
// On Windows, this is the "packer.config" file in the application data
// directory.
func ConfigFile() (string, error) {
	return configFile()
}

// ConfigDir returns the configuration directory for Packer.
func ConfigDir() (string, error) {
	return configDir()
}

func homeDir() (string, error) {
	// Prefer $HOME over user.Current due to glibc bug: golang.org/issue/13470
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if home := os.Getenv("APPDATA"); home != "" {
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

func configFile() (string, error) {
	var dir string
	if cd := os.Getenv("PACKER_CONFIG_DIR"); cd != "" {
		log.Printf("Detected config directory from env var: %s", cd)
		dir = cd
	} else {
		homedir, err := homeDir()
		if err != nil {
			return "", err
		}
		dir = homedir
	}
	return filepath.Join(dir, defaultConfigFile), nil
}

func configDir() (string, error) {
	var dir string
	if cd := os.Getenv("PACKER_CONFIG_DIR"); cd != "" {
		log.Printf("Detected config directory from env var: %s", cd)
		dir = cd
	} else {
		homedir, err := homeDir()
		if err != nil {
			return "", err
		}
		dir = homedir
	}

	return filepath.Join(dir, defaultConfigDir), nil
}

// Given a path, check to see if it's using ~ to reference a user directory.
// If so, then replace that component with the requested user directory.
// In "~/", "~" gets replaced by current user's home dir.
// In "~root/", "~user" gets replaced by root's home dir.
// ~ has to be the first character of path for ExpandUser change it.
func ExpandUser(path string) (string, error) {
	var (
		u   *user.User
		err error
	)

	// refuse to do anything with a zero-length path
	if len(path) == 0 {
		return path, nil
	}

	// If no expansion was specified, then refuse that too
	if path[0] != '~' {
		return path, nil
	}

	// Grab everything up to the first filepath.Separator
	idx := strings.IndexAny(path, `/\`)
	if idx == -1 {
		idx = len(path)
	}

	// Now we should be able to extract the username
	username := path[:idx]

	// Check if the current user was requested
	if username == "~" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(username[1:])
	}

	// If we couldn't figure that out, then fail here
	if err != nil {
		return "", err
	}

	// Now we can replace the path with u.HomeDir
	return filepath.Join(u.HomeDir, path[idx:]), nil
}
