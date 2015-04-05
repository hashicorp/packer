package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/mitchellh/packer/packer"
)

// AccessConfig is for common configuration related to AWS access
type AccessConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	RawRegion string `mapstructure:"region"`
	Token     string `mapstructure:"token"`
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AccessConfig) Config() (*aws.Config, error) {
	credsProvider := aws.DetectCreds(c.AccessKey, c.SecretKey, c.Token)

	creds, err := credsProvider.Credentials()
	if err != nil {
		return nil, err
	}

	c.AccessKey = creds.AccessKeyID
	c.SecretKey = creds.SecretAccessKey
	c.Token = creds.SessionToken

	region, err := c.Region()
	if err != nil {
		return nil, err
	}

	return &aws.Config{
		Region:      region,
		Credentials: credsProvider,
	}, nil
}

// Region returns the aws.Region object for access to AWS services, requesting
// the region from the instance metadata if possible.
func (c *AccessConfig) Region() (string, error) {
	if c.RawRegion != "" {
		if valid := ValidateRegion(c.RawRegion); valid == false {
			return "", fmt.Errorf("Not a valid region: %s", c.RawRegion)
		}
		return c.RawRegion, nil
	}

	md, err := GetInstanceMetaData("placement/availability-zone")
	if err != nil {
		return "", err
	}

	region := strings.TrimRightFunc(string(md), unicode.IsLetter)
	return region, nil
}

func (c *AccessConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
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
		if valid := ValidateRegion(c.RawRegion); valid == false {
			errs = append(errs, fmt.Errorf("Unknown region: %s", c.RawRegion))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func GetInstanceMetaData(path string) (contents []byte, err error) {
	url := "http://169.254.169.254/latest/meta-data/" + path

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Code %d returned for url %s", resp.StatusCode, url)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return []byte(body), err
}
