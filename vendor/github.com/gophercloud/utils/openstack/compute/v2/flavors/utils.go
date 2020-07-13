package flavors

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
)

// IDFromName is a convienience function that returns a flavor's ID given its
// name.
func IDFromName(client *gophercloud.ServiceClient, name string) (string, error) {
	count := 0
	id := ""
	allPages, err := flavors.ListDetail(client, nil).AllPages()
	if err != nil {
		return "", err
	}

	all, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", err
	}

	for _, f := range all {
		if f.Name == name {
			count++
			id = f.ID
		}
	}

	switch count {
	case 0:
		err := &gophercloud.ErrResourceNotFound{}
		err.ResourceType = "flavor"
		err.Name = name
		return "", err
	case 1:
		return id, nil
	default:
		err := &gophercloud.ErrMultipleResourcesFound{}
		err.ResourceType = "flavor"
		err.Name = name
		err.Count = count
		return "", err
	}
}
