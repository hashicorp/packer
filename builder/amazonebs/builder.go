// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"errors"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazonebs"

type config struct {
	// Access information
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	// Information for the source instance
	Region       string
	SourceAmi    string `mapstructure:"source_ami"`
	InstanceType string `mapstructure:"instance_type"`
	SSHUsername  string `mapstructure:"ssh_username"`
	SSHPort      int    `mapstructure:"ssh_port"`

	// Configuration of the resulting AMI
	AMIName string `mapstructure:"ami_name"`
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	err = mapstructure.Decode(raw, &b.config)
	if err != nil {
		return
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	// Accumulate any errors
	errs := make([]error, 0)

	if b.config.AccessKey == "" {
		errs = append(errs, errors.New("An access_key must be specified"))
	}

	if b.config.SecretKey == "" {
		errs = append(errs, errors.New("A secret_key must be specified"))
	}

	if b.config.SourceAmi == "" {
		errs = append(errs, errors.New("A source_ami must be specified"))
	}

	if b.config.InstanceType == "" {
		errs = append(errs, errors.New("An instance_type must be specified"))
	}

	if b.config.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	// TODO: config validation and asking for fields:
	// * region (exists and valid)

	log.Printf("Config: %+v", b.config)
	return
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook) packer.Artifact {
	// Basic sanity checks. These are panics now because the Prepare
	// method should verify these exist and such.
	if b.config.AccessKey == "" {
		panic("access key not filled in")
	}

	if b.config.SecretKey == "" {
		panic("secret key not filled in")
	}

	if b.config.Region == "" {
		panic("region not filled in")
	}

	region, ok := aws.Regions[b.config.Region]
	if !ok {
		panic("region not found")
	}

	auth := aws.Auth{b.config.AccessKey, b.config.SecretKey}
	ec2conn := ec2.New(auth, region)

	// Setup the state bag and initial state for the steps
	state := make(map[string]interface{})
	state["config"] = b.config
	state["ec2"] = ec2conn
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&stepKeyPair{},
		&stepRunSourceInstance{},
		&stepConnectSSH{},
		&stepProvision{},
		&stepStopInstance{},
		&stepCreateAMI{},
	}

	// Run!
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(state)

	// If there are no AMIs, then jsut return
	if _, ok := state["amis"]; !ok {
		return nil
	}

	// Build the artifact and return it
	return &artifact{state["amis"].(map[string]string)}
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
