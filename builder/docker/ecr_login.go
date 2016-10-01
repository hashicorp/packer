package docker

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type AwsAccessConfig struct {
	AccessKey string `mapstructure:"aws_access_key"`
	SecretKey string `mapstructure:"aws_secret_key"`
	Token     string `mapstructure:"aws_token"`
}

// Config returns a valid aws.Config object for access to AWS services, or
// an error if the authentication and region couldn't be resolved
func (c *AwsAccessConfig) config(region string) (*aws.Config, error) {
	var creds *credentials.Credentials

	config := aws.NewConfig().WithRegion(region).WithMaxRetries(11)
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
	return config.WithCredentials(creds), nil
}

// Get a login token for Amazon AWS ECR. Returns username and password
// or an error.
func (c *AwsAccessConfig) EcrGetLogin(ecrUrl string) (string, string, error) {

	exp := regexp.MustCompile("(?:http://|https://|)([0-9]*)\\.dkr\\.ecr\\.(.*)\\.amazonaws\\.com.*")
	splitUrl := exp.FindStringSubmatch(ecrUrl)
	accountId := splitUrl[1]
	region := splitUrl[2]

	log.Println(fmt.Sprintf("Getting ECR token for account: %s in %s..", accountId, region))

	awsConfig, err := c.config(region)
	if err != nil {
		return "", "", err
	}

	session, err := session.NewSession(awsConfig)
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
