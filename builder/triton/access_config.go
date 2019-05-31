//go:generate struct-markdown

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
	// The URL of the Triton cloud API to use. If omitted
    // it will default to the us-sw-1 region of the Joyent Public cloud. If you
    // are using your own private Triton installation you will have to supply the
    // URL of the cloud API of your own Triton installation.
	Endpoint              string `mapstructure:"triton_url" required:"false"`
	// The username of the Triton account to use when
    // using the Triton Cloud API.
	Account               string `mapstructure:"triton_account" required:"true"`
	// The username of a user who has access to your
    // Triton account.
	Username              string `mapstructure:"triton_user" required:"false"`
	// The fingerprint of the public key of the SSH key
    // pair to use for authentication with the Triton Cloud API. If
    // triton_key_material is not set, it is assumed that the SSH agent has the
    // private key corresponding to this key ID loaded.
	KeyID                 string `mapstructure:"triton_key_id" required:"true"`
	// Path to the file in which the private key
    // of triton_key_id is stored. For example /home/soandso/.ssh/id_rsa. If
    // this is not specified, the SSH agent is used to sign requests with the
    // triton_key_id specified.
	KeyMaterial           string `mapstructure:"triton_key_material" required:"false"`
	//secure_skip_tls_verify - (bool) This allows skipping TLS verification
    // of the Triton endpoint. It is useful when connecting to a temporary Triton
    // installation such as Cloud-On-A-Laptop which does not generally use a
    // certificate signed by a trusted root CA. The default is false.
	InsecureSkipTLSVerify bool   `mapstructure:"insecure_skip_tls_verify" required:"false"`

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
		config:                config,
		insecureSkipTLSVerify: c.InsecureSkipTLSVerify,
	}, nil
}

type Client struct {
	config                *tgo.ClientConfig
	insecureSkipTLSVerify bool
}

func (c *Client) Compute() (*compute.ComputeClient, error) {
	computeClient, err := compute.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Compute Client: {{err}}", err)
	}

	if c.insecureSkipTLSVerify {
		computeClient.Client.InsecureSkipTLSVerify()
	}

	return computeClient, nil
}

func (c *Client) Network() (*network.NetworkClient, error) {
	networkClient, err := network.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Network Client: {{err}}", err)
	}

	if c.insecureSkipTLSVerify {
		networkClient.Client.InsecureSkipTLSVerify()
	}

	return networkClient, nil
}

func (c *AccessConfig) Comm() communicator.Config {
	return communicator.Config{}
}
