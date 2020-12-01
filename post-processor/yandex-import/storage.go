package yandeximport

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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

// Get path-style S3 URL and return presigned URL
func presignUrl(s3conn *s3.S3, ui packersdk.Ui, fullUrl string) (cloudImageSource, error) {
	bucket, key, err := s3URLToBucketKey(fullUrl)
	if err != nil {
		return nil, err
	}

	req, _ := s3conn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	// Compute service allow only `https://storage.yandexcloud.net/...` URLs for Image create process
	req.Config.S3ForcePathStyle = aws.Bool(true)

	urlStr, _, err := req.PresignRequest(30 * time.Minute)
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to presign url: %s", err))
		return nil, err
	}

	return &objectSource{
		urlStr,
	}, nil
}

func s3URLToBucketKey(storageURL string) (bucket string, key string, err error) {
	u, err := url.Parse(storageURL)
	if err != nil {
		return
	}

	if u.Scheme == "s3" {
		// s3://bucket/key
		bucket = u.Host
		key = strings.TrimLeft(u.Path, "/")
	} else if u.Scheme == "https" {
		// https://***.storage.yandexcloud.net/...
		if u.Host == defaultStorageEndpoint {
			// No bucket name in the host part
			path := strings.SplitN(u.Path, "/", 3)
			bucket = path[1]
			key = path[2]
		} else {
			// Bucket name in host
			bucket = strings.TrimSuffix(u.Host, "."+defaultStorageEndpoint)
			key = strings.TrimLeft(u.Path, "/")
		}
	}
	return
}
