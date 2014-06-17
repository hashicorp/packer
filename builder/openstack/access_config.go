package openstack

import (
	"crypto/tls"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	ApiKey    string `mapstructure:"api_key"`
	Project   string `mapstructure:"project"`
	Provider  string `mapstructure:"provider"`
	RawRegion string `mapstructure:"region"`
	ProxyUrl  string `mapstructure:"proxy_url"`
	TenantId  string `mapstructure:"tenant_id"`
	Insecure  bool   `mapstructure:"insecure"`
}

// Auth returns a valid Auth object for access to openstack services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (gophercloud.AccessProvider, error) {
	c.Username = common.ChooseString(c.Username, os.Getenv("SDK_USERNAME"), os.Getenv("OS_USERNAME"))
	c.Password = common.ChooseString(c.Password, os.Getenv("SDK_PASSWORD"), os.Getenv("OS_PASSWORD"))
	c.ApiKey = common.ChooseString(c.ApiKey, os.Getenv("SDK_API_KEY"))
	c.Project = common.ChooseString(c.Project, os.Getenv("SDK_PROJECT"), os.Getenv("OS_TENANT_NAME"))
	c.Provider = common.ChooseString(c.Provider, os.Getenv("SDK_PROVIDER"), os.Getenv("OS_AUTH_URL"))
	c.RawRegion = common.ChooseString(c.RawRegion, os.Getenv("SDK_REGION"), os.Getenv("OS_REGION_NAME"))
	c.TenantId = common.ChooseString(c.TenantId, os.Getenv("OS_TENANT_ID"))

	// OpenStack's auto-generated openrc.sh files do not append the suffix
	// /tokens to the authentication URL. This ensures it is present when
	// specifying the URL.
	if strings.Contains(c.Provider, "://") && !strings.HasSuffix(c.Provider, "/tokens") {
		c.Provider += "/tokens"
	}

	authoptions := gophercloud.AuthOptions{
		AllowReauth: true,

		ApiKey:     c.ApiKey,
		TenantId:   c.TenantId,
		TenantName: c.Project,
		Username:   c.Username,
		Password:   c.Password,
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

	return gophercloud.Authenticate(c.Provider, authoptions)
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
		"api_key":   &c.ApiKey,
		"provider":  &c.Provider,
		"project":   &c.Project,
		"tenant_id": &c.TenantId,
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

	if c.Region() == "" {
		errs = append(errs, fmt.Errorf("region must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
