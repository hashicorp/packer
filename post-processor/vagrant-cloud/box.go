package vagrantcloud

import (
	"fmt"
)

type Box struct {
	client *VagrantCloudClient
	Tag    string `json:"tag"`
}

// https://vagrantcloud.com/docs/boxes
func (v VagrantCloudClient) Box(tag string) (*Box, error) {
	resp, err := v.Get(tag)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving box: %s", err)
	}

	box := &Box{}

	if err = decodeBody(resp, box); err != nil {
		return nil, fmt.Errorf("Error parsing box response: %s", err)
	}

	return box, nil
}

// Save persist the provider over HTTP to Vagrant Cloud
func (b Box) Save(tag string) (bool, error) {
	resp, err := b.client.Get(tag)

	if err != nil {
		return false, fmt.Errorf("Error retrieving box: %s", err)
	}

	box := &Box{}

	if err = decodeBody(resp, box); err != nil {
		return false, fmt.Errorf("Error parsing box response: %s", err)
	}

	return true, nil
}
