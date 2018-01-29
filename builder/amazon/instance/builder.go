// The instance package contains a packer.Builder implementation that builds
// AMIs for Amazon EC2 backed by instance storage, as opposed to EBS storage.
package instance

import (
	"errors"
	"fmt"
	"log"
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

	if b.config.IsSpotInstance() && (b.config.AMIENASupport || b.config.AMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an AMI that already has either SR-IOV or ENA enabled."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey, b.config.Token))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}
	ec2conn := ec2.New(session)

	// If the subnet is specified but not the VpcId or AZ, try to determine them automatically
	if b.config.SubnetId != "" && (b.config.AvailabilityZone == "" || b.config.VpcId == "") {
		log.Printf("[INFO] Finding AZ and VpcId for the given subnet '%s'", b.config.SubnetId)
		resp, err := ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: []*string{&b.config.SubnetId}})
		if err != nil {
			return nil, err
		}
		if b.config.AvailabilityZone == "" {
			b.config.AvailabilityZone = *resp.Subnets[0].AvailabilityZone
			log.Printf("[INFO] AvailabilityZone found: '%s'", b.config.AvailabilityZone)
		}
		if b.config.VpcId == "" {
			b.config.VpcId = *resp.Subnets[0].VpcId
			log.Printf("[INFO] VpcId found: '%s'", b.config.VpcId)
		}
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("ec2", ec2conn)
	state.Put("awsSession", session)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var instanceStep multistep.Step

	if b.config.IsSpotInstance() {
		instanceStep = &awscommon.StepRunSpotInstance{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             b.config.BlockDevices,
			Ctx:                      b.config.ctx,
			Debug:                    b.config.PackerDebug,
			EbsOptimized:             b.config.EbsOptimized,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			InstanceType:             b.config.InstanceType,
			SourceAMI:                b.config.SourceAmi,
			SpotPrice:                b.config.SpotPrice,
			SpotPriceProduct:         b.config.SpotPriceAutoProduct,
			SubnetId:                 b.config.SubnetId,
			Tags:                     b.config.RunTags,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
		}
	} else {
		instanceStep = &awscommon.StepRunSourceInstance{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             b.config.BlockDevices,
			Ctx:                      b.config.ctx,
			Debug:                    b.config.PackerDebug,
			EbsOptimized:             b.config.EbsOptimized,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			InstanceType:             b.config.InstanceType,
			SourceAMI:                b.config.SourceAmi,
			SubnetId:                 b.config.SubnetId,
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
		},
		&awscommon.StepKeyPair{
			Debug:                b.config.PackerDebug,
			SSHAgentAuth:         b.config.Comm.SSHAgentAuth,
			DebugKeyPath:         fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
			KeyPairName:          b.config.SSHKeyPairName,
			PrivateKeyFile:       b.config.RunConfig.Comm.SSHPrivateKey,
			TemporaryKeyPairName: b.config.TemporaryKeyPairName,
		},
		&awscommon.StepSecurityGroup{
			CommConfig:       &b.config.RunConfig.Comm,
			SecurityGroupIds: b.config.SecurityGroupIds,
			VpcId:            b.config.VpcId,
			TemporarySGSourceCidr: b.config.TemporarySGSourceCidr,
		},
		instanceStep,
		&awscommon.StepGetPassword{
			Debug:   b.config.PackerDebug,
			Comm:    &b.config.RunConfig.Comm,
			Timeout: b.config.WindowsPasswordTimeout,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: awscommon.SSHHost(
				ec2conn,
				b.config.SSHInterface),
			SSHConfig: awscommon.SSHConfig(
				b.config.RunConfig.Comm.SSHAgentAuth,
				b.config.RunConfig.Comm.SSHUsername,
				b.config.RunConfig.Comm.SSHPassword),
		},
		&common.StepProvision{},
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
			RegionKeyIds:      b.config.AMIRegionKMSKeyIDs,
			EncryptBootVolume: b.config.AMIEncryptBootVolume,
			Name:              b.config.AMIName,
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
	b.runner.Run(state)

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

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
