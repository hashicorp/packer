// Package secretsmanager provide methods to get data from
// AWS Secret Manager
package secretsmanager

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// New creates an AWS Session Manager Client
func New() (*Secret, error) {
	sess := session.Must(session.NewSession())

	var c *aws.Config
	s := Secret{
		Client: secretsmanager.New(sess, c),
	}
	return &s, nil
}

// GetSecret return an AWS Secret Manager secret
// in plain text from a given secret name
func (s *Secret) GetSecret(name string) (string, error) {
	params := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	}

	resp, err := s.Client.GetSecretValue(params)
	if err != nil {
		return "", err
	}

	if resp.SecretString == nil {
		return "", errors.New("Secret is not string")
	}

	secret := SecretString{
		Name:   *resp.Name,
		Secret: *resp.SecretString,
	}
	value, err := getSecretValue(&secret)
	if err != nil {
		return "", err
	}

	return value, nil
}

func getSecretValue(s *SecretString) (string, error) {
	var secretValue map[string]string

	blob := []byte(s.Secret)

	err := json.Unmarshal(blob, &secretValue)
	if err != nil {
		return "", err
	}

	for _, v := range secretValue {
		return v, nil
	}
	return "", errors.New("Secret not found")
}
