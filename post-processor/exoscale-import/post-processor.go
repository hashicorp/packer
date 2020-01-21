//go:generate mapstructure-to-hcl2 -type Config

package exoscaleimport

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/exoscale/egoscale"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/builder/qemu"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/version"
)

var (
	defaultTemplateZone = "ch-gva-2"
	defaultAPIEndpoint  = "https://api.exoscale.com/compute"
	defaultSOSEndpoint  = "https://sos-" + defaultTemplateZone + ".exo.io"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	SkipClean           bool `mapstructure:"skip_clean"`

	SOSEndpoint             string `mapstructure:"sos_endpoint"`
	APIEndpoint             string `mapstructure:"api_endpoint"`
	APIKey                  string `mapstructure:"api_key"`
	APISecret               string `mapstructure:"api_secret"`
	ImageBucket             string `mapstructure:"image_bucket"`
	TemplateZone            string `mapstructure:"template_zone"`
	TemplateName            string `mapstructure:"template_name"`
	TemplateDescription     string `mapstructure:"template_description"`
	TemplateUsername        string `mapstructure:"template_username"`
	TemplateDisablePassword bool   `mapstructure:"template_disable_password"`
	TemplateDisableSSHKey   bool   `mapstructure:"template_disable_sshkey"`
}

func init() {
	egoscale.UserAgent = "Packer-Exoscale/" + version.FormattedVersion() + " " + egoscale.UserAgent
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.config.TemplateZone = defaultTemplateZone
	p.config.APIEndpoint = defaultAPIEndpoint
	p.config.SOSEndpoint = defaultSOSEndpoint

	if err := config.Decode(&p.config, nil, raws...); err != nil {
		return err
	}

	if p.config.APIKey == "" {
		p.config.APIKey = os.Getenv("EXOSCALE_API_KEY")
	}

	if p.config.APISecret == "" {
		p.config.APISecret = os.Getenv("EXOSCALE_API_SECRET")
	}

	requiredArgs := map[string]*string{
		"api_key":              &p.config.APIKey,
		"api_secret":           &p.config.APISecret,
		"api_endpoint":         &p.config.APIEndpoint,
		"sos_endpoint":         &p.config.SOSEndpoint,
		"image_bucket":         &p.config.ImageBucket,
		"template_zone":        &p.config.TemplateZone,
		"template_name":        &p.config.TemplateName,
		"template_description": &p.config.TemplateDescription,
	}

	errs := new(packer.MultiError)
	for k, v := range requiredArgs {
		if *v == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", k))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	packer.LogSecretFilter.Set(p.config.APIKey, p.config.APISecret)

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, bool, error) {
	switch a.BuilderId() {
	case qemu.BuilderId, file.BuilderId, artifice.BuilderId:
		break

	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from QEMU/file builders and Artifice post-processor artifacts.",
			a.BuilderId())
		return nil, false, false, err
	}

	ui.Message("Uploading template image")
	url, md5sum, err := p.uploadImage(ctx, ui, a)
	if err != nil {
		return nil, false, false, fmt.Errorf("unable to upload image: %s", err)
	}

	ui.Message("Registering template")
	id, err := p.registerTemplate(ctx, ui, url, md5sum)
	if err != nil {
		return nil, false, false, fmt.Errorf("unable to register template: %s", err)
	}

	if !p.config.SkipClean {
		ui.Message("Deleting uploaded template image")
		if err = p.deleteImage(ctx, ui, a); err != nil {
			return nil, false, false, fmt.Errorf("unable to delete uploaded template image: %s", err)
		}
	}

	return &Artifact{id}, false, false, nil
}

func (p *PostProcessor) uploadImage(ctx context.Context, ui packer.Ui, a packer.Artifact) (string, string, error) {
	var (
		imageFile  = a.Files()[0]
		bucketFile = filepath.Base(imageFile)
	)

	f, err := os.Open(imageFile)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return "", "", err
	}

	// For tracking image file upload progress
	pf := ui.TrackProgress(imageFile, 0, fileInfo.Size(), f)
	defer pf.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", "", fmt.Errorf("image checksumming failed: %s", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", "", err
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{
		Region:      aws.String(p.config.TemplateZone),
		Endpoint:    aws.String(p.config.SOSEndpoint),
		Credentials: credentials.NewStaticCredentials(p.config.APIKey, p.config.APISecret, "")}}))

	uploader := s3manager.NewUploader(sess)
	output, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Body:       pf,
		Bucket:     aws.String(p.config.ImageBucket),
		Key:        aws.String(bucketFile),
		ContentMD5: aws.String(base64.StdEncoding.EncodeToString(hash.Sum(nil))),
		ACL:        aws.String("public-read"),
	})
	if err != nil {
		return "", "", err
	}

	return output.Location, fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (p *PostProcessor) deleteImage(ctx context.Context, ui packer.Ui, a packer.Artifact) error {
	var (
		imageFile  = a.Files()[0]
		bucketFile = filepath.Base(imageFile)
	)

	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{
		Region:      aws.String(p.config.TemplateZone),
		Endpoint:    aws.String(p.config.SOSEndpoint),
		Credentials: credentials.NewStaticCredentials(p.config.APIKey, p.config.APISecret, "")}}))

	svc := s3.New(sess)
	if _, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.config.ImageBucket),
		Key:    aws.String(bucketFile),
	}); err != nil {
		return err
	}

	return nil
}

func (p *PostProcessor) registerTemplate(ctx context.Context, ui packer.Ui, url, md5sum string) (string, error) {
	var (
		passwordEnabled = !p.config.TemplateDisablePassword
		sshkeyEnabled   = !p.config.TemplateDisableSSHKey
		regErr          error
	)

	exo := egoscale.NewClient(p.config.APIEndpoint, p.config.APIKey, p.config.APISecret)
	exo.RetryStrategy = egoscale.FibonacciRetryStrategy

	zone := egoscale.Zone{Name: p.config.TemplateZone}
	if resp, err := exo.GetWithContext(ctx, &zone); err != nil {
		return "", fmt.Errorf("template zone lookup failed: %s", err)
	} else {
		zone.ID = resp.(*egoscale.Zone).ID
	}

	req := egoscale.RegisterCustomTemplate{
		URL:             url,
		ZoneID:          zone.ID,
		Name:            p.config.TemplateName,
		Displaytext:     p.config.TemplateDescription,
		PasswordEnabled: &passwordEnabled,
		SSHKeyEnabled:   &sshkeyEnabled,
		Details:         map[string]string{"username": p.config.TemplateUsername},
		Checksum:        md5sum,
	}

	res := make([]egoscale.Template, 0)

	exo.AsyncRequestWithContext(ctx, req, func(jobRes *egoscale.AsyncJobResult, err error) bool {
		if err != nil {
			regErr = fmt.Errorf("request failed: %s", err)
			return false
		} else if jobRes.JobStatus == egoscale.Pending {
			// Job is not completed yet
			ui.Message("template registration in progress")
			return true
		}

		if err := jobRes.Result(&res); err != nil {
			regErr = err
			return false
		}

		if len(res) != 1 {
			regErr = fmt.Errorf("unexpected response from API (expected 1 item, got %d)", len(res))
			return false
		}

		return false
	})
	if regErr != nil {
		return "", regErr
	}

	return res[0].ID.String(), nil
}
