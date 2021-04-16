// Package testutils provides some simple ease-of-use tools for implementing
// acceptance testing.
package testutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// CleanupFiles removes all the provided filenames.
func CleanupFiles(moreFiles ...string) {
	for _, file := range moreFiles {
		os.RemoveAll(file)
	}
}

// FileExists returns true if the filename is found.
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// Use the manifest to load information about artifact ID and region.
// to properly use this, you should make sure the manifestfilepath is unique
// for each test, and remember to clean up the manifest file when your build
// is done!
func GetArtifact(manifestfilepath string) (ManifestFile, error) {
	// example manifest.json
	// {
	//   "builds": [
	//     {
	//       "name": "test",
	//       "builder_type": "alicloud-ecs",
	//       "build_time": 1618424957,
	//       "files": null,
	//       "artifact_id": "us-east-1:m-0xi15a442knfbtmnymm9",
	//       "packer_run_uuid": "81fc083f-0b78-d815-ed3a-2e5f53b36bff",
	//       "custom_data": null
	//     }
	//   ],
	//   "last_run_uuid": "81fc083f-0b78-d815-ed3a-2e5f53b36bff"
	// }
	manifest := ManifestFile{}
	data, err := ioutil.ReadFile(manifestfilepath)
	if err != nil {
		return manifest, fmt.Errorf("failed to open manifest file %s", manifestfilepath)
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("Failed to decode manifest: %v", err)
	}

	return manifest, nil
}

// "A little copying is better than a lot of dependecy"
// This code comes from the manifest post-processor shipped with Packer core.
// These structs allow us to re-decode the manifest
type ManifestFile struct {
	Builds      []ManifestArtifact `json:"builds"`
	LastRunUUID string             `json:"last_run_uuid"`
}

type ManifestArtifactFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ManifestArtifact struct {
	BuildName     string                 `json:"name"`
	BuilderType   string                 `json:"builder_type"`
	BuildTime     int64                  `json:"build_time,omitempty"`
	ArtifactFiles []ManifestArtifactFile `json:"files"`
	ArtifactId    string                 `json:"artifact_id"`
	PackerRunUUID string                 `json:"packer_run_uuid"`
	CustomData    map[string]string      `json:"custom_data"`
}
