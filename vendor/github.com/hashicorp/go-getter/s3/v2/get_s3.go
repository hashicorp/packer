package s3

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hashicorp/go-getter/v2"
)

// Getter is a Getter implementation that will download a module from
// a S3 bucket.
type Getter struct{}

func (g *Getter) Mode(ctx context.Context, u *url.URL) (getter.Mode, error) {
	// Parse URL
	region, bucket, path, _, creds, err := g.parseUrl(u)
	if err != nil {
		return 0, err
	}

	// Create client config
	config := g.getAWSConfig(region, u, creds)
	sess := session.New(config)
	client := s3.New(sess)

	// List the object(s) at the given prefix
	req := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	}
	resp, err := client.ListObjects(req)
	if err != nil {
		return 0, err
	}

	for _, o := range resp.Contents {
		// Use file mode on exact match.
		if *o.Key == path {
			return getter.ModeFile, nil
		}

		// Use dir mode if child keys are found.
		if strings.HasPrefix(*o.Key, path+"/") {
			return getter.ModeDir, nil
		}
	}

	// There was no match, so just return file mode. The download is going
	// to fail but we will let S3 return the proper error later.
	return getter.ModeFile, nil
}

func (g *Getter) Get(ctx context.Context, req *getter.Request) error {

	// Parse URL
	region, bucket, path, _, creds, err := g.parseUrl(req.URL())
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	config := g.getAWSConfig(region, req.URL(), creds)
	sess := session.New(config)
	client := s3.New(sess)

	// List files in path, keep listing until no more objects are found
	lastMarker := ""
	hasMore := true
	for hasMore {
		s3Req := &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(path),
		}
		if lastMarker != "" {
			s3Req.Marker = aws.String(lastMarker)
		}

		resp, err := client.ListObjects(s3Req)
		if err != nil {
			return err
		}

		hasMore = aws.BoolValue(resp.IsTruncated)

		// Get each object storing each file relative to the destination path
		for _, object := range resp.Contents {
			lastMarker = aws.StringValue(object.Key)
			objPath := aws.StringValue(object.Key)

			// If the key ends with a backslash assume it is a directory and ignore
			if strings.HasSuffix(objPath, "/") {
				continue
			}

			// Get the object destination path
			objDst, err := filepath.Rel(path, objPath)
			if err != nil {
				return err
			}
			objDst = filepath.Join(req.Dst, objDst)

			if err := g.getObject(ctx, client, objDst, bucket, objPath, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Getter) GetFile(ctx context.Context, req *getter.Request) error {
	region, bucket, path, version, creds, err := g.parseUrl(req.URL())
	if err != nil {
		return err
	}

	config := g.getAWSConfig(region, req.URL(), creds)
	sess := session.New(config)
	client := s3.New(sess)
	return g.getObject(ctx, client, req.Dst, bucket, path, version)
}

func (g *Getter) getObject(ctx context.Context, client *s3.S3, dst, bucket, key, version string) error {
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		req.VersionId = aws.String(version)
	}

	resp, err := client.GetObject(req)
	if err != nil {
		return err
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = getter.Copy(ctx, f, resp.Body)
	return err
}

func (g *Getter) getAWSConfig(region string, url *url.URL, creds *credentials.Credentials) *aws.Config {
	conf := &aws.Config{}
	if creds == nil {
		// Grab the metadata URL
		metadataURL := os.Getenv("AWS_METADATA_URL")
		if metadataURL == "" {
			metadataURL = "http://169.254.169.254:80/latest"
		}

		creds = credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
				&ec2rolecreds.EC2RoleProvider{
					Client: ec2metadata.New(session.New(&aws.Config{
						Endpoint: aws.String(metadataURL),
					})),
				},
			})
	}

	if creds != nil {
		conf.Endpoint = &url.Host
		conf.S3ForcePathStyle = aws.Bool(true)
		if url.Scheme == "http" {
			conf.DisableSSL = aws.Bool(true)
		}
	}

	conf.Credentials = creds
	if region != "" {
		conf.Region = aws.String(region)
	}

	return conf
}

func (g *Getter) parseUrl(u *url.URL) (region, bucket, path, version string, creds *credentials.Credentials, err error) {
	// This just check whether we are dealing with S3 or
	// any other S3 compliant service. S3 has a predictable
	// url as others do not
	if strings.Contains(u.Host, "amazonaws.com") {
		// Expected host style: s3.amazonaws.com. They always have 3 parts,
		// although the first may differ if we're accessing a specific region.
		hostParts := strings.Split(u.Host, ".")
		if len(hostParts) != 3 {
			err = fmt.Errorf("URL is not a valid S3 URL")
			return
		}

		// Parse the region out of the first part of the host
		region = strings.TrimPrefix(strings.TrimPrefix(hostParts[0], "s3-"), "s3")
		if region == "" {
			region = "us-east-1"
		}

		pathParts := strings.SplitN(u.Path, "/", 3)
		if len(pathParts) != 3 {
			err = fmt.Errorf("URL is not a valid S3 URL")
			return
		}

		bucket = pathParts[1]
		path = pathParts[2]
		version = u.Query().Get("version")

	} else {
		pathParts := strings.SplitN(u.Path, "/", 3)
		if len(pathParts) != 3 {
			err = fmt.Errorf("URL is not a valid S3 complaint URL")
			return
		}
		bucket = pathParts[1]
		path = pathParts[2]
		version = u.Query().Get("version")
		region = u.Query().Get("region")
		if region == "" {
			region = "us-east-1"
		}
	}

	_, hasAwsId := u.Query()["aws_access_key_id"]
	_, hasAwsSecret := u.Query()["aws_access_key_secret"]
	_, hasAwsToken := u.Query()["aws_access_token"]
	if hasAwsId || hasAwsSecret || hasAwsToken {
		creds = credentials.NewStaticCredentials(
			u.Query().Get("aws_access_key_id"),
			u.Query().Get("aws_access_key_secret"),
			u.Query().Get("aws_access_token"),
		)
	}

	return
}

func (g *Getter) Detect(req *getter.Request) (bool, error) {
	src := req.Src
	if len(src) == 0 {
		return false, nil
	}

	if req.Forced != "" {
		// There's a getter being forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the forced one
			// Don't use it to try to download the artifact
			return false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return false, nil
	}

	src, ok, err := new(Detector).Detect(src, req.Pwd)
	if err != nil {
		return ok, err
	}
	if ok {
		req.Src = src
	}

	return ok, nil
}

func (g *Getter) validScheme(scheme string) bool {
	return scheme == "s3"
}
