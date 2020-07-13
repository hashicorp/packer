package extensions

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// Get retrieves information for a specific extension using its alias.
func Get(c *gophercloud.ServiceClient, alias string) (r GetResult) {
	resp, err := c.Get(ExtensionURL(c, alias), &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// List returns a Pager which allows you to iterate over the full collection of extensions.
// It does not accept query parameters.
func List(c *gophercloud.ServiceClient) pagination.Pager {
	return pagination.NewPager(c, ListExtensionURL(c), func(r pagination.PageResult) pagination.Page {
		return ExtensionPage{pagination.SinglePageBase(r)}
	})
}
