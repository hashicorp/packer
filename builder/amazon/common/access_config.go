package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
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
		sess := session.New(config)
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.StaticProvider{Value: credentials.Value{
				AccessKeyID:     c.AccessKey,
				SecretAccessKey: c.SecretKey,
				SessionToken:    c.Token,
			}},
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})
	}
	return config.WithCredentials(creds), nil
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

	md, err := GetInstanceMetaData("placement/availability-zone")
	if err != nil {
		return "", err
	}

	region := strings.TrimRightFunc(string(md), unicode.IsLetter)
	return region, nil
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.RawRegion != "" && !c.SkipValidation {
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
