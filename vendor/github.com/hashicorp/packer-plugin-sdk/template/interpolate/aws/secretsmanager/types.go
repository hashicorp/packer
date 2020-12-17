package secretsmanager

// AWSConfig store configuration used to initialize
// secrets manager client.
type AWSConfig struct {
	Region string
}

// SecretSpec represent specs of secret to be searched
// If Key field is not set then package will return first
// secret key stored in secret name.
type SecretSpec struct {
	Name string
	Key  string
}

// SecretString is a concret representation
// of an AWS Secrets Manager Secret String
type SecretString struct {
	Name         string
	SecretString string
}
