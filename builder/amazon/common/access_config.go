package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey            string `mapstructure:"access_key"`
	CustomEndpointEc2    string `mapstructure:"custom_endpoint_ec2"`
	MFACode              string `mapstructure:"mfa_code"`
	ProfileName          string `mapstructure:"profile"`
	RawRegion            string `mapstructure:"region"`
	SecretKey            string `mapstructure:"secret_key"`
	SkipValidation       bool   `mapstructure:"skip_region_validation"`
	SkipMetadataApiCheck bool   `mapstructure:"skip_metadata_api_check"`
	Token                string `mapstructure:"token"`
	session              *session.Session
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Session() (*session.Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	// build a chain provider, lazy-evaluated by aws-sdk
	providers := []credentials.Provider{
		&credentials.StaticProvider{Value: credentials.Value{
			AccessKeyID:     c.AccessKey,
			SecretAccessKey: c.SecretKey,
			SessionToken:    c.Token,
		}},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{
			Filename: "",
			Profile:  c.ProfileName,
		},
	}

	// Build isolated HTTP client to avoid issues with globally-shared settings
	client := cleanhttp.DefaultClient()

	// Keep the default timeout (100ms) low as we don't want to wait in non-EC2 environments
	client.Timeout = 100 * time.Millisecond

	const userTimeoutEnvVar = "AWS_METADATA_TIMEOUT"
	userTimeout := os.Getenv(userTimeoutEnvVar)
	if userTimeout != "" {
		newTimeout, err := time.ParseDuration(userTimeout)
		if err == nil {
			if newTimeout.Nanoseconds() > 0 {
				client.Timeout = newTimeout
			} else {
				log.Printf("[WARN] Non-positive value of %s (%s) is meaningless, ignoring", userTimeoutEnvVar, newTimeout.String())
			}
		} else {
			log.Printf("[WARN] Error converting %s to time.Duration: %s", userTimeoutEnvVar, err)
		}
	}

	log.Printf("[INFO] Setting AWS metadata API timeout to %s", client.Timeout.String())
	cfg := &aws.Config{
		HTTPClient: client,
	}
	if !c.SkipMetadataApiCheck {
		// Real AWS should reply to a simple metadata request.
		// We check it actually does to ensure something else didn't just
		// happen to be listening on the same IP:Port
		metadataClient := ec2metadata.New(session.New(cfg))
		if metadataClient.Available() {
			providers = append(providers, &ec2rolecreds.EC2RoleProvider{
				Client: metadataClient,
			})
			log.Print("[INFO] AWS EC2 instance detected via default metadata" +
				" API endpoint, EC2RoleProvider added to the auth chain")
		} else {
			log.Printf("[INFO] Ignoring AWS metadata API endpoint " +
				"as it doesn't return any instance-id")
		}
	}

	creds := credentials.NewChainCredentials(providers)
	cp, err := creds.Get()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoCredentialProviders" {
			return nil, errors.New("No valid credential sources found for AWS Builder. " +
				"Please see https://www.packer.io/docs/builders/amazon.html#specifying-amazon-credentials " +
				"for more information on providing credentials for the AWS Builder.")
		}

		return nil, fmt.Errorf("Error loading credentials for AWS Provider: %s", err)
	}
	log.Printf("[INFO] AWS Auth provider used: %q", cp.ProviderName)

	config := aws.NewConfig().WithMaxRetries(11).WithCredentialsChainVerboseErrors(true)
	config = config.WithCredentials(creds)

	if c.RawRegion != "" {
		config = config.WithRegion(c.RawRegion)
	} else if region := c.metadataRegion(); region != "" {
		config = config.WithRegion(region)
	}

	if c.CustomEndpointEc2 != "" {
		config = config.WithEndpoint(c.CustomEndpointEc2)
	}

	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *config,
	}

	if c.MFACode != "" {
		opts.AssumeRoleTokenProvider = func() (string, error) {
			return c.MFACode, nil
		}
	}

	if sess, err := session.NewSessionWithOptions(opts); err != nil {
		return nil, err
	} else if *sess.Config.Region == "" {
		return nil, fmt.Errorf("Could not find AWS region, make sure it's set.")
	} else {
		log.Printf("Found region %s", *sess.Config.Region)
		c.session = sess
	}

	return c.session, nil
}

// metadataRegion returns the region from the metadata service
func (c *AccessConfig) metadataRegion() string {

	client := cleanhttp.DefaultClient()

	// Keep the default timeout (100ms) low as we don't want to wait in non-EC2 environments
	client.Timeout = 100 * time.Millisecond
	ec2meta := ec2metadata.New(session.New(), &aws.Config{
		HTTPClient: client,
	})
	region, err := ec2meta.Region()
	if err != nil {
		log.Println("Error getting region from metadata service, "+
			"probably because we're not running on AWS.", err)
		return ""
	}
	return region
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	// Either both access and secret key must be set or neither of them should
	// be.
	if (len(c.AccessKey) > 0) != (len(c.SecretKey) > 0) {
		errs = append(errs,
			fmt.Errorf("`access_key` and `secret_key` must both be either set or not set."))
	}

	if c.RawRegion != "" && !c.SkipValidation {
		if valid := ValidateRegion(c.RawRegion); !valid {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	return errs
}
