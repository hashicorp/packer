package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-sdk-go/osc"
)

// AccessConfig is for common configuration related to Outscale API access
type AccessConfig struct {
	AccessKey             string `mapstructure:"access_key"`
	CustomEndpointOAPI    string `mapstructure:"custom_endpoint_oapi"`
	InsecureSkipTLSVerify bool   `mapstructure:"insecure_skip_tls_verify"`
	MFACode               string `mapstructure:"mfa_code"`
	ProfileName           string `mapstructure:"profile"`
	RawRegion             string `mapstructure:"region"`
	SecretKey             string `mapstructure:"secret_key"`
	SkipValidation        bool   `mapstructure:"skip_region_validation"`
	SkipMetadataApiCheck  bool   `mapstructure:"skip_metadata_api_check"`
	Token                 string `mapstructure:"token"`
	X509certPath          string `mapstructure:"x509_cert_path"`
	X509keyPath           string `mapstructure:"x509_key_path"`
}

// NewOSCClient retrieves the Outscale OSC-SDK client
func (c *AccessConfig) NewOSCClient() *osc.APIClient {
	if c.AccessKey == "" {
		c.AccessKey = os.Getenv("OUTSCALE_ACCESSKEYID")
	}

	if c.SecretKey == "" {
		c.SecretKey = os.Getenv("OUTSCALE_SECRETKEYID")
	}

	if c.RawRegion == "" {
		c.RawRegion = os.Getenv("OUTSCALE_REGION")
	}

	if c.CustomEndpointOAPI == "" {
		c.CustomEndpointOAPI = os.Getenv("OUTSCALE_OAPI_URL")
	}

	if c.CustomEndpointOAPI == "" {
		c.CustomEndpointOAPI = "outscale.com/oapi/latest"

		if c.RawRegion == "cn-southeast-1" {
			c.CustomEndpointOAPI = "outscale.hk/oapi/latest"
		}

	}

	if c.X509certPath == "" {
		c.X509certPath = os.Getenv("OUTSCALE_X509CERT")
	}

	if c.X509keyPath == "" {
		c.X509keyPath = os.Getenv("OUTSCALE_X509KEY")
	}

	return c.NewOSCClientByRegion(c.RawRegion)
}

// GetRegion retrieves the Outscale OSC-SDK Region set
func (c *AccessConfig) GetRegion() string {
	return c.RawRegion
}

// NewOSCClientByRegion returns the connection depdending of the region given
func (c *AccessConfig) NewOSCClientByRegion(region string) *osc.APIClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipTLSVerify},
		Proxy:           http.ProxyFromEnvironment,
	}

	if c.X509certPath != "" && c.X509keyPath != "" {
		cert, err := tls.LoadX509KeyPair(c.X509certPath, c.X509keyPath)
		if err == nil {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: c.InsecureSkipTLSVerify,
				Certificates:       []tls.Certificate{cert},
			}
		}
	}

	skipClient := &http.Client{
		Transport: transport,
	}

	skipClient.Transport = NewTransport(c.AccessKey, c.SecretKey, c.RawRegion, skipClient.Transport)

	return osc.NewAPIClient(&osc.Configuration{
		BasePath:      fmt.Sprintf("https://api.%s.%s", region, c.CustomEndpointOAPI),
		DefaultHeader: make(map[string]string),
		UserAgent:     "packer-osc",
		HTTPClient:    skipClient,
		Debug:         true,
	})
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.SkipMetadataApiCheck {
		log.Println("(WARN) skip_metadata_api_check ignored.")
	}
	// Either both access and secret key must be set or neither of them should
	// be.
	if (len(c.AccessKey) > 0) != (len(c.SecretKey) > 0) {
		errs = append(errs,
			fmt.Errorf("`access_key` and `secret_key` must both be either set or not set."))
	}

	return errs
}
