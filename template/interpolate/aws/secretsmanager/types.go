package secretsmanager

import (
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// AWSConfig store configuration used to initialize
// secrets manager client.
type AWSConfig struct {
	Region string
}

// SecretSpec represent specs of secret to be searched
// If Key field is not set then package will return first
// secret key stored in secret name.
//
// maps to ClusterConfig
type SecretSpec struct {
	Name string
	Key  string
}

// Client represents an AWS Secrets Manager client
//
// maps to ProviderServices
type Client struct {
	config *AWSConfig
	api    secretsmanageriface.SecretsManagerAPI
}

// SecretString is a concret representation
// of an AWS Secrets Manager Secret String
type SecretString struct {
	Name         string
	SecretString string
}
