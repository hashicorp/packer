// Package secretsmanager provide methods to get data from
// AWS Secret Manager
package secretsmanager

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// Client represents an AWS Secrets Manager client
type Client struct {
	config *AWSConfig
	api    secretsmanageriface.SecretsManagerAPI
}

// New creates an AWS Session Manager Client
func New(config *AWSConfig) *Client {
	c := &Client{
		config: config,
	}

	s := c.newSession(config)
	c.api = secretsmanager.New(s)
	return c
}

func (c *Client) newSession(config *AWSConfig) *session.Session {
	// Initialize config with error verbosity
	sess := aws.NewConfig().WithCredentialsChainVerboseErrors(true)

	if config.Region != "" {
		sess = sess.WithRegion(config.Region)
	}

	opts := session.Options{
		Config: *sess,
	}

	return session.Must(session.NewSessionWithOptions(opts))
}

// GetSecret return an AWS Secret Manager secret
// in plain text from a given secret name
func (c *Client) GetSecret(spec *SecretSpec) (string, error) {
	params := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(spec.Name),
		VersionStage: aws.String("AWSCURRENT"),
	}

	resp, err := c.api.GetSecretValue(params)
	if err != nil {
		return "", err
	}

	if resp.SecretString == nil {
		return "", errors.New("Secret is not string")
	}

	secret := SecretString{
		Name:         *resp.Name,
		SecretString: *resp.SecretString,
	}
	value, err := getSecretValue(&secret, spec)
	if err != nil {
		return "", err
	}

	return value, nil
}

func getSecretValue(s *SecretString, spec *SecretSpec) (string, error) {
	var secretValue map[string]string

	blob := []byte(s.SecretString)

	err := json.Unmarshal(blob, &secretValue)
	if err != nil {
		return "", err
	}

	// If key is not set then return first value stored in secret
	if spec.Key == "" {
		for _, v := range secretValue {
			return v, nil
		}
	}

	if v, ok := secretValue[spec.Key]; ok {
		return v, nil
	}

	return "", errors.New("No secret found")
}
