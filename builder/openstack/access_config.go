package openstack

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"os"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Project   string `mapstructure:"project"`
	Provider  string `mapstructure:"provider"`
	RawRegion string `mapstructure:"region"`
}

// Auth returns a valid Auth object for access to openstack services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (gophercloud.AccessProvider, error) {
	username := c.Username
	password := c.Password
	project := c.Project
	provider := c.Provider

	if username == "" {
		username = os.Getenv("SDK_USERNAME")
	}
	if password == "" {
		password = os.Getenv("SDK_PASSWORD")
	}
	if project == "" {
		project = os.Getenv("SDK_PROJECT")
	}
	if provider == "" {
		provider = os.Getenv("SDK_PROVIDER")
	}

	authoptions := gophercloud.AuthOptions{
		Username:    username,
		Password:    password,
		AllowReauth: true,
	}

	if project != "" {
		authoptions.TenantName = project
	}

	return gophercloud.Authenticate(provider, authoptions)
}

func (c *AccessConfig) Region() string {
	return c.RawRegion
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
