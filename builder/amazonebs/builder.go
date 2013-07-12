// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"text/template"
	"time"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazonebs"

type config struct {
	// Access information
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	// Information for the source instance
	Region          string
	SourceAmi       string `mapstructure:"source_ami"`
	InstanceType    string `mapstructure:"instance_type"`
	SSHUsername     string `mapstructure:"ssh_username"`
	SSHPort         int    `mapstructure:"ssh_port"`
	SSHTimeout      time.Duration
	SecurityGroupId string `mapstructure:"security_group_id"`

	// Configuration of the resulting AMI
	AMIName string `mapstructure:"ami_name"`

	PackerDebug   bool   `mapstructure:"packer_debug"`
	RawSSHTimeout string `mapstructure:"ssh_timeout"`
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	var err error

	for _, raw := range raws {
		err := mapstructure.Decode(raw, &b.config)
		if err != nil {
			return err
		}
	}

	if b.config.AccessKey == "" {
		b.config.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}

	if b.config.AccessKey == "" {
		b.config.AccessKey = os.Getenv("AWS_ACCESS_KEY")
	}

	if b.config.SecretKey == "" {
		b.config.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	if b.config.SecretKey == "" {
		b.config.SecretKey = os.Getenv("AWS_SECRET_KEY")
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.RawSSHTimeout == "" {
		b.config.RawSSHTimeout = "1m"
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

	if b.config.Region == "" {
		errs = append(errs, errors.New("A region must be specified"))
	} else if _, ok := aws.Regions[b.config.Region]; !ok {
		errs = append(errs, fmt.Errorf("Unknown region: %s", b.config.Region))
	}

	if b.config.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	b.config.SSHTimeout, err = time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	if b.config.AMIName == "" {
		errs = append(errs, errors.New("ami_name must be specified"))
	} else {
		_, err = template.New("ami").Parse(b.config.AMIName)
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed parsing ami_name: %s", err))
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	log.Printf("Config: %+v", b.config)
	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
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
		&stepSecurityGroup{},
		&stepRunSourceInstance{},
		&stepConnectSSH{},
		&stepProvision{},
		&stepStopInstance{},
		&stepCreateAMI{},
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
	artifact := &artifact{
		amis: state["amis"].(map[string]string),
		conn: ec2conn,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
