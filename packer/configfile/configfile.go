package configfile

import (
	"os"
	"path/filepath"
)

const EnvPackerTmpDir = "PACKER_TMP_DIR"

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

// ConfigTmpDir returns a "deterministic" (based on environment or Packer config)
// path intended as the root of subsequent temporary items, to minimize scatter.
//
// The caller must ensure safe tempfile practice via ioutil.TempDir() and friends.
func ConfigTmpDir() (string, error) {
	var tmpdir, td string
	var err error

	if tmpdir = os.Getenv(EnvPackerTmpDir); tmpdir != "" {
		td, err = filepath.Abs(tmpdir)
	} else if tmpdir, err = configDir(); err == nil {
		td = filepath.Join(tmpdir, "tmp")
	} else if tmpdir = os.TempDir(); tmpdir != "" {
		td = filepath.Join(tmpdir, "packer")
	}

	if _, err = os.Stat(td); os.IsNotExist(err) {
		err = os.MkdirAll(td, 0700)
	}
	if err != nil {
		return "", err
	}

	return td, nil
}
