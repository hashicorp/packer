//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type VaultAWSEngineOptions,AssumeRoleConfig

package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/packer-plugin-amazon/builder/common/awserrors"
	vaultapi "github.com/hashicorp/vault/api"
)

// AssumeRoleConfig lets users set configuration options for assuming a special
// role when executing Packer.
//
// Usage example:
//
// HCL config example:
//
// ```HCL
// source "amazon-ebs" "example" {
// 	assume_role {
// 		role_arn     = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
// 		session_name = "SESSION_NAME"
// 		external_id  = "EXTERNAL_ID"
// 	}
// }
// ```
//
// JSON config example:
//
// ```json
// builder{
// 	"type": "amazon-ebs",
// 	"assume_role": {
// 		"role_arn"    :  "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME",
// 		"session_name":  "SESSION_NAME",
// 		"external_id" :  "EXTERNAL_ID"
// 	}
// }
// ```
type AssumeRoleConfig struct {
	// Amazon Resource Name (ARN) of the IAM Role to assume.
	AssumeRoleARN string `mapstructure:"role_arn" required:"false"`
	// Number of seconds to restrict the assume role session duration.
	AssumeRoleDurationSeconds int `mapstructure:"duration_seconds" required:"false"`
	// The external ID to use when assuming the role. If omitted, no external
	// ID is passed to the AssumeRole call.
	AssumeRoleExternalID string `mapstructure:"external_id" required:"false"`
	// IAM Policy JSON describing further restricting permissions for the IAM
	// Role being assumed.
	AssumeRolePolicy string `mapstructure:"policy" required:"false"`
	// Set of Amazon Resource Names (ARNs) of IAM Policies describing further
	// restricting permissions for the IAM Role being
	AssumeRolePolicyARNs []string `mapstructure:"policy_arns" required:"false"`
	// Session name to use when assuming the role.
	AssumeRoleSessionName string `mapstructure:"session_name" required:"false"`
	// Map of assume role session tags.
	AssumeRoleTags map[string]string `mapstructure:"tags" required:"false"`
	// Set of assume role session tag keys to pass to any subsequent sessions.
	AssumeRoleTransitiveTagKeys []string `mapstructure:"transitive_tag_keys" required:"false"`
}

