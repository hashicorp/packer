package packer

import (
	"os"
	"path/filepath"
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
	var tmpdir, td string
	var found bool

	if tmpdir = os.Getenv("PACKER_TMP_DIR"); tmpdir == "" {
		for e := range []string{"TEMP", "TMP", "LOCALAPPDATA"} {
			if tmpdir, found := os.LookupEnv(e); found {
				td = filepath.Join(tmpdir, "packer")
				break
			}
		}
	}
	if tmpdir == "" {
		td = filepath.Join(configDir(), "tmp")
	}

	_, err = os.Stat(td)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(td, 0700); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return td, nil
}
