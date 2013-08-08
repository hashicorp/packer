package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/packer/common"
	"strings"
	"unicode"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	RawRegion string `mapstructure:"region"`
}

// Auth returns a valid aws.Auth object for access to AWS services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (aws.Auth, error) {
	return aws.GetAuth(c.AccessKey, c.SecretKey)
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

func (c *AccessConfig) Prepare(t *common.Template) []error {
	if t == nil {
		var err error
		t, err = common.NewTemplate()
		if err != nil {
			return []error{err}
		}
	}

	templates := map[string]*string{
		"access_key": &c.AccessKey,
		"secret_key": &c.SecretKey,
		"region":     &c.RawRegion,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

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
