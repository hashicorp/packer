package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// This is used in the BasicPlaceholderData() func in the packer/provisioner.go
// To force users to access generated data via the "generated" func.
const PlaceholderMsg = "To set this dynamically in the Packer template, " +
	"you must use the `build` function"

// Used to set variables which we need to access later in the build, where
// state bag and config information won't work
func sharedStateFilename(suffix string, buildName string) string {
	uuid := os.Getenv("PACKER_RUN_UUID")
	return filepath.Join(os.TempDir(), fmt.Sprintf("packer-%s-%s-%s", uuid, suffix, buildName))
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
