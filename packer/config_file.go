package packer

import (
	"io/ioutil"
	"log"
	"os"
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
	var tmpdir, td, cd string
	var err error

	cd, _ = ConfigDir()
	for _, tmpdir = range []string{os.Getenv("PACKER_TMP_DIR"), os.TempDir(), cd} {
		if tmpdir != "" {
			break
		}
	}

	if td, err = ioutil.TempDir(tmpdir, "packer"); err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(td)

	return td, nil
}
