package vagrantcloud

import (
	"fmt"
)

type Version struct {
	client  *VagrantCloudClient
	Version string `json:"version"`
	Number  string `json:"number"`
}

// https://vagrantcloud.com/docs/versions
func (v VagrantCloudClient) Version(number string) (*Version, error) {
	resp, err := v.Get(number)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving version: %s", err)
	}

	version := &Version{}

	if err = decodeBody(resp, version); err != nil {
		return nil, fmt.Errorf("Error parsing version response: %s", err)
	}

	return version, nil
}

// Save persists the Version over HTTP to Vagrant Cloud
func (v Version) Create() (bool, error) {
	resp, err := v.client.Post(v.Number, v)

	if err != nil {
		return false, fmt.Errorf("Error retrieving box: %s", err)
	}

	if err = decodeBody(resp, v); err != nil {
		return false, fmt.Errorf("Error parsing box response: %s", err)
	}

	return true, nil
}

// Deletes the Version over HTTP to Vagrant Cloud
func (v Version) Destroy() (bool, error) {
	resp, err := v.client.Delete(v.Number)

	if err != nil {
		return false, fmt.Errorf("Error destroying version: %s", err)
	}

	if err = decodeBody(resp, v); err != nil {
		return false, fmt.Errorf("Error parsing box response: %s", err)
	}

	return true, nil
}
