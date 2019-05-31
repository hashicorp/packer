//go:generate struct-markdown

package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/packer/template/interpolate"
	vaultapi "github.com/hashicorp/vault/api"
)

type VaultAWSEngineOptions struct {
	Name       string `mapstructure:"name"`
	RoleARN    string `mapstructure:"role_arn"`
	// Specifies the TTL for the use of the STS token. This
    // is specified as a string with a duration suffix. Valid only when
    // credential_type is assumed_role or federation_token. When not
    // specified, the default_sts_ttl set for the role will be used. If that
    // is also not set, then the default value of 3600s will be used. AWS
    // places limits on the maximum TTL allowed. See the AWS documentation on
    // the DurationSeconds parameter for AssumeRole (for assumed_role
    // credential types) and GetFederationToken (for federation_token
    // credential types) for more details.
	TTL        string `mapstructure:"ttl" required:"false"`
	EngineName string `mapstructure:"engine_name"`
}

func (v *VaultAWSEngineOptions) Empty() bool {
	return len(v.Name) == 0 && len(v.RoleARN) == 0 &&
		len(v.EngineName) == 0 && len(v.TTL) == 0
}

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	// The access key used to communicate with AWS. Learn
    // how to set this
	AccessKey             string `mapstructure:"access_key" required:"true"`
	// This option is useful if you use a cloud
    // provider whose API is compatible with aws EC2. Specify another endpoint
    // like this https://ec2.custom.endpoint.com.
	CustomEndpointEc2     string `mapstructure:"custom_endpoint_ec2" required:"false"`
	// Enable automatic decoding of
    // any encoded authorization (error) messages using the
    // sts:DecodeAuthorizationMessage API. Note: requires that the effective
    // user/role have permissions to sts:DecodeAuthorizationMessage on resource
    // *. Default false.
	DecodeAuthZMessages   bool   `mapstructure:"decode_authorization_messages" required:"false"`
	// This allows skipping TLS
    // verification of the AWS EC2 endpoint. The default is false.
	InsecureSkipTLSVerify bool   `mapstructure:"insecure_skip_tls_verify" required:"false"`
	// The MFA
    // TOTP
    // code. This should probably be a user variable since it changes all the
    // time.
	MFACode               string `mapstructure:"mfa_code" required:"false"`
	// The profile to use in the shared credentials file for
    // AWS. See Amazon's documentation on specifying
    // profiles
    // for more details.
	ProfileName           string `mapstructure:"profile" required:"false"`
	// The name of the region, such as us-east-1, in which
    // to launch the EC2 instance to create the AMI.
	RawRegion             string `mapstructure:"region" required:"true"`
	// The secret key used to communicate with AWS. Learn
    // how to set this
	SecretKey             string `mapstructure:"secret_key" required:"true"`
	// Set to true if you want to skip
    // validation of the ami_regions configuration option. Default false.
	SkipValidation        bool   `mapstructure:"skip_region_validation" required:"false"`
	SkipMetadataApiCheck  bool   `mapstructure:"skip_metadata_api_check"`
	// The access token to use. This is different from the
    // access key and secret key. If you're not sure what this is, then you
    // probably don't need it. This will also be read from the AWS_SESSION_TOKEN
    // environmental variable.
	Token                 string `mapstructure:"token" required:"false"`
	session               *session.Session
	// Get credentials from Hashicorp Vault's aws
    // secrets engine. You must already have created a role to use. For more
    // information about generating credentials via the Vault engine, see the
    // Vault
    // docs.
    // If you set this flag, you must also set the below options:
	VaultAWSEngine        VaultAWSEngineOptions `mapstructure:"vault_aws_engine" required:"false"`

	getEC2Connection func() ec2iface.EC2API
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Session() (*session.Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)

	staticCreds := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	if _, err := staticCreds.Get(); err != credentials.ErrStaticCredentialsEmpty {
		config.WithCredentials(staticCreds)
	}

	if c.RawRegion != "" {
		config = config.WithRegion(c.RawRegion)
	}

	if c.CustomEndpointEc2 != "" {
		config = config.WithEndpoint(c.CustomEndpointEc2)
	}

	config = config.WithHTTPClient(cleanhttp.DefaultClient())
	transport := config.HTTPClient.Transport.(*http.Transport)
	if c.InsecureSkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	transport.Proxy = http.ProxyFromEnvironment

	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *config,
	}

	if c.ProfileName != "" {
		opts.Profile = c.ProfileName
	}

	if c.MFACode != "" {
		opts.AssumeRoleTokenProvider = func() (string, error) {
			return c.MFACode, nil
		}
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}
	log.Printf("Found region %s", *sess.Config.Region)
	c.session = sess

	cp, err := c.session.Config.Credentials.Get()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoCredentialProviders" {
			return nil, fmt.Errorf("No valid credential sources found for AWS Builder. " +
				"Please see https://www.packer.io/docs/builders/amazon.html#specifying-amazon-credentials " +
				"for more information on providing credentials for the AWS Builder.")
		} else {
			return nil, fmt.Errorf("Error loading credentials for AWS Provider: %s", err)
		}
	}
	log.Printf("[INFO] AWS Auth provider used: %q", cp.ProviderName)

	if c.DecodeAuthZMessages {
		DecodeAuthZMessages(c.session)
	}
	LogEnvOverrideWarnings()

	return c.session, nil
}

