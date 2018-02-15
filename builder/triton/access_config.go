package triton

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
	tgo "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
)

// AccessConfig is for common configuration related to Triton access
type AccessConfig struct {
	Endpoint    string `mapstructure:"triton_url"`
	Account     string `mapstructure:"triton_account"`
	Username    string `mapstructure:"triton_user"`
	KeyID       string `mapstructure:"triton_key_id"`
	KeyMaterial string `mapstructure:"triton_key_material"`

	signer authentication.Signer
}

// Prepare performs basic validation on the AccessConfig and ensures we can sign
// a request.
func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.Endpoint == "" {
		// Use Joyent public cloud as the default endpoint if none is specified
		c.Endpoint = "https://us-sw-1.api.joyent.com"
	}

	if c.Account == "" {
		errs = append(errs, errors.New("triton_account is required to use the triton builder"))
	}

	if c.KeyID == "" {
		errs = append(errs, errors.New("triton_key_id is required to use the triton builder"))
	}

	if c.KeyMaterial == "" {
		signer, err := c.createSSHAgentSigner()
		if err != nil {
			errs = append(errs, err)
		}
		c.signer = signer
	} else {
		signer, err := c.createPrivateKeySigner()
		if err != nil {
			errs = append(errs, err)
		}
		c.signer = signer
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AccessConfig) createSSHAgentSigner() (authentication.Signer, error) {
	input := authentication.SSHAgentSignerInput{
		KeyID:       c.KeyID,
		AccountName: c.Account,
		Username:    c.Username,
	}
	signer, err := authentication.NewSSHAgentSigner(input)
	if err != nil {
		return nil, fmt.Errorf("Error creating Triton request signer: %s", err)
	}

	// Ensure we can sign a request
	_, err = signer.Sign("Wed, 26 Apr 2017 16:01:11 UTC")
	if err != nil {
		return nil, fmt.Errorf("Error signing test request: %s", err)
	}

	return signer, nil
}

func (c *AccessConfig) createPrivateKeySigner() (authentication.Signer, error) {
	var privateKeyMaterial []byte
	var err error

	// Check for keyMaterial being a file path
	if _, err = os.Stat(c.KeyMaterial); err != nil {
		privateKeyMaterial = []byte(c.KeyMaterial)
	} else {
		privateKeyMaterial, err = ioutil.ReadFile(c.KeyMaterial)
		if err != nil {
			return nil, fmt.Errorf("Error reading key material from path '%s': %s",
				c.KeyMaterial, err)
		}
	}

	input := authentication.PrivateKeySignerInput{
		KeyID:              c.KeyID,
		AccountName:        c.Account,
		Username:           c.Username,
		PrivateKeyMaterial: privateKeyMaterial,
	}

	signer, err := authentication.NewPrivateKeySigner(input)
	if err != nil {
		return nil, fmt.Errorf("Error creating Triton request signer: %s", err)
	}

	// Ensure we can sign a request
	_, err = signer.Sign("Wed, 26 Apr 2017 16:01:11 UTC")
	if err != nil {
		return nil, fmt.Errorf("Error signing test request: %s", err)
	}

	return signer, nil
}

func (c *AccessConfig) CreateTritonClient() (*Client, error) {

	config := &tgo.ClientConfig{
		AccountName: c.Account,
		TritonURL:   c.Endpoint,
		Username:    c.Username,
		Signers:     []authentication.Signer{c.signer},
	}

	return &Client{
		config: config,
	}, nil
}

type Client struct {
	config *tgo.ClientConfig
}

func (c *Client) Compute() (*compute.ComputeClient, error) {
	computeClient, err := compute.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Compute Client: {{err}}", err)
	}

	return computeClient, nil
}

func (c *Client) Network() (*network.NetworkClient, error) {
	networkClient, err := network.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Network Client: {{err}}", err)
	}

	return networkClient, nil
}

func (c *AccessConfig) Comm() communicator.Config {
	return communicator.Config{}
}
