package ucloudimport

import (
	"context"
	"fmt"
	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
	"log"
	"strings"
	"time"
)

const (
	BuilderId       = "packer.post-processor.ucloud-import"
	RAWFileFormat   = "raw"
	VHDFileFormat   = "vhd"
	VMDKFileFormat  = "vmdk"
	QCOW2FileFormat = "qcow2"

	CentOSOsType  = "CentOS"
	UbuntuOsType  = "Ubuntu"
	WindowsOsType = "Windows"
	RedHatOsType  = "RedHat"
	DebianOsType  = "Debian"
	OtherOsType   = "Other"
)

var regionForFileMap = ucloudcommon.NewStringConverter(map[string]string{
	"cn-bj2": "cn-bj",
	"cn-bj1": "cn-bj",
})

var imageFormatMap = ucloudcommon.NewStringConverter(map[string]string{
	"raw":  "RAW",
	"vhd":  "VHD",
	"vmdk": "VMDK",
})

// Configuration of this post processor
type Config struct {
	common.PackerConfig       `mapstructure:",squash"`
	ucloudcommon.AccessConfig `mapstructure:",squash"`

	// Variables specific to this post processor
	UFileBucket           string `mapstructure:"ufile_bucket_name"`
	UFileKey              string `mapstructure:"ufile_key_name"`
	SkipClean             bool   `mapstructure:"skip_clean"`
	ImageName             string `mapstructure:"image_name"`
	ImageDescription      string `mapstructure:"image_description"`
	OSType                string `mapstructure:"image_os_type"`
	OSName                string `mapstructure:"image_os_name"`
	Format                string `mapstructure:"format"`
	WaitImageReadyTimeout int    `mapstructure:"wait_image_ready_timeout"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

// Entry point for configuration parsing when we've defined
func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"ufile_key_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Set defaults
	if p.config.UFileKey == "" {
		p.config.UFileKey = "packer-import-{{timestamp}}." + p.config.Format
	}

	if p.config.WaitImageReadyTimeout <= 0 {
		p.config.WaitImageReadyTimeout = ucloudcommon.DefaultCreateImageTimeOut
	}

	errs := new(packer.MultiError)

	// Check and render ufile_key_name
	if err = interpolate.Validate(p.config.UFileKey, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing ufile_key_name template: %s", err))
	}

	// Check we have ucloud access variables defined somewhere
	errs = packer.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	// define all our required parameters
	templates := map[string]*string{
		"ufile_bucket_name": &p.config.UFileBucket,
		"image_name":        &p.config.ImageName,
		"image_os_type":     &p.config.OSType,
		"image_os_name":     &p.config.OSName,
		"format":            &p.config.Format,
	}
	// Check out required params are defined
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	imageName := p.config.ImageName
	if !ucloudcommon.ImageNamePattern.MatchString(imageName) {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("expected %q to be 1-63 characters and only support chinese, english, numbers, '-_,.:[]', got %q", "image_name", imageName))
	}

	switch p.config.OSType {
	case CentOSOsType, UbuntuOsType, WindowsOsType, RedHatOsType, DebianOsType, OtherOsType:
	default:
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("expected %q only be one of 'CentOS', 'Ubuntu', 'Windows', 'RedHat', 'Debian' or 'Other', got %q", "image_os_type", p.config.OSType))
	}

	switch p.config.Format {
	case VHDFileFormat, RAWFileFormat, VMDKFileFormat, QCOW2FileFormat:
	default:
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("expected %q only be one of 'raw', 'vhd', 'vmdk', or 'qcow2', got %q", "format", p.config.Format))
	}

	// Anything which flagged return back up the stack
	if len(errs.Errors) > 0 {
		return errs
	}

	packer.LogSecretFilter.Set(p.config.PublicKey, p.config.PrivateKey)
	log.Println(p.config)
	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	var err error

	client, err := p.config.Client()
	if err != nil {
		return nil, false, false, fmt.Errorf("Failed to connect ucloud client %s", err)
	}
	uhostconn := client.UHostConn
	ufileconn := client.UFileConn

	// Render this key since we didn't in the configure phase
	p.config.UFileKey, err = interpolate.Render(p.config.UFileKey, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error rendering ufile_key_name template: %s", err)
	}

	ui.Message(fmt.Sprintf("Rendered ufile_key_name as %s", p.config.UFileKey))

	ui.Message("Looking for image in artifact")
	// Locate the files output from the builder
	var source string
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, "."+p.config.Format) {
			source = path
			break
		}
	}

	// Hope we found something useful
	if source == "" {
		return nil, false, false, fmt.Errorf("No %s image file found in artifact from builder", p.config.Format)
	}

	convertedRegion := regionForFileMap.Convert(p.config.Region)
	keyName := p.config.UFileKey
	bucketName := p.config.UFileBucket

	config := &ufsdk.Config{
		PublicKey:  p.config.PublicKey,
		PrivateKey: p.config.PrivateKey,
		BucketName: bucketName,
		FileHost:   fmt.Sprintf(convertedRegion + ".ufileos.com"),
		BucketHost: "api.ucloud.cn",
	}

	// query or create bucket
	if err := queryOrCreateBucket(ufileconn, config); err != nil {
		return nil, false, false, fmt.Errorf("Failed to query or create bucket, %s", err)
	}

	bucketUrl := fmt.Sprintf("http://" + bucketName + "." + convertedRegion + ".ufileos.com")

	ui.Say(fmt.Sprintf("Waiting for uploading image file %s to UFile: %s/%s...", source, bucketUrl, p.config.UFileKey))

	// upload file to bucket
	ufileUrl, err := uploadFile(ufileconn, config, keyName, source)
	if err != nil {
		return nil, false, false, fmt.Errorf("Failed to Upload image file, %s", err)
	}

	ui.Say(fmt.Sprintf("Image file %s has been uploaded to UFile %s", source, ufileUrl))

	importImageRequest := p.buildImportImageRequest(uhostconn, ufileUrl)
	importImageResponse, err := uhostconn.ImportCustomImage(importImageRequest)
	if err != nil {
		return nil, false, false, fmt.Errorf("Failed to import image from %s/%s, %s", bucketUrl, p.config.UFileKey, err)
	}

	ui.Say(fmt.Sprintf("Waiting for importing %s/%s to ucloud...", bucketUrl, p.config.UFileKey))

	imageId := importImageResponse.ImageId
	err = retry.Config{
		StartTimeout: time.Duration(p.config.WaitImageReadyTimeout) * time.Second,
		ShouldRetry: func(err error) bool {
			return ucloudcommon.IsExpectedStateError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		image, err := client.DescribeImageById(imageId)
		if err != nil {
			return err
		}

		if image.State == ucloudcommon.ImageStateUnavailable {
			return fmt.Errorf("Unavailable importing image %q", imageId)
		}

		if image.State != ucloudcommon.ImageStateAvailable {
			return ucloudcommon.NewExpectedStateError("image", imageId)
		}

		return nil
	})

	if err != nil {
		return nil, false, false, fmt.Errorf("Error on waiting for importing image %q, %s",
			imageId, err.Error())
	}

	// Add the reported UCloud image ID to the artifact list
	ui.Say(fmt.Sprintf("Importing created ucloud image %q in region %q Complete.", imageId, p.config.Region))
	images := []ucloudcommon.ImageInfo{
		{
			ImageId:   imageId,
			ProjectId: p.config.ProjectId,
			Region:    p.config.Region,
		},
	}

	artifact = &ucloudcommon.Artifact{
		UCloudImages:   ucloudcommon.NewImageInfoSet(images),
		BuilderIdValue: BuilderId,
		Client:         client,
	}

	if !p.config.SkipClean {
		ui.Message(fmt.Sprintf("Deleting import source UFile: %s/%s", p.config.UFileBucket, p.config.UFileKey))
		if err = deleteFile(config, p.config.UFileKey); err != nil {
			return nil, false, false, fmt.Errorf("Failed to delete UFile: %s/%s, %s", p.config.UFileBucket, p.config.UFileKey, err)
		}
	}

	return artifact, false, false, nil
}

func (p *PostProcessor) buildImportImageRequest(conn *uhost.UHostClient, privateUrl string) *uhost.ImportCustomImageRequest {
	req := conn.NewImportCustomImageRequest()
	req.ImageName = ucloud.String(p.config.ImageName)
	req.ImageDescription = ucloud.String(p.config.ImageDescription)
	req.UFileUrl = ucloud.String(privateUrl)
	req.OsType = ucloud.String(p.config.OSType)
	req.OsName = ucloud.String(p.config.OSName)
	req.Format = ucloud.String(imageFormatMap.Convert(p.config.Format))
	req.Auth = ucloud.Bool(true)
	return req
}

func queryOrCreateBucket(conn *ufile.UFileClient, config *ufsdk.Config) error {
	var limit = 100
	var offset int
	var bucketList []ufile.UFileBucketSet
	for {
		req := conn.NewDescribeBucketRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeBucket(req)
		if err != nil {
			return fmt.Errorf("error on reading bucket list when create bucket, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		bucketList = append(bucketList, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	var bucketNames []string
	for _, v := range bucketList {
		bucketNames = append(bucketNames, v.BucketName)
	}

	if !ucloudcommon.IsStringIn(config.BucketName, bucketNames) {
		req := conn.NewCreateBucketRequest()
		req.BucketName = ucloud.String(config.BucketName)
		req.Type = ucloud.String("private")

		_, err := conn.CreateBucket(req)
		if err != nil {
			return fmt.Errorf("error on creating bucket %s, %s", config.BucketName, err)
		}
	}

	return nil
}

func uploadFile(conn *ufile.UFileClient, config *ufsdk.Config, keyName, source string) (string, error) {
	reqFile, err := ufsdk.NewFileRequest(config, nil)
	if err != nil {
		return "", fmt.Errorf("error on building upload file request, %s", err)
	}

	// upload file in segments
	err = reqFile.AsyncMPut(source, keyName, "")
	if err != nil {
		return "", fmt.Errorf("error on upload file, %s, details: %s", err, reqFile.DumpResponse(true))
	}

	reqBucket := conn.NewDescribeBucketRequest()
	reqBucket.BucketName = ucloud.String(config.BucketName)
	resp, err := conn.DescribeBucket(reqBucket)
	if err != nil {
		return "", fmt.Errorf("error on reading bucket list when upload file, %s", err)
	}

	if resp.DataSet[0].Type == "private" {
		return reqFile.GetPrivateURL(keyName, 24*60*60), nil
	}

	return reqFile.GetPublicURL(keyName), nil
}

func deleteFile(config *ufsdk.Config, keyName string) error {
	req, err := ufsdk.NewFileRequest(config, nil)
	if err != nil {
		return fmt.Errorf("error on new deleting file, %s", err)
	}
	req.DeleteFile(keyName)
	if err != nil {
		return fmt.Errorf("error on deleting file, %s", err)
	}

	return nil
}
