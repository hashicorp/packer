// The instance package contains a packer.Builder implementation that builds
// AMIs for Amazon EC2 backed by instance storage, as opposed to EBS storage.
package instance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazon.instance"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.AMIConfig    `mapstructure:",squash"`
	awscommon.BlockDevices `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`

	AccountId           string `mapstructure:"account_id"`
	BundleDestination   string `mapstructure:"bundle_destination"`
	BundlePrefix        string `mapstructure:"bundle_prefix"`
	BundleUploadCommand string `mapstructure:"bundle_upload_command"`
	BundleVolCommand    string `mapstructure:"bundle_vol_command"`
	S3Bucket            string `mapstructure:"s3_bucket"`
	X509CertPath        string `mapstructure:"x509_cert_path"`
	X509KeyPath         string `mapstructure:"x509_key_path"`
	X509UploadPath      string `mapstructure:"x509_upload_path"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	configs := make([]interface{}, len(raws)+1)
	configs[0] = map[string]interface{}{
		"bundle_prefix": "image-{{timestamp}}",
	}
	copy(configs[1:], raws)

	b.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"ami_description",
				"bundle_upload_command",
				"bundle_vol_command",
				"run_tags",
				"run_volume_tags",
				"snapshot_tags",
				"tags",
				"spot_tags",
			},
		},
	}, configs...)
	if err != nil {
		return nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.AMIForceDeregister = true
	}

	if b.config.BundleDestination == "" {
		b.config.BundleDestination = "/tmp"
	}

	if b.config.BundleUploadCommand == "" {
		if b.config.IamInstanceProfile != "" {
			b.config.BundleUploadCommand = "sudo -i -n ec2-upload-bundle " +
				"-b {{.BucketName}} " +
				"-m {{.ManifestPath}} " +
				"-d {{.BundleDirectory}} " +
				"--batch " +
				"--region {{.Region}} " +
				"--retry"
		} else {
			b.config.BundleUploadCommand = "sudo -i -n ec2-upload-bundle " +
				"-b {{.BucketName}} " +
				"-m {{.ManifestPath}} " +
				"-a {{.AccessKey}} " +
				"-s {{.SecretKey}} " +
				"-d {{.BundleDirectory}} " +
				"--batch " +
				"--region {{.Region}} " +
				"--retry"
		}
	}

	if b.config.BundleVolCommand == "" {
		b.config.BundleVolCommand = "sudo -i -n ec2-bundle-vol " +
			"-k {{.KeyPath}} " +
			"-u {{.AccountId}} " +
			"-c {{.CertPath}} " +
			"-r {{.Architecture}} " +
			"-e {{.PrivatePath}}/* " +
			"-d {{.Destination}} " +
			"-p {{.Prefix}} " +
			"--batch " +
			"--no-filter"
	}

	if b.config.X509UploadPath == "" {
		b.config.X509UploadPath = "/tmp"
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.BlockDevices.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.AMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if b.config.AccountId == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("account_id is required"))
	} else {
		b.config.AccountId = strings.Replace(b.config.AccountId, "-", "", -1)
	}

	if b.config.S3Bucket == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("s3_bucket is required"))
	}

	if b.config.X509CertPath == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("x509_cert_path is required"))
	} else if _, err := os.Stat(b.config.X509CertPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("x509_cert_path points to bad file: %s", err))
	}

	if b.config.X509KeyPath == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("x509_key_path is required"))
	} else if _, err := os.Stat(b.config.X509KeyPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("x509_key_path points to bad file: %s", err))
	}

	if b.config.IsSpotInstance() && ((b.config.AMIENASupport != nil && *b.config.AMIENASupport) || b.config.AMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an AMI that already has either SR-IOV or ENA enabled."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}
	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}
	ec2conn := ec2.New(session)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("access_config", &b.config.AccessConfig)
	state.Put("ami_config", &b.config.AMIConfig)
	state.Put("ec2", ec2conn)
	state.Put("awsSession", session)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var instanceStep multistep.Step

	if b.config.IsSpotInstance() {
		instanceStep = &awscommon.StepRunSpotInstance{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			BlockDevices:             b.config.BlockDevices,
			BlockDurationMinutes:     b.config.BlockDurationMinutes,
			Ctx:                      b.config.ctx,
			Comm:                     &b.config.RunConfig.Comm,
			Debug:                    b.config.PackerDebug,
			EbsOptimized:             b.config.EbsOptimized,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			InstanceType:             b.config.InstanceType,
			SourceAMI:                b.config.SourceAmi,
			SpotPrice:                b.config.SpotPrice,
			SpotPriceProduct:         b.config.SpotPriceAutoProduct,
			Tags:                     b.config.RunTags,
			SpotTags:                 b.config.SpotTags,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
		}
	} else {
		instanceStep = &awscommon.StepRunSourceInstance{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			BlockDevices:             b.config.BlockDevices,
			Comm:                     &b.config.RunConfig.Comm,
			Ctx:                      b.config.ctx,
			Debug:                    b.config.PackerDebug,
			EbsOptimized:             b.config.EbsOptimized,
			EnableT2Unlimited:        b.config.EnableT2Unlimited,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			InstanceType:             b.config.InstanceType,
			IsRestricted:             b.config.IsChinaCloud() || b.config.IsGovCloud(),
			SourceAMI:                b.config.SourceAmi,
			Tags:                     b.config.RunTags,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
		}
	}

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:     b.config.AMIName,
			ForceDeregister: b.config.AMIForceDeregister,
		},
		&awscommon.StepSourceAMIInfo{
			SourceAmi:                b.config.SourceAmi,
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
			AmiFilters:               b.config.SourceAmiFilter,
			AMIVirtType:              b.config.AMIVirtType,
		},
		&awscommon.StepNetworkInfo{
			VpcId:               b.config.VpcId,
			VpcFilter:           b.config.VpcFilter,
			SecurityGroupIds:    b.config.SecurityGroupIds,
			SecurityGroupFilter: b.config.SecurityGroupFilter,
			SubnetId:            b.config.SubnetId,
			SubnetFilter:        b.config.SubnetFilter,
			AvailabilityZone:    b.config.AvailabilityZone,
		},
		&awscommon.StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.RunConfig.Comm,
			DebugKeyPath: fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
		},
		&awscommon.StepSecurityGroup{
			CommConfig:             &b.config.RunConfig.Comm,
			SecurityGroupFilter:    b.config.SecurityGroupFilter,
			SecurityGroupIds:       b.config.SecurityGroupIds,
			TemporarySGSourceCidrs: b.config.TemporarySGSourceCidrs,
		},
		instanceStep,
		&awscommon.StepGetPassword{
			Debug:     b.config.PackerDebug,
			Comm:      &b.config.RunConfig.Comm,
			Timeout:   b.config.WindowsPasswordTimeout,
			BuildName: b.config.PackerBuildName,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: awscommon.SSHHost(
				ec2conn,
				b.config.Comm.SSHInterface),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&StepUploadX509Cert{},
		&StepBundleVolume{
			Debug: b.config.PackerDebug,
		},
		&StepUploadBundle{
			Debug: b.config.PackerDebug,
		},
		&awscommon.StepDeregisterAMI{
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.AMIForceDeregister,
			ForceDeleteSnapshot: b.config.AMIForceDeleteSnapshot,
			AMIName:             b.config.AMIName,
			Regions:             b.config.AMIRegions,
		},
		&StepRegisterAMI{
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
		},
		&awscommon.StepAMIRegionCopy{
			AccessConfig:      &b.config.AccessConfig,
			Regions:           b.config.AMIRegions,
			AMIKmsKeyId:       b.config.AMIKmsKeyId,
			RegionKeyIds:      b.config.AMIRegionKMSKeyIDs,
			EncryptBootVolume: b.config.AMIEncryptBootVolume,
			Name:              b.config.AMIName,
			OriginalRegion:    *ec2conn.Config.Region,
		},
		&awscommon.StepModifyAMIAttributes{
			Description:    b.config.AMIDescription,
			Users:          b.config.AMIUsers,
			Groups:         b.config.AMIGroups,
			ProductCodes:   b.config.AMIProductCodes,
			SnapshotUsers:  b.config.SnapshotUsers,
			SnapshotGroups: b.config.SnapshotGroups,
			Ctx:            b.config.ctx,
		},
		&awscommon.StepCreateTags{
			Tags:         b.config.AMITags,
			SnapshotTags: b.config.SnapshotTags,
			Ctx:          b.config.ctx,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state.GetOk("amis"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state.Get("amis").(map[string]string),
		BuilderIdValue: BuilderId,
		Session:        session,
	}

	return artifact, nil
}
