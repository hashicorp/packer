package openstack

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"net/http"
	"net/url"
	"os"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Project   string `mapstructure:"project"`
	Provider  string `mapstructure:"provider"`
	RawRegion string `mapstructure:"region"`
	ProxyUrl  string `mapstructure:"proxy_url"`
}

// Auth returns a valid Auth object for access to openstack services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (gophercloud.AccessProvider, error) {
	c.Username = common.CoalesceVals(c.Username, os.Getenv("SDK_USERNAME"), os.Getenv("OS_USERNAME"))
	c.Password = common.CoalesceVals(c.Password, os.Getenv("SDK_PASSWORD"), os.Getenv("OS_PASSWORD"))
	c.Project = common.CoalesceVals(c.Project, os.Getenv("SDK_PROJECT"), os.Getenv("OS_TENANT_NAME"))
	c.Provider = common.CoalesceVals(c.Provider, os.Getenv("SDK_PROVIDER"), os.Getenv("OS_AUTH_URL"))
	c.RawRegion = common.CoalesceVals(c.RawRegion, os.Getenv("SDK_REGION"), os.Getenv("OS_REGION_NAME"))

	authoptions := gophercloud.AuthOptions{
		Username:    c.Username,
		Password:    c.Password,
		AllowReauth: true,
	}

	if c.Project != "" {
		authoptions.TenantName = c.Project
	}

	// For corporate networks it may be the case where we want our API calls
	// to be sent through a separate HTTP proxy than external traffic.
	if c.ProxyUrl != "" {
		url, err := url.Parse(c.ProxyUrl)
		if err != nil {
			return nil, err
		}

		// The gophercloud.Context has a UseCustomClient method which
		// would allow us to override with a new instance of http.Client.
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(url)}
	}

	return gophercloud.Authenticate(c.Provider, authoptions)
}

func (c *AccessConfig) Region() string {
	return common.CoalesceVals(c.RawRegion, os.Getenv("SDK_REGION"), os.Getenv("OS_REGION_NAME"))
}

func (c *AccessConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	templates := map[string]*string{
		"username": &c.Username,
		"password": &c.Password,
		"provider": &c.Provider,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.RawRegion == "" {
		errs = append(errs, fmt.Errorf("region must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
