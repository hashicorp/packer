package common

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mitchellh/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	RawRegion      string `mapstructure:"region"`
	SkipValidation bool   `mapstructure:"skip_region_validation"`
	Token          string `mapstructure:"token"`
	ProfileName    string `mapstructure:"profile"`
	AssumeRoleArn  string `mapstructure:"assume_role_arn"`
	MFASerial      string `mapstructure:"mfa_serial"`
	MFACode        string `mapstructure:"mfa_code"`
	ExternalID     string `mapstructure:"external_id"`
	session        *session.Session
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Session() (*session.Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	region, err := c.Region()
	if err != nil {
		return nil, err
	}

	if c.ProfileName != "" {
		err := os.Setenv("AWS_PROFILE", c.ProfileName)
		if err != nil {
			log.Printf("Set env error: %s", err)
		}
	}

	config := aws.NewConfig().WithRegion(region).WithMaxRetries(11).WithCredentialsChainVerboseErrors(true)

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
	c.session, err = session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	return c.session, nil
}

// Region returns the aws.Region object for access to AWS services, requesting
// the region from the instance metadata if possible.
func (c *AccessConfig) Region() (string, error) {
	if c.RawRegion != "" {
		if !c.SkipValidation {
			if valid := ValidateRegion(c.RawRegion); valid == false {
				return "", fmt.Errorf("Not a valid region: %s", c.RawRegion)
			}
		}
		return c.RawRegion, nil
	}

	sess := session.New()
	ec2meta := ec2metadata.New(sess)
	identity, err := ec2meta.GetInstanceIdentityDocument()
	if err != nil {
		return "", err
	}
	return identity.Region, nil
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.RawRegion != "" && !c.SkipValidation {
		if valid := ValidateRegion(c.RawRegion); valid == false {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	hasAssumeRoleArn := len(c.AssumeRoleArn) > 0
	hasMFASerial := len(c.MFASerial) > 0
	hasMFACode := len(c.MFACode) > 0
	if hasAssumeRoleArn && (hasMFACode != hasMFASerial) {
		// either both mfa code and serial must be set, or neither.
		errs = append(errs, fmt.Errorf("Both mfa_serial and mfa_code must be specified."))

	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
