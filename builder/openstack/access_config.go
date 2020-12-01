//go:generate struct-markdown

package openstack

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/hashicorp/go-cleanhttp"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// AccessConfig is for common configuration related to openstack access
type AccessConfig struct {
	// The username or id used to connect to the OpenStack service. If not
	// specified, Packer will use the environment variable OS_USERNAME or
	// OS_USERID, if set. This is not required if using access token or
	// application credential instead of password, or if using cloud.yaml.
	Username string `mapstructure:"username" required:"true"`
	// Sets username
	UserID string `mapstructure:"user_id"`
	// The password used to connect to the OpenStack service. If not specified,
	// Packer will use the environment variables OS_PASSWORD, if set. This is
	// not required if using access token or application credential instead of
	// password, or if using cloud.yaml.
	Password string `mapstructure:"password" required:"true"`
	// The URL to the OpenStack Identity service. If not specified, Packer will
	// use the environment variables OS_AUTH_URL, if set. This is not required
	// if using cloud.yaml.
	IdentityEndpoint string `mapstructure:"identity_endpoint" required:"true"`
	// The tenant ID or name to boot the instance into. Some OpenStack
	// installations require this. If not specified, Packer will use the
	// environment variable OS_TENANT_NAME or OS_TENANT_ID, if set. Tenant is
	// also called Project in later versions of OpenStack.
	TenantID   string `mapstructure:"tenant_id" required:"false"`
	TenantName string `mapstructure:"tenant_name"`
	DomainID   string `mapstructure:"domain_id"`
	// The Domain name or ID you are authenticating with. OpenStack
	// installations require this if identity v3 is used. Packer will use the
	// environment variable OS_DOMAIN_NAME or OS_DOMAIN_ID, if set.
	DomainName string `mapstructure:"domain_name" required:"false"`
	// Whether or not the connection to OpenStack can be done over an insecure
	// connection. By default this is false.
	Insecure bool `mapstructure:"insecure" required:"false"`
	// The name of the region, such as "DFW", in which to launch the server to
	// create the image. If not specified, Packer will use the environment
	// variable OS_REGION_NAME, if set.
	Region string `mapstructure:"region" required:"false"`
	// The endpoint type to use. Can be any of "internal", "internalURL",
	// "admin", "adminURL", "public", and "publicURL". By default this is
	// "public".
	EndpointType string `mapstructure:"endpoint_type" required:"false"`
	// Custom CA certificate file path. If omitted the OS_CACERT environment
	// variable can be used.
	CACertFile string `mapstructure:"cacert" required:"false"`
	// Client certificate file path for SSL client authentication. If omitted
	// the OS_CERT environment variable can be used.
	ClientCertFile string `mapstructure:"cert" required:"false"`
	// Client private key file path for SSL client authentication. If omitted
	// the OS_KEY environment variable can be used.
	ClientKeyFile string `mapstructure:"key" required:"false"`
	// the token (id) to use with token based authorization. Packer will use
	// the environment variable OS_TOKEN, if set.
	Token string `mapstructure:"token" required:"false"`
	// The application credential name to use with application credential based
	// authorization. Packer will use the environment variable
	// OS_APPLICATION_CREDENTIAL_NAME, if set.
	ApplicationCredentialName string `mapstructure:"application_credential_name" required:"false"`
	// The application credential id to use with application credential based
	// authorization. Packer will use the environment variable
	// OS_APPLICATION_CREDENTIAL_ID, if set.
	ApplicationCredentialID string `mapstructure:"application_credential_id" required:"false"`
	// The application credential secret to use with application credential
	// based authorization. Packer will use the environment variable
	// OS_APPLICATION_CREDENTIAL_SECRET, if set.
	ApplicationCredentialSecret string `mapstructure:"application_credential_secret" required:"false"`
	// An entry in a `clouds.yaml` file. See the OpenStack os-client-config
	// [documentation](https://docs.openstack.org/os-client-config/latest/user/configuration.html)
	// for more information about `clouds.yaml` files. If omitted, the
	// `OS_CLOUD` environment variable is used.
	Cloud string `mapstructure:"cloud" required:"false"`

	osClient *gophercloud.ProviderClient
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	if c.EndpointType != "internal" && c.EndpointType != "internalURL" &&
		c.EndpointType != "admin" && c.EndpointType != "adminURL" &&
		c.EndpointType != "public" && c.EndpointType != "publicURL" &&
		c.EndpointType != "" {
		return []error{fmt.Errorf("Invalid endpoint type provided")}
	}

	// Legacy RackSpace stuff. We're keeping this around to keep things BC.
	if c.Password == "" {
		c.Password = os.Getenv("SDK_PASSWORD")
	}
	if c.Region == "" {
		c.Region = os.Getenv("SDK_REGION")
	}
	if c.TenantName == "" {
		c.TenantName = os.Getenv("SDK_PROJECT")
	}
	if c.Username == "" {
		c.Username = os.Getenv("SDK_USERNAME")
	}
	// End RackSpace

	if c.Cloud == "" {
		c.Cloud = os.Getenv("OS_CLOUD")
	}
	if c.Region == "" {
		c.Region = os.Getenv("OS_REGION_NAME")
	}

	if c.CACertFile == "" {
		c.CACertFile = os.Getenv("OS_CACERT")
	}
	if c.ClientCertFile == "" {
		c.ClientCertFile = os.Getenv("OS_CERT")
	}
	if c.ClientKeyFile == "" {
		c.ClientKeyFile = os.Getenv("OS_KEY")
	}

	clientOpts := new(clientconfig.ClientOpts)

	// If a cloud entry was given, base AuthOptions on a clouds.yaml file.
	if c.Cloud != "" {
		clientOpts.Cloud = c.Cloud

		cloud, err := clientconfig.GetCloudFromYAML(clientOpts)
		if err != nil {
			return []error{err}
		}

		if c.Region == "" && cloud.RegionName != "" {
			c.Region = cloud.RegionName
		}
	} else {
		authInfo := &clientconfig.AuthInfo{
			AuthURL:     c.IdentityEndpoint,
			DomainID:    c.DomainID,
			DomainName:  c.DomainName,
			Password:    c.Password,
			ProjectID:   c.TenantID,
			ProjectName: c.TenantName,
			Token:       c.Token,
			Username:    c.Username,
			UserID:      c.UserID,
		}
		clientOpts.AuthInfo = authInfo
	}

	ao, err := clientconfig.AuthOptions(clientOpts)
	if err != nil {
		return []error{err}
	}

	// Make sure we reauth as needed
	ao.AllowReauth = true

	// Override values if we have them in our config
	overrides := []struct {
		From, To *string
	}{
		{&c.Username, &ao.Username},
		{&c.UserID, &ao.UserID},
		{&c.Password, &ao.Password},
		{&c.IdentityEndpoint, &ao.IdentityEndpoint},
		{&c.TenantID, &ao.TenantID},
		{&c.TenantName, &ao.TenantName},
		{&c.DomainID, &ao.DomainID},
		{&c.DomainName, &ao.DomainName},
		{&c.Token, &ao.TokenID},
		{&c.ApplicationCredentialName, &ao.ApplicationCredentialName},
		{&c.ApplicationCredentialID, &ao.ApplicationCredentialID},
		{&c.ApplicationCredentialSecret, &ao.ApplicationCredentialSecret},
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

	tls_config := &tls.Config{}

	if c.CACertFile != "" {
		caCert, err := ioutil.ReadFile(c.CACertFile)
		if err != nil {
			return []error{err}
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tls_config.RootCAs = caCertPool
	}

	// If we have insecure set, then create a custom HTTP client that ignores
	// SSL errors.
	if c.Insecure {
		tls_config.InsecureSkipVerify = true
	}

	if c.ClientCertFile != "" && c.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile)
		if err != nil {
			return []error{err}
		}

		tls_config.Certificates = []tls.Certificate{cert}
	}

	transport := cleanhttp.DefaultTransport()
	transport.TLSClientConfig = tls_config
	client.HTTPClient.Transport = transport

	// Auth
	err = openstack.Authenticate(client, *ao)
	if err != nil {
		return []error{err}
	}

	c.osClient = client
	return nil
}