func (c *AccessConfig) SessionRegion() string {
	if c.session == nil {
		panic("access config session should be set.")
	}
	return aws.StringValue(c.session.Config.Region)
}

func (c *AccessConfig) IsGovCloud() bool {
	return strings.HasPrefix(c.SessionRegion(), "us-gov-")
}

func (c *AccessConfig) IsChinaCloud() bool {
	return strings.HasPrefix(c.SessionRegion(), "cn-")
}

func (c *AccessConfig) GetCredsFromVault() error {
	// const EnvVaultAddress = "VAULT_ADDR"
	// const EnvVaultToken = "VAULT_TOKEN"
	vaultConfig := vaultapi.DefaultConfig()
	cli, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("Error getting Vault client: %s", err)
	}
	if c.VaultAWSEngine.EngineName == "" {
		c.VaultAWSEngine.EngineName = "aws"
	}
	path := fmt.Sprintf("/%s/creds/%s", c.VaultAWSEngine.EngineName,
		c.VaultAWSEngine.Name)
	secret, err := cli.Logical().Read(path)
	if err != nil {
		return fmt.Errorf("Error reading vault secret: %s", err)
	}
	if secret == nil {
		return fmt.Errorf("Vault Secret does not exist at the given path.")
	}

	c.AccessKey = secret.Data["access_key"].(string)
	c.SecretKey = secret.Data["secret_key"].(string)
	token := secret.Data["security_token"]
	if token != nil {
		c.Token = token.(string)
	} else {
		c.Token = ""
	}

	return nil
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.SkipMetadataApiCheck {
		log.Println("(WARN) skip_metadata_api_check ignored.")
	}

	// Make sure it's obvious from the config how we're getting credentials:
	// Vault, Packer config, or environemnt.
	if !c.VaultAWSEngine.Empty() {
		if len(c.AccessKey) > 0 {
			errs = append(errs,
				fmt.Errorf("If you have set vault_aws_engine, you must not set"+
					" the access_key or secret_key."))
		}
		// Go ahead and grab those credentials from Vault now, so we can set
		// the keys and token now.
		err := c.GetCredsFromVault()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if (len(c.AccessKey) > 0) != (len(c.SecretKey) > 0) {
		errs = append(errs,
			fmt.Errorf("`access_key` and `secret_key` must both be either set or not set."))
	}

	return errs
}

func (c *AccessConfig) NewEC2Connection() (ec2iface.EC2API, error) {
	if c.getEC2Connection != nil {
		return c.getEC2Connection(), nil
	}
	sess, err := c.Session()
	if err != nil {
		return nil, err
	}

	ec2conn := ec2.New(sess)

	return ec2conn, nil
}
