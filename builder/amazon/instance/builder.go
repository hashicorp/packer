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
	"text/template"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazon.instance"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`

	AccountId           string `mapstructure:"account_id"`
	AMIName             string `mapstructure:"ami_name"`
	BundleDestination   string `mapstructure:"bundle_destination"`
	BundlePrefix        string `mapstructure:"bundle_prefix"`
	BundleUploadCommand string `mapstructure:"bundle_upload_command"`
	BundleVolCommand    string `mapstructure:"bundle_vol_command"`
	S3Bucket            string `mapstructure:"s3_bucket"`
	Tags                map[string]string
	X509CertPath        string `mapstructure:"x509_cert_path"`
	X509KeyPath         string `mapstructure:"x509_key_path"`
	X509UploadPath      string `mapstructure:"x509_upload_path"`
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	if b.config.BundleDestination == "" {
		b.config.BundleDestination = "/tmp"
	}

	if b.config.BundlePrefix == "" {
		b.config.BundlePrefix = "image-{{.CreateTime}}"
	}

	if b.config.BundleUploadCommand == "" {
		b.config.BundleUploadCommand = "sudo -n ec2-upload-bundle " +
			"-b {{.BucketName}} " +
			"-m {{.ManifestPath}} " +
			"-a {{.AccessKey}} " +
			"-s {{.SecretKey}} " +
			"-d {{.BundleDirectory}} " +
			"--batch " +
			"--retry"
	}

	if b.config.BundleVolCommand == "" {
		b.config.BundleVolCommand = "sudo -n ec2-bundle-vol " +
			"-k {{.KeyPath}} " +
			"-u {{.AccountId}} " +
			"-c {{.CertPath}} " +
			"-r {{.Architecture}} " +
			"-e {{.PrivatePath}} " +
			"-d {{.Destination}} " +
			"-p {{.Prefix}} " +
			"--batch"
	}

	if b.config.X509UploadPath == "" {
		b.config.X509UploadPath = "/tmp"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare()...)

	if b.config.AccountId == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("account_id is required"))
	} else {
		b.config.AccountId = strings.Replace(b.config.AccountId, "-", "", -1)
	}

	if b.config.AMIName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ami_name must be specified"))
	} else {
		_, err = template.New("ami").Parse(b.config.AMIName)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed parsing ami_name: %s", err))
		}
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
		return errs
	}

	log.Printf("Config: %+v", b.config)
	return nil
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
	state := make(map[string]interface{})
	state["config"] = &b.config
	state["ec2"] = ec2conn
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepKeyPair{},
		&awscommon.StepSecurityGroup{
			SecurityGroupId: b.config.SecurityGroupId,
			SSHPort:         b.config.SSHPort,
			VpcId:           b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			ExpectedRootDevice: "instance-store",
			InstanceType:       b.config.InstanceType,
			UserData:           b.config.UserData,
			SourceAMI:          b.config.SourceAmi,
			SubnetId:           b.config.SubnetId,
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
		&awscommon.StepCreateTags{b.config.Tags},
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
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state["amis"]; !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state["amis"].(map[string]string),
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
