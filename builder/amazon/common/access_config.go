package common

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/packer/template/interpolate"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	RawRegion string `mapstructure:"region"`
	Token     string `mapstructure:"token"`
}

// Auth returns a valid aws.Auth object for access to AWS services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (aws.Auth, error) {
	auth, err := aws.GetAuth(c.AccessKey, c.SecretKey)
	if err == nil {
		// Store the accesskey and secret that we got...
		c.AccessKey = auth.AccessKey
		c.SecretKey = auth.SecretKey
		c.Token = auth.Token
	}
	if c.Token != "" {
		auth.Token = c.Token
	}

	return auth, err
}

// Region returns the aws.Region object for access to AWS services, requesting
// the region from the instance metadata if possible.
func (c *AccessConfig) Region() (aws.Region, error) {
	if c.RawRegion != "" {
		return aws.Regions[c.RawRegion], nil
	}

	md, err := aws.GetMetaData("placement/availability-zone")
	if err != nil {
		return aws.Region{}, err
	}

	region := strings.TrimRightFunc(string(md), unicode.IsLetter)
	return aws.Regions[region], nil
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.RawRegion != "" {
		if _, ok := aws.Regions[c.RawRegion]; !ok {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
