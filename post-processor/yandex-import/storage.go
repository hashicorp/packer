package yandeximport

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const defaultS3Region = "ru-central1"
const defaultStorageEndpoint = "storage.yandexcloud.net"

func newYCStorageClient(storageEndpoint, accessKey, secretKey string) (*s3.S3, error) {
	var creds *credentials.Credentials

	if storageEndpoint == "" {
		storageEndpoint = defaultStorageEndpoint
	}

	s3Config := &aws.Config{
		Endpoint: aws.String(storageEndpoint),
		Region:   aws.String(defaultS3Region),
	}

	switch {
	case accessKey != "" && secretKey != "":
		creds = credentials.NewStaticCredentials(accessKey, secretKey, "")
	default:
		return nil, fmt.Errorf("either access or secret key not provided")
	}

	s3Config.Credentials = creds
	newSession, err := session.NewSession(s3Config)

	if err != nil {
		return nil, err
	}

	return s3.New(newSession), nil
}
