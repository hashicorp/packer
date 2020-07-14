package startstop

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
)

// Start is the operation responsible for starting a Compute server.
func Start(client *gophercloud.ServiceClient, id string) (r StartResult) {
	resp, err := client.Post(extensions.ActionURL(client, id), map[string]interface{}{"os-start": nil}, nil, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Stop is the operation responsible for stopping a Compute server.
func Stop(client *gophercloud.ServiceClient, id string) (r StopResult) {
	resp, err := client.Post(extensions.ActionURL(client, id), map[string]interface{}{"os-stop": nil}, nil, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
