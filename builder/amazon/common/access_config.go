package common

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)

	staticCreds := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	if _, err := staticCreds.Get(); err != credentials.ErrStaticCredentialsEmpty {
		config.WithCredentials(staticCreds)
	}

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

	if c.ProfileName != "" {
		opts.Profile = c.ProfileName
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

	if c.SkipMetadataApiCheck {
		log.Println("(WARN) skip_metadata_api_check ignored.")
	}
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