type VaultAWSEngineOptions struct {
	Name    string `mapstructure:"name"`
	RoleARN string `mapstructure:"role_arn"`
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
	// The access key used to communicate with AWS. [Learn how  to set this]
	// (/docs/builders/amazon#specifying-amazon-credentials). On EBS, this
	// is not required if you are using `use_vault_aws_engine` for
	// authentication instead.
	AccessKey string `mapstructure:"access_key" required:"true"`
	// If provided with a role ARN, Packer will attempt to assume this role
	// using the supplied credentials. See
	// [AssumeRoleConfig](#assume-role-configuration) below for more
	// details on all of the options available, and for a usage example.
	AssumeRole AssumeRoleConfig `mapstructure:"assume_role" required:"false"`
	// This option is useful if you use a cloud
	// provider whose API is compatible with aws EC2. Specify another endpoint
	// like this https://ec2.custom.endpoint.com.
	CustomEndpointEc2 string `mapstructure:"custom_endpoint_ec2" required:"false"`
	// Path to a credentials file to load credentials from
	CredsFilename string `mapstructure:"shared_credentials_file" required:"false"`
	// Enable automatic decoding of any encoded authorization (error) messages
	// using the `sts:DecodeAuthorizationMessage` API. Note: requires that the
	// effective user/role have permissions to `sts:DecodeAuthorizationMessage`
	// on resource `*`. Default `false`.
	DecodeAuthZMessages bool `mapstructure:"decode_authorization_messages" required:"false"`
	// This allows skipping TLS
	// verification of the AWS EC2 endpoint. The default is false.
	InsecureSkipTLSVerify bool `mapstructure:"insecure_skip_tls_verify" required:"false"`
	// This is the maximum number of times an API call is retried, in the case
	// where requests are being throttled or experiencing transient failures.
	// The delay between the subsequent API calls increases exponentially.
	MaxRetries int `mapstructure:"max_retries" required:"false"`
	// The MFA
	// [TOTP](https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm)
	// code. This should probably be a user variable since it changes all the
	// time.
	MFACode string `mapstructure:"mfa_code" required:"false"`
	// The profile to use in the shared credentials file for
	// AWS. See Amazon's documentation on [specifying
	// profiles](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-profiles)
	// for more details.
	ProfileName string `mapstructure:"profile" required:"false"`
	// The name of the region, such as `us-east-1`, in which
	// to launch the EC2 instance to create the AMI.
	// When chroot building, this value is guessed from environment.
	RawRegion string `mapstructure:"region" required:"true"`
	// The secret key used to communicate with AWS. [Learn how to set
	// this](/docs/builders/amazon#specifying-amazon-credentials). This is not required
	// if you are using `use_vault_aws_engine` for authentication instead.
	SecretKey            string `mapstructure:"secret_key" required:"true"`
	SkipMetadataApiCheck bool   `mapstructure:"skip_metadata_api_check"`
	// Set to true if you want to skip validating AWS credentials before runtime.
	SkipCredsValidation bool `mapstructure:"skip_credential_validation"`
	// The access token to use. This is different from the
	// access key and secret key. If you're not sure what this is, then you
	// probably don't need it. This will also be read from the AWS_SESSION_TOKEN
	// environmental variable.
	Token   string `mapstructure:"token" required:"false"`
	session *session.Session
	// Get credentials from Hashicorp Vault's aws secrets engine. You must
	// already have created a role to use. For more information about
	// generating credentials via the Vault engine, see the [Vault
	// docs.](https://www.vaultproject.io/api/secret/aws#generate-credentials)
	// If you set this flag, you must also set the below options:
	// -   `name` (string) - Required. Specifies the name of the role to generate
	//     credentials against. This is part of the request URL.
	// -   `engine_name` (string) - The name of the aws secrets engine. In the
	//     Vault docs, this is normally referred to as "aws", and Packer will
	//     default to "aws" if `engine_name` is not set.
	// -   `role_arn` (string)- The ARN of the role to assume if credential\_type
	//     on the Vault role is assumed\_role. Must match one of the allowed role
	//     ARNs in the Vault role. Optional if the Vault role only allows a single
	//     AWS role ARN; required otherwise.
	// -   `ttl` (string) - Specifies the TTL for the use of the STS token. This
	//     is specified as a string with a duration suffix. Valid only when
	//     credential\_type is assumed\_role or federation\_token. When not
	//     specified, the default\_sts\_ttl set for the role will be used. If that
	//     is also not set, then the default value of 3600s will be used. AWS
	//     places limits on the maximum TTL allowed. See the AWS documentation on
	//     the DurationSeconds parameter for AssumeRole (for assumed\_role
	//     credential types) and GetFederationToken (for federation\_token
	//     credential types) for more details.
	//
	// HCL2 example:
	//
	// ```hcl
	// vault_aws_engine {
	//     name = "myrole"
	//     role_arn = "myarn"
	//     ttl = "3600s"
	// }
	// ```
	//
	// JSON example:
	//
	// ```json
	// {
	//     "vault_aws_engine": {
	//         "name": "myrole",
	//         "role_arn": "myarn",
	//         "ttl": "3600s"
	//     }
	// }
	// ```
	VaultAWSEngine VaultAWSEngineOptions `mapstructure:"vault_aws_engine" required:"false"`
	// [Polling configuration](#polling-configuration) for the AWS waiter. Configures the waiter that checks
	// resource state.
	PollingConfig *AWSPollingConfig `mapstructure:"aws_polling" required:"false"`

	getEC2Connection func() ec2iface.EC2API
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Session() (*session.Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	// Create new AWS config
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	if c.MaxRetries > 0 {
		config = config.WithMaxRetries(c.MaxRetries)
	}

	// Set AWS config defaults.
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

	// Figure out which possible credential providers are valid; test that we
	// can get credentials via the selected providers, and set the providers in
	// the config.
	creds, err := c.GetCredentials(config)
	if err != nil {
		return nil, err
	}
	config.WithCredentials(creds)

	// Create session options based on our AWS config
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

	if awserrors.Matches(err, "NoCredentialProviders", "") {
		return nil, c.NewNoValidCredentialSourcesError(err)
	}

	if err != nil {
		return nil, fmt.Errorf("Error loading credentials for AWS Provider: %s", err)
	}

	log.Printf("[INFO] AWS Auth provider used: %q", cp.ProviderName)

	if c.DecodeAuthZMessages {
		DecodeAuthZMessages(c.session)
	}

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

// GetCredentials gets credentials from the environment, shared credentials,
// the session (which may include a credential process), or ECS/EC2 metadata
// endpoints. GetCredentials also validates the credentials and the ability to
// assume a role or will return an error if unsuccessful.
func (c *AccessConfig) GetCredentials(config *aws.Config) (*awsCredentials.Credentials, error) {
	// Reload values into the config used by the Packer-Terraform shared SDK
	awsbaseConfig := &awsbase.Config{
		AccessKey:                   c.AccessKey,
		AssumeRoleARN:               c.AssumeRole.AssumeRoleARN,
		AssumeRoleDurationSeconds:   c.AssumeRole.AssumeRoleDurationSeconds,
		AssumeRoleExternalID:        c.AssumeRole.AssumeRoleExternalID,
		AssumeRolePolicy:            c.AssumeRole.AssumeRolePolicy,
		AssumeRolePolicyARNs:        c.AssumeRole.AssumeRolePolicyARNs,
		AssumeRoleSessionName:       c.AssumeRole.AssumeRoleSessionName,
		AssumeRoleTags:              c.AssumeRole.AssumeRoleTags,
		AssumeRoleTransitiveTagKeys: c.AssumeRole.AssumeRoleTransitiveTagKeys,
		CredsFilename:               c.CredsFilename,
		DebugLogging:                false,
		// TODO: implement for Packer
		// IamEndpoint:                 c.Endpoints["iam"],
		Insecure:             c.InsecureSkipTLSVerify,
		MaxRetries:           c.MaxRetries,
		Profile:              c.ProfileName,
		Region:               c.RawRegion,
		SecretKey:            c.SecretKey,
		SkipCredsValidation:  c.SkipCredsValidation,
		SkipMetadataApiCheck: c.SkipMetadataApiCheck,
		// TODO: implement for Packer
		// SkipRequestingAccountId:     c.SkipRequestingAccountId,
		// StsEndpoint:                 c.Endpoints["sts"],
		Token: c.Token,
	}

	return awsbase.GetCredentials(awsbaseConfig)
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

func (c *AccessConfig) Prepare() []error {
	var errs []error

	if c.SkipMetadataApiCheck {
		log.Println("(WARN) skip_metadata_api_check ignored.")
	}

	// Make sure it's obvious from the config how we're getting credentials:
	// Vault, Packer config, or environment.
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

	if c.PollingConfig == nil {
		c.PollingConfig = new(AWSPollingConfig)
	}
	c.PollingConfig.LogEnvOverrideWarnings()

	// Default MaxRetries to 10, to make throttling issues less likely. The
	// Aws sdk defaults this to 3, which regularly gets tripped by users.
	if c.MaxRetries == 0 {
		c.MaxRetries = 10
	}

	return errs
}

func (c *AccessConfig) NewNoValidCredentialSourcesError(err error) error {
	return fmt.Errorf("No valid credential sources found for AWS Builder. "+
		"Please see https://www.packer.io/docs/builders/amazon#authentication "+
		"for more information on providing credentials for the AWS Builder. "+
		"Error: %w", err)
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
