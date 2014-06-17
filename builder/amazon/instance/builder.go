// The instance package contains a packer.Builder implementation that builds
// AMIs for Amazon EC2 backed by instance storage, as opposed to EBS storage.
package instance

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"strings"
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

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars
	b.config.tpl.Funcs(awscommon.TemplateFuncs)

	if b.config.BundleDestination == "" {
		b.config.BundleDestination = "/tmp"
	}

	if b.config.BundlePrefix == "" {
		b.config.BundlePrefix = "image-{{timestamp}}"
	}

	if b.config.BundleUploadCommand == "" {
		b.config.BundleUploadCommand = "sudo -n ec2-upload-bundle " +
			"-b {{.BucketName}} " +
			"-m {{.ManifestPath}} " +
			"-a {{.AccessKey}} " +
			"-s {{.SecretKey}} " +
			"-d {{.BundleDirectory}} " +
			"--batch " +
			"--url {{.S3Endpoint}} " +
			"--retry"
	}

	if b.config.BundleVolCommand == "" {
		b.config.BundleVolCommand = "sudo -n ec2-bundle-vol " +
			"-k {{.KeyPath}} " +
			"-u {{.AccountId}} " +
			"-c {{.CertPath}} " +
			"-r {{.Architecture}} " +
			"-e {{.PrivatePath}}/* " +
			"-d {{.Destination}} " +
			"-p {{.Prefix}} " +
			"--batch"
	}

	if b.config.X509UploadPath == "" {
		b.config.X509UploadPath = "/tmp"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)

	validates := map[string]*string{
		"bundle_upload_command": &b.config.BundleUploadCommand,
		"bundle_vol_command":    &b.config.BundleVolCommand,
	}

	for n, ptr := range validates {
		if err := b.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
		}
	}

	templates := map[string]*string{
		"account_id":         &b.config.AccountId,
		"ami_name":           &b.config.AMIName,
		"bundle_destination": &b.config.BundleDestination,
		"bundle_prefix":      &b.config.BundlePrefix,
		"s3_bucket":          &b.config.S3Bucket,
		"x509_cert_path":     &b.config.X509CertPath,
		"x509_key_path":      &b.config.X509KeyPath,
		"x509_upload_path":   &b.config.X509UploadPath,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

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
	region, err := b.config.Region()
	if err != nil {
		return nil, err
	}

	auth, err := b.config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}

	ec2conn := ec2.New(auth, region)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepKeyPair{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
			KeyPairName:    b.config.TemporaryKeyPairName,
			PrivateKeyFile: b.config.SSHPrivateKeyFile,
		},
		&awscommon.StepSecurityGroup{
			SecurityGroupIds: b.config.SecurityGroupIds,
			SSHPort:          b.config.SSHPort,
			VpcId:            b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			Debug:                    b.config.PackerDebug,
			ExpectedRootDevice:       "instance-store",
			InstanceType:             b.config.InstanceType,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			SourceAMI:                b.config.SourceAmi,
			SubnetId:                 b.config.SubnetId,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             b.config.BlockDevices,
			Tags:                     b.config.RunTags,
		},
		&common.StepConnectSSH{
			SSHAddress:     awscommon.SSHAddress(ec2conn, b.config.SSHPort),
			SSHConfig:      awscommon.SSHConfig(b.config.SSHUsername),
			SSHWaitTimeout: b.config.SSHTimeout(),
		},
		&common.StepProvision{},
		&StepUploadX509Cert{},
		&StepBundleVolume{},
		&StepUploadBundle{},
		&StepRegisterAMI{},
		&awscommon.StepAMIRegionCopy{
			Regions: b.config.AMIRegions,
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
