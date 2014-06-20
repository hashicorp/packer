package vagrantcloud

import (
	"fmt"
)

type Provider struct {
	client *VagrantCloudClient
	Name   string `json:"name"`
}

// https://vagrantcloud.com/docs/providers
func (v VagrantCloudClient) Provider(tag string) (*Box, error) {
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

// Save persist the box over HTTP to Vagrant Cloud
func (p Provider) Save(name string) (bool, error) {
	resp, err := p.client.Get(name)

	if err != nil {
		return false, fmt.Errorf("Error retrieving box: %s", err)
	}

	provider := &Provider{}

	if err = decodeBody(resp, provider); err != nil {
		return false, fmt.Errorf("Error parsing box response: %s", err)
	}

	return true, nil
}
