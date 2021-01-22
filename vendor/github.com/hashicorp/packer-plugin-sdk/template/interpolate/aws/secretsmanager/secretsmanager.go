// Package secretsmanager provide methods to get data from
// AWS Secret Manager
package secretsmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	sessConfig := aws.NewConfig().WithCredentialsChainVerboseErrors(true)

	if config.Region != "" {
		sessConfig = sessConfig.WithRegion(config.Region)
	}

	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *sessConfig,
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
	var secretValue map[string]interface{}
	blob := []byte(s.SecretString)

	//For those plaintext secrets just return the value
	if json.Valid(blob) != true {
		return s.SecretString, nil
	}

	err := json.Unmarshal(blob, &secretValue)
	if err != nil {
		return "", err
	}

	// If key is not set and secret has multiple keys, return error
	if spec.Key == "" && len(secretValue) > 1 {
		return "", errors.New("Secret has multiple values and no key was set")
	}

	if spec.Key == "" {
		for _, v := range secretValue {
			return getStringSecretValue(v)
		}
	}

	if v, ok := secretValue[spec.Key]; ok {
		return getStringSecretValue(v)
	}

	return "", fmt.Errorf("No secret found for key %q", spec.Key)
}

func getStringSecretValue(v interface{}) (string, error) {
	switch valueType := v.(type) {
	case string:
		return valueType, nil
	case float64:
		return strconv.FormatFloat(valueType, 'f', 0, 64), nil
	default:
		return "", fmt.Errorf("Unsupported secret value type: %T", valueType)
	}
}
