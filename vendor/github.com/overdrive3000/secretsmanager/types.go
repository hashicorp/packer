package secretsmanager

import (
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// Secret represents an AWS Secrets Manager
// client
type Secret struct {
	Client secretsmanageriface.SecretsManagerAPI
}

// SecretString is a concret representation
// of an AWS Secrets Manager Secret String
type SecretString struct {
	Name   string
	Secret string
}
