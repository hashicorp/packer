package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Used to set variables which we need to access later in the build, where
// state bag and config information won't work
func sharedStateFilename(suffix string) string {
	uuid := os.Getenv("PACKER_RUN_UUID")
	return filepath.Join(os.TempDir(), fmt.Sprintf("packer-%s-%s", uuid, suffix))
}

func SetSharedState(key string, value string) error {
	return ioutil.WriteFile(sharedStateFilename(key), []byte(value), 0600)
}

func RetrieveSharedState(key string) (string, error) {
	value, err := ioutil.ReadFile(sharedStateFilename(key))
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func RemoveSharedStateFile(key string) {
	os.Remove(sharedStateFilename(key))
}
