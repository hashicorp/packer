package clientconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

// findAndLoadYAML attempts to locate a clouds.yaml file in the following
// locations:
//
// 1. OS_CLIENT_CONFIG_FILE
// 2. Current directory.
// 3. unix-specific user_config_dir (~/.config/openstack/clouds.yaml)
// 4. unix-specific site_config_dir (/etc/openstack/clouds.yaml)
//
// If found, the contents of the file is returned.
func findAndReadYAML() ([]byte, error) {
	// OS_CLIENT_CONFIG_FILE
	if v := os.Getenv("OS_CLIENT_CONFIG_FILE"); v != "" {
		if ok := fileExists(v); ok {
			return ioutil.ReadFile(v)
		}
	}

	// current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to determine working directory: %s", err)
	}

	filename := filepath.Join(cwd, "clouds.yaml")
	if ok := fileExists(filename); ok {
		return ioutil.ReadFile(filename)
	}

	// unix user config directory: ~/.config/openstack.
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %s", err)
	}

	homeDir := currentUser.HomeDir
	if homeDir != "" {
		filename := filepath.Join(homeDir, ".config/openstack/clouds.yaml")
		if ok := fileExists(filename); ok {
			return ioutil.ReadFile(filename)
		}
	}

	// unix-specific site config directory: /etc/openstack.
	if ok := fileExists("/etc/openstack/clouds.yaml"); ok {
		return ioutil.ReadFile("/etc/openstack/clouds.yaml")
	}

	return nil, fmt.Errorf("no clouds.yaml file found")
}

// fileExists checks for the existence of a file at a given location.
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
