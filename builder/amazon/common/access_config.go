package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/mitchellh/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey   string `mapstructure:"access_key"`
	SecretKey   string `mapstructure:"secret_key"`
	RawRegion   string `mapstructure:"region"`
	Token       string `mapstructure:"token"`
	ProfileName string `mapstructure:"profile"`
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
	if c.ProfileName != "" {
		profile, err := NewFromProfile(c.ProfileName)
		if err != nil {
			return nil, err
		}
		creds, err = profile.CredentialsFromProfile(config)
		if err != nil {
			return nil, err
		}
	} else {
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.StaticProvider{Value: credentials.Value{
				AccessKeyID:     c.AccessKey,
				SecretAccessKey: c.SecretKey,
				SessionToken:    c.Token,
			}},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
			&ec2rolecreds.EC2RoleProvider{},
		})
	}
	return config.WithCredentials(creds), nil
}

func (c *AccessConfig) Region() (string, error) {
	client := ec2metadata.New(&ec2metadata.Config{})
	return client.Region()
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.RawRegion != "" {
		if valid := ValidateRegion(c.RawRegion); valid == false {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
