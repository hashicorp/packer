package triton

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joyent/gocommon/client"
	"github.com/joyent/gosdc/cloudapi"
	"github.com/joyent/gosign/auth"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

// AccessConfig is for common configuration related to Triton access
type AccessConfig struct {
	Endpoint    string `mapstructure:"triton_url"`
	Account     string `mapstructure:"triton_account"`
	KeyID       string `mapstructure:"triton_key_id"`
	KeyMaterial string `mapstructure:"triton_key_material"`
}

// Prepare performs basic validation on the AccessConfig
func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.Endpoint == "" {
		// Use Joyent public cloud as the default endpoint if none is in environment
		c.Endpoint = "https://us-east-1.api.joyent.com"
	}

	if c.Account == "" {
		errs = append(errs, fmt.Errorf("triton_account is required to use the triton builder"))
	}

	if c.KeyID == "" {
		errs = append(errs, fmt.Errorf("triton_key_id is required to use the triton builder"))
	}

	var err error
	c.KeyMaterial, err = processKeyMaterial(c.KeyMaterial)
	if c.KeyMaterial == "" || err != nil {
		errs = append(errs, fmt.Errorf("valid triton_key_material is required to use the triton builder"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// CreateTritonClient returns an SDC client configured with the appropriate client credentials
// or an error if creating the client fails.
func (c *AccessConfig) CreateTritonClient() (*cloudapi.Client, error) {
	keyData, err := processKeyMaterial(c.KeyMaterial)
	if err != nil {
		return nil, err
	}

	userauth, err := auth.NewAuth(c.Account, string(keyData), "rsa-sha256")
	if err != nil {
		return nil, err
	}

	creds := &auth.Credentials{
		UserAuthentication: userauth,
		SdcKeyId:           c.KeyID,
		SdcEndpoint:        auth.Endpoint{URL: c.Endpoint},
	}

	return cloudapi.New(client.NewClient(
		c.Endpoint,
		cloudapi.DefaultAPIVersion,
		creds,
		log.New(os.Stdout, "", log.Flags()),
	)), nil
}

func (c *AccessConfig) Comm() communicator.Config {
	return communicator.Config{}
}

func processKeyMaterial(keyMaterial string) (string, error) {
	// Check for keyMaterial being a file path
	if _, err := os.Stat(keyMaterial); err != nil {
		// Not a valid file. Assume that keyMaterial is the key data
		return keyMaterial, nil
	}

	b, err := ioutil.ReadFile(keyMaterial)
	if err != nil {
		return "", fmt.Errorf("Error reading key_material from path '%s': %s",
			keyMaterial, err)
	}

	return string(b), nil
}
