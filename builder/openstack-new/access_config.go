package openstack

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/mitchellh/packer/template/interpolate"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	Username         string `mapstructure:"username"`
	UserID           string `mapstructure:"user_id"`
	Password         string `mapstructure:"password"`
	APIKey           string `mapstructure:"api_key"`
	IdentityEndpoint string `mapstructure:"identity_endpoint"`
	TenantID         string `mapstructure:"tenant_id"`
	TenantName       string `mapstructure:"tenant_name"`
	DomainID         string `mapstructure:"domain_id"`
	DomainName       string `mapstructure:"domain_name"`
	Insecure         bool   `mapstructure:"insecure"`
	Region           string `mapstructure:"region"`
	EndpointType     string `mapstructure:"endpoint_type"`

	osClient *gophercloud.ProviderClient
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	if c.EndpointType != "internal" && c.EndpointType != "internalURL" &&
		c.EndpointType != "admin" && c.EndpointType != "adminURL" &&
		c.EndpointType != "public" && c.EndpointType != "publicURL" &&
		c.EndpointType != "" {
		return []error{fmt.Errorf("Invalid endpoint type provided")}
	}

	if c.Region == "" {
		c.Region = os.Getenv("OS_REGION_NAME")
	}

	// Get as much as possible from the end
	ao, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return []error{err}
	}

	// Override values if we have them in our config
	overrides := []struct {
		From, To *string
	}{
		{&c.Username, &ao.Username},
		{&c.UserID, &ao.UserID},
		{&c.Password, &ao.Password},
		{&c.APIKey, &ao.APIKey},
		{&c.IdentityEndpoint, &ao.IdentityEndpoint},
		{&c.TenantID, &ao.TenantID},
		{&c.TenantName, &ao.TenantName},
		{&c.DomainID, &ao.DomainID},
		{&c.DomainName, &ao.DomainName},
	}
	for _, s := range overrides {
		if *s.From != "" {
			*s.To = *s.From
		}
	}

	// Build the client itself
	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return []error{err}
	}

	// If we have insecure set, then create a custom HTTP client that
	// ignores SSL errors.
	if c.Insecure {
		config := &tls.Config{InsecureSkipVerify: true}
		transport := &http.Transport{TLSClientConfig: config}
		client.HTTPClient.Transport = transport
	}

	// Auth
	err = openstack.Authenticate(client, ao)
	if err != nil {
		return []error{err}
	}

	c.osClient = client
	return nil
}

func (c *AccessConfig) computeV2Client() (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(c.osClient, gophercloud.EndpointOpts{
		Region:       c.Region,
		Availability: c.getEndpointType(),
	})
}

func (c *AccessConfig) getEndpointType() gophercloud.Availability {
	if c.EndpointType == "internal" || c.EndpointType == "internalURL" {
		return gophercloud.AvailabilityInternal
	}
	if c.EndpointType == "admin" || c.EndpointType == "adminURL" {
		return gophercloud.AvailabilityAdmin
	}
	return gophercloud.AvailabilityPublic
}
