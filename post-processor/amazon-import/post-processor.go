package amazonimport

import (
	"fmt"
	"strings"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
// This is bad, it should be pulled out into a common folder across
// both builders and post-processors
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.post-processor.amazon-import"

// We accept the output from vmware or vmware-esx
var builtins = map[string]string{
	"mitchellh.vmware":	"amazon-import",
	"mitchellh.vmware-esx":	"amazon-import",
}

// Configuration of this post processor
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`

// Variables specific to this post processor
	S3Bucket	string `mapstructure:"s3_bucket_name"`
	S3Key		string `mapstructure:"s3_key_name"`
	ImportTaskDesc	string `mapstructure:"import_task_desc"`
	ImportDiskDesc	string `mapstructure:"import_disk_desc"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

// Entry point for configuration parisng when we've defined
func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:		true,
		InterpolateContext:	&p.config.ctx,
		InterpolateFilter:	&interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Set defaults
	if p.config.ImportTaskDesc == "" {
		p.config.ImportTaskDesc = "packer-amazon-ova task"
	}
	if p.config.ImportDiskDesc == "" {
		p.config.ImportDiskDesc = "packer-amazon-ova disk"
	}

	errs := new(packer.MultiError)

	// Check we have AWS access variables defined somewhere
	errs = packer.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	// define all our required paramaters
	templates := map[string]*string{
		"s3_bucket_name":	&p.config.S3Bucket,
		"s3_key_name":		&p.config.S3Key,
	}
	// Check out required params are defined
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	// Anything which flagged return back up the stack
	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error

	config, err := p.config.Config()
	if err != nil {
		return nil, false, err
	}
	// Confirm we're dealing with the result of a builder we like
	if _, ok := builtins[artifact.BuilderId()]; !ok {
		return nil, false, fmt.Errorf("Artifact type %s is not supported by this post-processor", artifact.BuilderId())
	}

	// Locate the files output from the builder
	source := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".ova") {
			source = path
			break
		}
	}

	// Hope we found something useful
	if source == "" {
		return nil, false, fmt.Errorf("OVA file not found")
	}

	// Set up the AWS session
	session := session.New(config)

	// open the source file
	file, err := os.Open(source)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to open %s: %s", source, err)
	}

	ui.Message(fmt.Sprintf("Uploading %s to s3://%s/%s", source, p.config.S3Bucket, p.config.S3Key))

	// Copy the OVA file into the S3 bucket specified 
	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:	file,
		Bucket:	&p.config.S3Bucket,
		Key:	&p.config.S3Key,
	})
	if err != nil {
		return nil, false, fmt.Errorf("Failed to upload %s: %s", source, err)
	}

	ui.Message(fmt.Sprintf("Completed upload of %s to s3://%s/%s", source, p.config.S3Bucket, p.config.S3Key))

	// Call EC2 image import process
	ec2conn := ec2.New(session)
	import_start, err := ec2conn.ImportImage(&ec2.ImportImageInput{
		Description:	&p.config.ImportTaskDesc,
		DiskContainers: []*ec2.ImageDiskContainer{
			{
				Description: &p.config.ImportDiskDesc,
				UserBucket: &ec2.UserBucket{
					S3Bucket:	&p.config.S3Bucket,
					S3Key:		&p.config.S3Key,
				},
			},
		},
	})

	if err != nil {
		return nil, false, fmt.Errorf("Failed to start import from s3://%s/%s: %s", p.config.S3Bucket, p.config.S3Key, err)
	}

	ui.Message(fmt.Sprintf("Started import of s3://%s/%s, task id %s", p.config.S3Bucket, p.config.S3Key, *import_start.ImportTaskId))

	// Wait for import process to complete, this takess a while
	ui.Message(fmt.Sprintf("Waiting for task %s to complete (may take a while)", *import_start.ImportTaskId))

	stateChange := awscommon.StateChangeConf{
		Pending: []string{"pending","active"},
		Refresh: awscommon.ImportImageRefreshFunc(ec2conn, *import_start.ImportTaskId),
		Target: "completed",
	}
	_, err = awscommon.WaitForState(&stateChange)

	if err != nil {
		return nil, false, fmt.Errorf("Import task %s failed: %s", *import_start.ImportTaskId, err)
	}

	// Extract the AMI ID and return this as the artifact of the
	// post processor
	import_result, err := ec2conn.DescribeImportImageTasks(&ec2.DescribeImportImageTasksInput{
		ImportTaskIds:	[]*string{
				import_start.ImportTaskId,
		},
	})

	if err != nil {
		return nil, false, fmt.Errorf("API error for import task id %s: %s", *import_start.ImportTaskId, err)
	}

	// Add the discvered AMI ID to the artifact list
	artifact = &awscommon.Artifact{
		Amis:		map[string]string{
				*config.Region: *import_result.ImportImageTasks[0].ImageId,
		},
		BuilderIdValue:	BuilderId,
		Conn:		ec2conn,
	}

	return artifact, false, nil
}


