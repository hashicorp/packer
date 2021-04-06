//go:generate struct-markdown

package docker

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/go-cleanhttp"
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
	Token string `mapstructure:"aws_token" required:"false"`
	// The AWS shared credentials profile used to
	// communicate with AWS. Learn how to set
	// this.
	Profile string `mapstructure:"aws_profile" required:"false"`
	cfg     *awsbase.Config
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

	// Create new AWS config
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	config = config.WithRegion(region)

	config = config.WithHTTPClient(cleanhttp.DefaultClient())
	transport := config.HTTPClient.Transport.(*http.Transport)
	transport.Proxy = http.ProxyFromEnvironment

	// Figure out which possible credential providers are valid; test that we
	// can get credentials via the selected providers, and set the providers in
	// the config.
	creds, err := c.GetCredentials(config)
	if err != nil {
		return "", "", fmt.Errorf(err.Error())
	}
	config.WithCredentials(creds)

	// Create session options based on our AWS config
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *config,
	}

	if c.Profile != "" {
		opts.Profile = c.Profile
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return "", "", err
	}
	log.Printf("Found region %s", *sess.Config.Region)
	session := sess

	cp, err := session.Config.Credentials.Get()

	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %s", err)
	}

	log.Printf("[INFO] AWS authentication used: %q", cp.ProviderName)

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

// GetCredentials gets credentials from the environment, shared credentials,
// the session (which may include a credential process), or ECS/EC2 metadata
// endpoints. GetCredentials also validates the credentials and the ability to
// assume a role or will return an error if unsuccessful.
func (c *AwsAccessConfig) GetCredentials(config *aws.Config) (*awsCredentials.Credentials, error) {
	// Reload values into the config used by the Packer-Terraform shared SDK
	awsbaseConfig := &awsbase.Config{
		AccessKey:    c.AccessKey,
		DebugLogging: false,
		Profile:      c.Profile,
		SecretKey:    c.SecretKey,
		Token:        c.Token,
	}

	return awsbase.GetCredentials(awsbaseConfig)
}
