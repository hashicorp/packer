package common

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey         string `mapstructure:"access_key"`
	CustomEndpointEc2 string `mapstructure:"custom_endpoint_ec2"`
	MFACode           string `mapstructure:"mfa_code"`
	ProfileName       string `mapstructure:"profile"`
	RawRegion         string `mapstructure:"region"`
	SecretKey         string `mapstructure:"secret_key"`
	SkipValidation    bool   `mapstructure:"skip_region_validation"`
	Token             string `mapstructure:"token"`
	session           *session.Session
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Session() (*session.Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	config := aws.NewConfig().WithMaxRetries(11).WithCredentialsChainVerboseErrors(true)

	if c.ProfileName != "" {
		if err := os.Setenv("AWS_PROFILE", c.ProfileName); err != nil {
			return nil, fmt.Errorf("Set env error: %s", err)
		}
	} else if c.RawRegion != "" {
		config = config.WithRegion(c.RawRegion)
	} else if region := c.metadataRegion(); region != "" {
		config = config.WithRegion(region)
	}

	if c.CustomEndpointEc2 != "" {
		config = config.WithEndpoint(c.CustomEndpointEc2)
	}

	if c.AccessKey != "" {
		creds := credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     c.AccessKey,
						SecretAccessKey: c.SecretKey,
						SessionToken:    c.Token,
					},
				},
			})
		config = config.WithCredentials(creds)
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
	if c.RawRegion != "" && !c.SkipValidation {
		if valid := ValidateRegion(c.RawRegion); !valid {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	return errs
}
