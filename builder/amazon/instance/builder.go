// The instance package contains a packer.Builder implementation that builds
// AMIs for Amazon EC2 backed by instance storage, as opposed to EBS storage.
package instance

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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
				"bundle_upload_command",
				"bundle_vol_command",
			},
		},
	}, configs...)
	if err != nil {
		return nil, err
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
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(&b.config.ctx)...)
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

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	config, err := b.config.Config()
	if err != nil {
		return nil, err
	}

	session := session.New(config)
	ec2conn := ec2.New(session)

	// If the subnet is specified but not the AZ, try to determine the AZ automatically
	if b.config.SubnetId != "" && b.config.AvailabilityZone == "" {
		log.Printf("[INFO] Finding AZ for the given subnet '%s'", b.config.SubnetId)
		resp, err := ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: []*string{&b.config.SubnetId}})
		if err != nil {
			return nil, err
		}
		b.config.AvailabilityZone = *resp.Subnets[0].AvailabilityZone
		log.Printf("[INFO] AZ found: '%s'", b.config.AvailabilityZone)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:     b.config.AMIName,
			ForceDeregister: b.config.AMIForceDeregister,
		},
		&awscommon.StepSourceAMIInfo{
			SourceAmi:          b.config.SourceAmi,
			EnhancedNetworking: b.config.AMIEnhancedNetworking,
		},
		&awscommon.StepKeyPair{
			Debug:                b.config.PackerDebug,
			DebugKeyPath:         fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
			KeyPairName:          b.config.SSHKeyPairName,
			PrivateKeyFile:       b.config.RunConfig.Comm.SSHPrivateKey,
			TemporaryKeyPairName: b.config.TemporaryKeyPairName,
		},
		&awscommon.StepSecurityGroup{
			CommConfig:       &b.config.RunConfig.Comm,
			SecurityGroupIds: b.config.SecurityGroupIds,
			VpcId:            b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			Debug:                    b.config.PackerDebug,
			SpotPrice:                b.config.SpotPrice,
			SpotPriceProduct:         b.config.SpotPriceAutoProduct,
			InstanceType:             b.config.InstanceType,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			SourceAMI:                b.config.SourceAmi,
			SubnetId:                 b.config.SubnetId,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			EbsOptimized:             b.config.EbsOptimized,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             b.config.BlockDevices,
			Tags:                     b.config.RunTags,
		},
		&awscommon.StepGetPassword{
			Debug:   b.config.PackerDebug,
			Comm:    &b.config.RunConfig.Comm,
			Timeout: b.config.WindowsPasswordTimeout,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: awscommon.SSHHost(
				ec2conn,
				b.config.SSHPrivateIp),
			SSHConfig: awscommon.SSHConfig(
				b.config.RunConfig.Comm.SSHUsername),
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
			ForceDeregister: b.config.AMIForceDeregister,
			AMIName:         b.config.AMIName,
		},
		&StepRegisterAMI{},
		&awscommon.StepAMIRegionCopy{
			AccessConfig: &b.config.AccessConfig,
			Regions:      b.config.AMIRegions,
			Name:         b.config.AMIName,
		},
		&awscommon.StepModifyAMIAttributes{
			Description:  b.config.AMIDescription,
			Users:        b.config.AMIUsers,
			Groups:       b.config.AMIGroups,
			ProductCodes: b.config.AMIProductCodes,
		},
		&awscommon.StepCreateTags{
			Tags: b.config.AMITags,
		},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

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
		Conn:           ec2conn,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
