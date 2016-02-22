package vips

import "github.com/rackspace/gophercloud"

const (
	rootPath     = "lb"
	resourcePath = "vips"
)

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(rootPath, resourcePath)
}

func resourceURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, resourcePath, id)
}
