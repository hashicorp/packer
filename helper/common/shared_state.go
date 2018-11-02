package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer/configfile"
)

var sharedStateDir string

// Used to set variables which we need to access later in the build, where
// state bag and config information won't work
func sharedStateFilename(suffix string, buildName string) string {
	var uuid string

	if sharedStateDir == "" {
		prefix, _ := configfile.ConfigTmpDir()
		sharedStateDir, err := ioutil.TempDir(prefix, "state")
		if err != nil {
			return ""
		}
		defer os.RemoveAll(sharedStateDir)
	}

	uuid = os.Getenv("PACKER_RUN_UUID")
	if uuid == "" {
		uuid = "none"
	}
	return filepath.Join(sharedStateDir, fmt.Sprintf("%s-%s-%s", uuid, suffix, buildName))
}

func SetSharedState(key string, value string, buildName string) error {
	return ioutil.WriteFile(sharedStateFilename(key, buildName), []byte(value), 0600)
}

func RetrieveSharedState(key string, buildName string) (string, error) {
	value, err := ioutil.ReadFile(sharedStateFilename(key, buildName))
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func RemoveSharedStateFile(key string, buildName string) {
	os.Remove(sharedStateFilename(key, buildName))
}
