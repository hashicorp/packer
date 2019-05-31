//go:generate struct-markdown

package docker

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/packer/builder/amazon/common"
)

type AwsAccessConfig struct {
	// The AWS access key used to communicate with
    // AWS. Learn how to set
    // this.
	AccessKey string `mapstructure:"aws_access_key" required:"false"`
	// The AWS secret key used to communicate with
    // AWS. Learn how to set
    // this.
	SecretKey string `mapstructure:"aws_secret_key" required:"false"`
	// The AWS access token to use. This is different from
    // the access key and secret key. If you're not sure what this is, then you
    // probably don't need it. This will also be read from the AWS_SESSION_TOKEN
    // environmental variable.
	Token     string `mapstructure:"aws_token" required:"false"`
	// The AWS shared credentials profile used to
    // communicate with AWS. Learn how to set
    // this.
	Profile   string `mapstructure:"aws_profile" required:"false"`
	cfg       *common.AccessConfig
}

// Get a login token for Amazon AWS ECR. Returns username and password
// or an error.
func (c *AwsAccessConfig) EcrGetLogin(ecrUrl string) (string, string, error) {

	exp := regexp.MustCompile(`(?:http://|https://|)([0-9]*)\.dkr\.ecr\.(.*)\.amazonaws\.com.*`)
	splitUrl := exp.FindStringSubmatch(ecrUrl)
	if len(splitUrl) != 3 {
		return "", "", fmt.Errorf("Failed to parse the ECR URL: %s it should be on the form <account number>.dkr.ecr.<region>.amazonaws.com", ecrUrl)
	}
	accountId := splitUrl[1]
	region := splitUrl[2]

	log.Println(fmt.Sprintf("Getting ECR token for account: %s in %s..", accountId, region))

	c.cfg = &common.AccessConfig{
		AccessKey:   c.AccessKey,
		ProfileName: c.Profile,
		RawRegion:   region,
		SecretKey:   c.SecretKey,
		Token:       c.Token,
	}

	session, err := c.cfg.Session()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %s", err)
	}

	service := ecr.New(session)

	params := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{
			aws.String(accountId),
		},
	}
	resp, err := service.GetAuthorizationToken(params)
	if err != nil {
		return "", "", fmt.Errorf(err.Error())
	}

	auth, err := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("Error decoding ECR AuthorizationToken: %s", err)
	}

	authParts := strings.SplitN(string(auth), ":", 2)
	log.Printf("Successfully got login for ECR: %s", ecrUrl)

	return authParts[0], authParts[1], nil
}
