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

// ConfigTmpDir returns the configuration tmp directory for Packer
func ConfigTmpDir() (string, error) {
	if tmpdir := os.Getenv("PACKER_TMP_DIR"); tmpdir != "" {
		return filepath.Abs(tmpdir)
	}
	configdir, err := configDir()
	if err != nil {
		return "", err
	}
	td := filepath.Join(configdir, "tmp")
	_, err = os.Stat(td)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(td, 0755); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return td, nil
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

func configFile() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, defaultConfigFile), nil
}

func configDir() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, defaultConfigDir), nil
}

// Given a path, check to see if it's using ~ to reference a user directory.
// If so, then replace that component with the requrested user's directory.
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
