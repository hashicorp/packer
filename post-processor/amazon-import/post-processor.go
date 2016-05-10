package amazonimport

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.post-processor.amazon-import"

// Configuration of this post processor
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`

	// Variables specific to this post processor
	S3Bucket  string            `mapstructure:"s3_bucket_name"`
	S3Key     string            `mapstructure:"s3_key_name"`
	SkipClean bool              `mapstructure:"skip_clean"`
	Tags      map[string]string `mapstructure:"tags"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

// Entry point for configuration parisng when we've defined
func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"s3_key_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Set defaults
	if p.config.S3Key == "" {
		p.config.S3Key = "packer-import-{{timestamp}}.ova"
	}

	errs := new(packer.MultiError)

	// Check and render s3_key_name
	if err = interpolate.Validate(p.config.S3Key, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing s3_key_name template: %s", err))
	}

	// Check we have AWS access variables defined somewhere
	errs = packer.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	// define all our required paramaters
	templates := map[string]*string{
		"s3_bucket_name": &p.config.S3Bucket,
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

	log.Println(common.ScrubConfig(p.config, p.config.AccessKey, p.config.SecretKey))
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error

	config, err := p.config.Config()
	if err != nil {
		return nil, false, err
	}

	// Render this key since we didn't in the configure phase
	p.config.S3Key, err = interpolate.Render(p.config.S3Key, &p.config.ctx)
	if err != nil {
		return nil, false, fmt.Errorf("Error rendering s3_key_name template: %s", err)
	}
	log.Printf("Rendered s3_key_name as %s", p.config.S3Key)

	log.Println("Looking for OVA in artifact")
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
		return nil, false, fmt.Errorf("No OVA file found in artifact from builder")
	}

	// Set up the AWS session
	log.Println("Creating AWS session")
	session := session.New(config)

	// open the source file
	log.Printf("Opening file %s to upload", source)
	file, err := os.Open(source)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to open %s: %s", source, err)
	}

	ui.Message(fmt.Sprintf("Uploading %s to s3://%s/%s", source, p.config.S3Bucket, p.config.S3Key))

	// Copy the OVA file into the S3 bucket specified
	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   file,
		Bucket: &p.config.S3Bucket,
		Key:    &p.config.S3Key,
	})
	if err != nil {
		return nil, false, fmt.Errorf("Failed to upload %s: %s", source, err)
	}

	// May as well stop holding this open now
	file.Close()

	ui.Message(fmt.Sprintf("Completed upload of %s to s3://%s/%s", source, p.config.S3Bucket, p.config.S3Key))

	// Call EC2 image import process
	log.Printf("Calling EC2 to import from s3://%s/%s", p.config.S3Bucket, p.config.S3Key)

	ec2conn := ec2.New(session)
	import_start, err := ec2conn.ImportImage(&ec2.ImportImageInput{
		DiskContainers: []*ec2.ImageDiskContainer{
			{
				UserBucket: &ec2.UserBucket{
					S3Bucket: &p.config.S3Bucket,
					S3Key:    &p.config.S3Key,
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
		Pending: []string{"pending", "active"},
		Refresh: awscommon.ImportImageRefreshFunc(ec2conn, *import_start.ImportTaskId),
		Target:  "completed",
	}

	// Actually do the wait for state change
	// We ignore errors out of this and check job state in AWS API
	awscommon.WaitForState(&stateChange)

	// Retrieve what the outcome was for the import task
	import_result, err := ec2conn.DescribeImportImageTasks(&ec2.DescribeImportImageTasksInput{
		ImportTaskIds: []*string{
			import_start.ImportTaskId,
		},
	})

	if err != nil {
		return nil, false, fmt.Errorf("Failed to find import task %s: %s", *import_start.ImportTaskId, err)
	}

	// Check it was actually completed
	if *import_result.ImportImageTasks[0].Status != "completed" {
		// The most useful error message is from the job itself
		return nil, false, fmt.Errorf("Import task %s failed: %s", *import_start.ImportTaskId, *import_result.ImportImageTasks[0].StatusMessage)
	}

	ui.Message(fmt.Sprintf("Import task %s complete", *import_start.ImportTaskId))

	// Pull AMI ID out of the completed job
	createdami := *import_result.ImportImageTasks[0].ImageId

	// If we have tags, then apply them now to both the AMI and snaps
	// created by the import
	if len(p.config.Tags) > 0 {
		var ec2Tags []*ec2.Tag

		log.Printf("Repacking tags into AWS format")

		for key, value := range p.config.Tags {
			ui.Message(fmt.Sprintf("Adding tag \"%s\": \"%s\"", key, value))
			ec2Tags = append(ec2Tags, &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		resourceIds := []*string{&createdami}

		log.Printf("Getting details of %s", createdami)

		imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: resourceIds,
		})

		if err != nil {
			return nil, false, fmt.Errorf("Failed to retrieve details for AMI %s: %s", createdami, err)
		}

		if len(imageResp.Images) == 0 {
			return nil, false, fmt.Errorf("AMI %s has no images", createdami)
		}

		image := imageResp.Images[0]

		log.Printf("Walking block device mappings for %s to find snapshots", createdami)

		for _, device := range image.BlockDeviceMappings {
			if device.Ebs != nil && device.Ebs.SnapshotId != nil {
				ui.Message(fmt.Sprintf("Tagging snapshot %s", *device.Ebs.SnapshotId))
				resourceIds = append(resourceIds, device.Ebs.SnapshotId)
			}
		}

		ui.Message(fmt.Sprintf("Tagging AMI %s", createdami))

		_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
			Resources: resourceIds,
			Tags:      ec2Tags,
		})

		if err != nil {
			return nil, false, fmt.Errorf("Failed to add tags to resources %#v: %s", resourceIds, err)
		}

	}

	// Add the reported AMI ID to the artifact list
	log.Printf("Adding created AMI ID %s in region %s to output artifacts", createdami, *config.Region)
	artifact = &awscommon.Artifact{
		Amis: map[string]string{
			*config.Region: createdami,
		},
		BuilderIdValue: BuilderId,
		Conn:           ec2conn,
	}

	if !p.config.SkipClean {
		ui.Message(fmt.Sprintf("Deleting import source s3://%s/%s", p.config.S3Bucket, p.config.S3Key))
		s3conn := s3.New(session)
		_, err = s3conn.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &p.config.S3Bucket,
			Key:    &p.config.S3Key,
		})
		if err != nil {
			return nil, false, fmt.Errorf("Failed to delete s3://%s/%s: %s", p.config.S3Bucket, p.config.S3Key, err)
		}
	}

	return artifact, false, nil
}
