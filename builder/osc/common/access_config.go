package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
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
	clientConfig          *oapi.Config

	getOAPIConnection func() oapi.OAPIClient
}

// Config returns a valid oapi.Config object for access to Outscale services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Config() (*oapi.Config, error) {
	if c.clientConfig != nil {
		return c.clientConfig, nil
	}

	//Check env variables if access configuration is not set.

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
	}

	config := &oapi.Config{
		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
		Region:    c.RawRegion,
		URL:       c.CustomEndpointOAPI,
		Service:   "api",
	}

	return config, nil

}

func (c *AccessConfig) NewOAPIConnection() (oapi.OAPIClient, error) {
	if c.getOAPIConnection != nil {
		return c.getOAPIConnection(), nil
	}
	oapicfg, err := c.Config()
	if err != nil {
		return nil, err
	}

	skipClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipTLSVerify},
		},
	}

	oapiClient := oapi.NewClient(oapicfg, skipClient)

	return oapiClient, nil
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
