package openstack_id3

import (
	"crypto/tls"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Username  string `mapstructure:"username"`
	UserId    string `mapstructure:"user_id"`
	Password  string `mapstructure:"password"`
	ApiKey    string `mapstructure:"api_key"`
	Project   string `mapstructure:"project"`
	ProjectId string `mapstructure:"project_id"`
	Provider  string `mapstructure:"provider"`
	RawRegion string `mapstructure:"region"`
	ProxyUrl  string `mapstructure:"proxy_url"`
	TenantId  string `mapstructure:"tenant_id"`
	Insecure  bool   `mapstructure:"insecure"`
	Domain    string `mapstructure:"domain"`
	DomainId  string `mapstructure:"domain_id"`
}

// Auth returns a ProviderClient with an valid token and catalog for access to openstack services,
// or an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (*gophercloud.ProviderClient, error) {
	// Note: Some of the fetched env vars here are a bit odd but needs to be here because of compatibility
	c.Username = common.ChooseString(c.Username, os.Getenv("SDK_USERNAME"), os.Getenv("OS_USERNAME"))
	c.UserId = common.ChooseString(c.UserId, os.Getenv("OS_USER_ID"))
	c.Password = common.ChooseString(c.Password, os.Getenv("SDK_PASSWORD"), os.Getenv("OS_PASSWORD"))
	c.ApiKey = common.ChooseString(c.ApiKey, os.Getenv("SDK_API_KEY"))
	c.Project = common.ChooseString(c.Project, os.Getenv("SDK_PROJECT"), os.Getenv("OS_TENANT_NAME"), os.Getenv("OS_PROJECT_NAME"))
	c.Provider = common.ChooseString(c.Provider, os.Getenv("SDK_PROVIDER"), os.Getenv("OS_AUTH_URL"))
	c.RawRegion = c.Region()
	c.TenantId = common.ChooseString(c.TenantId, c.ProjectId, os.Getenv("OS_TENANT_ID"), os.Getenv("OS_PROJECT_ID"))
	c.Domain = common.ChooseString(c.Domain, os.Getenv("OS_DOMAIN_NAME"))
	c.DomainId = common.ChooseString(c.DomainId, os.Getenv("OS_PROJECT_DOMAIN_ID"), os.Getenv("OS_USER_DOMAIN_ID"))

	authoptions := gophercloud.AuthOptions{
		AllowReauth: true,

		IdentityEndpoint: c.Provider,
		Username:         c.Username,
		UserID:           c.UserId,
		Password:         c.Password,
		APIKey:           c.ApiKey,
		TenantName:       c.Project,
		TenantID:         c.TenantId,
		DomainName:       c.Domain,
		DomainID:         c.DomainId,
	}

	default_transport := &http.Transport{}

	if c.Insecure {
		cfg := new(tls.Config)
		cfg.InsecureSkipVerify = true
		default_transport.TLSClientConfig = cfg
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
		default_transport.Proxy = http.ProxyURL(url)
	}

	if c.Insecure || c.ProxyUrl != "" {
		http.DefaultTransport = default_transport
	}

	// Creates the provider client that will contain token and catalog
	provider, err := openstack.AuthenticatedClient(authoptions)
	if err != nil {
		return nil, err
	}
	// Attempts to autheticate and update provider with token and catalog
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
		"username":   &c.Username,
		"user_id":    &c.UserId,
		"password":   &c.Password,
		"api_key":    &c.ApiKey,
		"provider":   &c.Provider,
		"project":    &c.Project,
		"project_id": &c.ProjectId,
		"tenant_id":  &c.TenantId,
		"domain":     &c.Domain,
		"domain_id":  &c.DomainId,
		"region":     &c.RawRegion,
		"proxy_url":  &c.ProxyUrl,
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
