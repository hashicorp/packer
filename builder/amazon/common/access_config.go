package common

import (
	"github.com/mitchellh/goamz/aws"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// Auth returns a valid aws.Auth object for access to AWS services, or
// an error if the authentication couldn't be resolved.
func (c *AccessConfig) Auth() (aws.Auth, error) {
	return aws.GetAuth(c.AccessKey, c.SecretKey)
}

func (c *AccessConfig) Prepare() []error {
	return nil
}
