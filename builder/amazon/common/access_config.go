package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey         string `mapstructure:"access_key"`
	SecretKey         string `mapstructure:"secret_key"`
	RawRegion         string `mapstructure:"region"`
	SkipValidation    bool   `mapstructure:"skip_region_validation"`
	Token             string `mapstructure:"token"`
	ProfileName       string `mapstructure:"profile"`
	CustomEndpointEc2 string `mapstructure:"custom_endpoint_ec2"`
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Config() (*aws.Config, error) {
	var creds *credentials.Credentials

	region, err := c.Region()
	if err != nil {
		return nil, err
	}

	config := aws.NewConfig().WithRegion(region).WithMaxRetries(11)
	if c.CustomEndpointEc2 != "" {
		config.Endpoint = &c.CustomEndpointEc2
	}
	creds = credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     c.AccessKey,
					SecretAccessKey: c.SecretKey,
					SessionToken:    c.Token,
				},
			},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Profile: c.ProfileName,
			},
			defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
		})

	return config.WithCredentials(creds), nil
}

// Region returns the aws.Region object for access to AWS services, requesting
// the region from the instance metadata if possible.
func (c *AccessConfig) Region() (string, error) {
	if c.RawRegion != "" {
		if !c.SkipValidation {
			if valid := ValidateRegion(c.RawRegion); !valid {
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
		if valid := ValidateRegion(c.RawRegion); !valid {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
