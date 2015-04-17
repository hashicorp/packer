package openstack_id3

import (
	//"crypto/tls"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	//"net/http"
	//"net/url"
	"os"
	"strings"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Domain	  string `mapstructure:"domain"`
	DomainID  string `mapstructure:"domain_id"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Project   string `mapstructure:"project"`
	ProjectID string `mapstructure:"project_id"`
	Provider  string `mapstructure:"provider"`
	RawRegion string `mapstructure:"region"`
	ProxyUrl  string `mapstructure:"proxy_url"`
	Insecure  bool   `mapstructure:"insecure"`
}

// Auth returns a ProviderClient with an valid token and catalog for access to openstack services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (*gophercloud.ProviderClient, error) {
	c.Domain = common.ChooseString(c.Domain, 	   os.Getenv("OS_USER_DOMAIN_NAME"))
	c.DomainID = common.ChooseString(c.DomainID,   os.Getenv("OS_USER_DOMAIN_ID"), os.Getenv("OS_PROJECT_DOMAIN_ID"))
	
	c.Username = common.ChooseString(c.Username,   os.Getenv("OS_USERNAME"))
	c.Password = common.ChooseString(c.Password,   os.Getenv("OS_PASSWORD"))
	
	c.Project = common.ChooseString(c.Project, 	   os.Getenv("OS_PROJECT_NAME"))
	c.ProjectID = common.ChooseString(c.ProjectID,  os.Getenv("OS_PROJECT_ID"))
	
	c.Provider = common.ChooseString(c.Provider,   os.Getenv("OS_AUTH_URL"))
	c.RawRegion = common.ChooseString(c.RawRegion, os.Getenv("OS_REGION_NAME"))

	authoptions := gophercloud.AuthOptions{
		IdentityEndpoint: c.Provider,
		DomainName: c.Domain,
		Username: c.Username,
		Password: c.Password,
		TenantName: c.Project,
		AllowReauth: true,
	}
	// Creates the provider empty client that will contain token and catalog
	provider, err := openstack.AuthenticatedClient(authoptions)
	if err != nil {
		return nil, err
	}
	// Attempts to autheticate and update provider with token and catatlog 
	err = openstack.Authenticate(provider, authoptions)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (c *AccessConfig) Region() string {
	return common.ChooseString(c.RawRegion, os.Getenv("SDK_REGION"), os.Getenv("OS_REGION_NAME"))
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
		"username":  &c.Username,
		"password":  &c.Password,
		"provider":  &c.Provider,
		"project":   &c.Project,
		"region":    &c.RawRegion,
		"proxy_url": &c.ProxyUrl,
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

	if strings.HasPrefix(c.Provider, "rackspace") {
		if c.Region() == "" {
			errs = append(errs, fmt.Errorf("region must be specified when using rackspace"))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
