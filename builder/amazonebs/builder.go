// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/mapstructure"
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
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	err = mapstructure.Decode(raw, &b.config)
	if err != nil {
		return
	}

	log.Printf("Config: %+v", b.config)

	// TODO: config validation and asking for fields:
	// * access key
	// * secret key
	// * region
	// * source ami
	// * instance type
	// * SSH username
	// * SSH port? (or default to 22?)

	return
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook) packer.Artifact {
	auth := aws.Auth{b.config.AccessKey, b.config.SecretKey}
	region := aws.Regions[b.config.Region]
	ec2conn := ec2.New(auth, region)

	// Setup the state bag and initial state for the steps
	state := make(map[string]interface{})
	state["config"] = b.config
	state["ec2"] = ec2conn
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []Step{
		&stepKeyPair{},
		&stepRunSourceInstance{},
		&stepStopInstance{},
		&stepCreateAMI{},
	}

	// Run!
	RunSteps(state, steps)

	// Build the artifact and return it
	return &artifact{state["amis"].(map[string]string)}
}