func (c *AccessConfig) enableDebug(ui packersdk.Ui) {
	c.osClient.HTTPClient = http.Client{
		Transport: &DebugRoundTripper{
			ui: ui,
			rt: c.osClient.HTTPClient.Transport,
		},
	}
}

func (c *AccessConfig) computeV2Client() (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(c.osClient, gophercloud.EndpointOpts{
		Region:       c.Region,
		Availability: c.getEndpointType(),
	})
}

func (c *AccessConfig) imageV2Client() (*gophercloud.ServiceClient, error) {
	return openstack.NewImageServiceV2(c.osClient, gophercloud.EndpointOpts{
		Region:       c.Region,
		Availability: c.getEndpointType(),
	})
}

func (c *AccessConfig) blockStorageV3Client() (*gophercloud.ServiceClient, error) {
	return openstack.NewBlockStorageV3(c.osClient, gophercloud.EndpointOpts{
		Region:       c.Region,
		Availability: c.getEndpointType(),
	})
}

func (c *AccessConfig) networkV2Client() (*gophercloud.ServiceClient, error) {
	return openstack.NewNetworkV2(c.osClient, gophercloud.EndpointOpts{
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

type DebugRoundTripper struct {
	ui                packersdk.Ui
	rt                http.RoundTripper
	numReauthAttempts int
}

// RoundTrip performs a round-trip HTTP request and logs relevant information about it.
func (drt *DebugRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()

	var response *http.Response
	var err error

	response, err = drt.rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		if drt.numReauthAttempts == 3 {
			return response, fmt.Errorf("Tried to re-authenticate 3 times with no success.")
		}
		drt.numReauthAttempts++
	}

	drt.DebugMessage(fmt.Sprintf("Request %s %s %d", request.Method, request.URL, response.StatusCode))

	if response.StatusCode >= 400 {
		buf := bytes.NewBuffer([]byte{})
		body, _ := ioutil.ReadAll(io.TeeReader(response.Body, buf))
		drt.DebugMessage(fmt.Sprintf("Response Error: %+v\n", string(body)))
		bufWithClose := ioutil.NopCloser(buf)
		response.Body = bufWithClose
	}

	return response, err
}

func (drt *DebugRoundTripper) DebugMessage(message string) {
	drt.ui.Message(fmt.Sprintf("[DEBUG] %s", message))
}
