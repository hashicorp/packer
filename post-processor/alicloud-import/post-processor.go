package alicloudimport

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	packercommon "github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/ram"
	packerecs "github.com/hashicorp/packer/builder/alicloud/ecs"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	BuilderId                             = "packer.post-processor.alicloud-import"
	OSSSuffix                             = "oss-"
	RAWFileFormat                         = "raw"
	VHDFileFormat                         = "vhd"
	BUSINESSINFO                          = "packer"
	AliyunECSImageImportDefaultRolePolicy = `{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs.aliyuncs.com"
        ]
      }
    }
  ],
  "Version": "1"
}`
)

// Configuration of this post processor
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	packerecs.Config    `mapstructure:",squash"`

	// Variables specific to this post processor
	OSSBucket                       string            `mapstructure:"oss_bucket_name"`
	OSSKey                          string            `mapstructure:"oss_key_name"`
	SkipClean                       bool              `mapstructure:"skip_clean"`
	Tags                            map[string]string `mapstructure:"tags"`
	AlicloudImageName               string            `mapstructure:"image_name"`
	AlicloudImageVersion            string            `mapstructure:"image_version"`
	AlicloudImageDescription        string            `mapstructure:"image_description"`
	AlicloudImageShareAccounts      []string          `mapstructure:"image_share_account"`
	AlicloudImageDestinationRegions []string          `mapstructure:"image_copy_regions"`
	OSType                          string            `mapstructure:"image_os_type"`
	Platform                        string            `mapstructure:"image_platform"`
	Architecture                    string            `mapstructure:"image_architecture"`
	Size                            string            `mapstructure:"image_system_size"`
	Format                          string            `mapstructure:"format"`
	AlicloudImageForceDetele        bool              `mapstructure:"image_force_delete"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config            Config
	DiskDeviceMapping []ecs.DiskDeviceMapping
}

// Entry point for configuration parsing when we've defined
func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"oss_key_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	// Check and render oss_key_name
	if err = interpolate.Validate(p.config.OSSKey, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing oss_key_name template: %s", err))
	}

	// Check we have alicloud access variables defined somewhere
	errs = packer.MultiErrorAppend(errs, p.config.AlicloudAccessConfig.Prepare(&p.config.ctx)...)

	// define all our required parameters
	templates := map[string]*string{
		"oss_bucket_name": &p.config.OSSBucket,
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

	log.Println(common.ScrubConfig(p.config, p.config.AlicloudAccessKey, p.config.AlicloudSecretKey))
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error

	// Render this key since we didn't in the configure phase
	p.config.OSSKey, err = interpolate.Render(p.config.OSSKey, &p.config.ctx)
	if err != nil {
		return nil, false, fmt.Errorf("Error rendering oss_key_name template: %s", err)
	}
	if p.config.OSSKey == "" {
		p.config.OSSKey = "Packer_" + strconv.Itoa(time.Now().Nanosecond())
	}
	log.Printf("Rendered oss_key_name as %s", p.config.OSSKey)

	log.Println("Looking for RAW or VHD in artifact")
	// Locate the files output from the builder
	source := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, VHDFileFormat) || strings.HasSuffix(path, RAWFileFormat) {
			source = path
			break
		}
	}

	// Hope we found something useful
	if source == "" {
		return nil, false, fmt.Errorf("No vhd or raw file found in artifact from builder")
	}

	ecsClient, err := p.config.AlicloudAccessConfig.Client()
	if err != nil {
		return nil, false, fmt.Errorf("Failed to connect alicloud ecs  %s", err)
	}
	ecsClient.SetBusinessInfo(BUSINESSINFO)

	images, _, err := ecsClient.DescribeImages(&ecs.DescribeImagesArgs{
		RegionId:  packercommon.Region(p.config.AlicloudRegion),
		ImageName: p.config.AlicloudImageName,
	})
	if err != nil {
		return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
			getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
	}

	if len(images) > 0 && !p.config.AlicloudImageForceDetele {
		return nil, false, fmt.Errorf("Duplicated image exists, please delete the existing images " +
			"or set the 'image_force_delete' value as true")
	}

	// Set up the OSS client
	log.Println("Creating OSS Client")
	client, err := oss.New(getEndPonit(p.config.AlicloudRegion), p.config.AlicloudAccessKey,
		p.config.AlicloudSecretKey)
	if err != nil {
		return nil, false, fmt.Errorf("Creating oss connection failed: %s", err)
	}
	bucket, err := queryOrCreateBucket(p.config.OSSBucket, client)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to query or create bucket %s: %s", p.config.OSSBucket, err)
	}

	if err != nil {
		return nil, false, fmt.Errorf("Failed to open %s: %s", source, err)
	}

	err = bucket.PutObjectFromFile(p.config.OSSKey, source)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to upload image %s: %s", source, err)
	}
	if len(images) > 0 && p.config.AlicloudImageForceDetele {
		if err = ecsClient.DeleteImage(packercommon.Region(p.config.AlicloudRegion),
			images[0].ImageId); err != nil {
			return nil, false, fmt.Errorf("Delete duplicated image %s failed", images[0].ImageName)
		}
	}

	diskDeviceMapping := ecs.DiskDeviceMapping{
		Size:      p.config.Size,
		Format:    p.config.Format,
		OSSBucket: p.config.OSSBucket,
		OSSObject: p.config.OSSKey,
	}
	imageImageArgs := &ecs.ImportImageArgs{
		RegionId:     packercommon.Region(p.config.AlicloudRegion),
		ImageName:    p.config.AlicloudImageName,
		ImageVersion: p.config.AlicloudImageVersion,
		Description:  p.config.AlicloudImageDescription,
		Architecture: p.config.Architecture,
		OSType:       p.config.OSType,
		Platform:     p.config.Platform,
	}
	imageImageArgs.DiskDeviceMappings.DiskDeviceMapping = []ecs.DiskDeviceMapping{
		diskDeviceMapping,
	}
	imageId, err := ecsClient.ImportImage(imageImageArgs)

	if err != nil {
		e, _ := err.(*packercommon.Error)
		if e.Code == "NoSetRoletoECSServiceAcount" {
			ramClient := ram.NewClient(p.config.AlicloudAccessKey, p.config.AlicloudSecretKey)
			roleResponse, err := ramClient.GetRole(ram.RoleQueryRequest{
				RoleName: "AliyunECSImageImportDefaultRole",
			})
			if err != nil {
				return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
					getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
			}
			if roleResponse.Role.RoleId == "" {
				if _, err = ramClient.CreateRole(ram.RoleRequest{
					RoleName:                 "AliyunECSImageImportDefaultRole",
					AssumeRolePolicyDocument: AliyunECSImageImportDefaultRolePolicy,
				}); err != nil {
					return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
						getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
				}
				if _, err := ramClient.AttachPolicyToRole(ram.AttachPolicyToRoleRequest{
					PolicyRequest: ram.PolicyRequest{
						PolicyName: "AliyunECSImageImportRolePolicy",
						PolicyType: "System",
					},
					RoleName: "AliyunECSImageImportDefaultRole",
				}); err != nil {
					return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
						getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
				}
			} else {
				policyListResponse, err := ramClient.ListPoliciesForRole(ram.RoleQueryRequest{
					RoleName: "AliyunECSImageImportDefaultRole",
				})
				if err != nil {
					return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
						getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
				}
				isAliyunECSImageImportRolePolicyNotExit := true
				for _, policy := range policyListResponse.Policies.Policy {
					if policy.PolicyName == "AliyunECSImageImportRolePolicy" &&
						policy.PolicyType == "System" {
						isAliyunECSImageImportRolePolicyNotExit = false
						break
					}
				}
				if isAliyunECSImageImportRolePolicyNotExit {
					if _, err := ramClient.AttachPolicyToRole(ram.AttachPolicyToRoleRequest{
						PolicyRequest: ram.PolicyRequest{
							PolicyName: "AliyunECSImageImportRolePolicy",
							PolicyType: "System",
						},
						RoleName: "AliyunECSImageImportDefaultRole",
					}); err != nil {
						return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
							getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
					}
				}
				if _, err = ramClient.UpdateRole(
					ram.UpdateRoleRequest{
						RoleName:                    "AliyunECSImageImportDefaultRole",
						NewAssumeRolePolicyDocument: AliyunECSImageImportDefaultRolePolicy,
					}); err != nil {
					return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
						getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
				}
			}
			for i := 10; i > 0; i = i - 1 {
				imageId, err = ecsClient.ImportImage(imageImageArgs)
				if err != nil {
					e, _ = err.(*packercommon.Error)
					if e.Code == "NoSetRoletoECSServiceAcount" {
						time.Sleep(5 * time.Second)
						continue
					} else if e.Code == "ImageIsImporting" ||
						e.Code == "InvalidImageName.Duplicated" {
						break
					}
					return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
						getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
				}
				break
			}

		} else {

			return nil, false, fmt.Errorf("Failed to start import from %s/%s: %s",
				getEndPonit(p.config.OSSBucket), p.config.OSSKey, err)
		}
	}

	err = ecsClient.WaitForImageReady(packercommon.Region(p.config.AlicloudRegion),
		imageId, packerecs.ALICLOUD_DEFAULT_LONG_TIMEOUT)
	// Add the reported Alicloud image ID to the artifact list
	log.Printf("Importing created alicloud image ID %s in region %s Finished.", imageId, p.config.AlicloudRegion)
	artifact = &packerecs.Artifact{
		AlicloudImages: map[string]string{
			p.config.AlicloudRegion: imageId,
		},
		BuilderIdValue: BuilderId,
		Client:         ecsClient,
	}

	if !p.config.SkipClean {
		ui.Message(fmt.Sprintf("Deleting import source %s/%s/%s",
			getEndPonit(p.config.AlicloudRegion), p.config.OSSBucket, p.config.OSSKey))
		if err = bucket.DeleteObject(p.config.OSSKey); err != nil {
			return nil, false, fmt.Errorf("Failed to delete %s/%s/%s: %s",
				getEndPonit(p.config.AlicloudRegion), p.config.OSSBucket, p.config.OSSKey, err)
		}
	}

	return artifact, false, nil
}

func queryOrCreateBucket(bucketName string, client *oss.Client) (*oss.Bucket, error) {
	isExist, err := client.IsBucketExist(bucketName)
	if err != nil {
		return nil, err
	}
	if !isExist {
		err = client.CreateBucket(bucketName)
		if err != nil {
			return nil, err
		}
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return bucket, nil

}

func getEndPonit(region string) string {
	return "https://" + GetOSSRegion(region) + ".aliyuncs.com"
}

func GetOSSRegion(region string) string {
	if strings.HasPrefix(region, OSSSuffix) {
		return region
	}
	return OSSSuffix + region
}

func GetECSRegion(region string) string {
	if strings.HasPrefix(region, OSSSuffix) {
		return strings.TrimSuffix(region, OSSSuffix)
	}
	return region

}
