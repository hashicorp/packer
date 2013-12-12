// The googlecompute package contains a packer.Builder implementation that
// builds images for Google Compute Engine.
package googlecompute

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// The unique ID for this builder.
const BuilderId = "packer.googlecompute"

// Builder represents a Packer Builder.
type Builder struct {
	config config
	runner multistep.Runner
}

// config holds the googlecompute builder configuration settings.
type config struct {
	BucketName          string            `mapstructure:"bucket_name"`
	ClientSecretsFile   string            `mapstructure:"client_secrets_file"`
	ImageName           string            `mapstructure:"image_name"`
	ImageDescription    string            `mapstructure:"image_description"`
	MachineType         string            `mapstructure:"machine_type"`
	Metadata            map[string]string `mapstructure:"metadata"`
	Network             string            `mapstructure:"network"`
	Passphrase          string            `mapstructure:"passphrase"`
	PrivateKeyFile      string            `mapstructure:"private_key_file"`
	ProjectId           string            `mapstructure:"project_id"`
	SourceImage         string            `mapstructure:"source_image"`
	SSHUsername         string            `mapstructure:"ssh_username"`
	SSHPort             uint              `mapstructure:"ssh_port"`
	RawSSHTimeout       string            `mapstructure:"ssh_timeout"`
	RawStateTimeout     string            `mapstructure:"state_timeout"`
	Tags                []string          `mapstructure:"tags"`
	Zone                string            `mapstructure:"zone"`
	clientSecrets       *clientSecrets
	common.PackerConfig `mapstructure:",squash"`
	instanceName        string
	privateKeyBytes     []byte
	sshTimeout          time.Duration
	stateTimeout        time.Duration
	tpl                 *packer.ConfigTemplate
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	// Load the packer config.
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}
	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	errs := common.CheckUnusedConfig(md)
	// Collect errors if any.
	if err := common.CheckUnusedConfig(md); err != nil {
		return nil, err
	}
	// Set defaults.
	if b.config.Network == "" {
		b.config.Network = "default"
	}
	if b.config.ImageDescription == "" {
		b.config.ImageDescription = "Created by Packer"
	}
	if b.config.ImageName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.ImageName = "packer-{{timestamp}}"
	}
	if b.config.MachineType == "" {
		b.config.MachineType = "n1-standard-1"
	}
	if b.config.RawSSHTimeout == "" {
		b.config.RawSSHTimeout = "5m"
	}
	if b.config.RawStateTimeout == "" {
		b.config.RawStateTimeout = "5m"
	}
	if b.config.SSHUsername == "" {
		b.config.SSHUsername = "root"
	}
	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}
	// Process Templates
	templates := map[string]*string{
		"bucket_name":         &b.config.BucketName,
		"client_secrets_file": &b.config.ClientSecretsFile,
		"image_name":          &b.config.ImageName,
		"image_description":   &b.config.ImageDescription,
		"machine_type":        &b.config.MachineType,
		"network":             &b.config.Network,
		"passphrase":          &b.config.Passphrase,
		"private_key_file":    &b.config.PrivateKeyFile,
		"project_id":          &b.config.ProjectId,
		"source_image":        &b.config.SourceImage,
		"ssh_username":        &b.config.SSHUsername,
		"ssh_timeout":         &b.config.RawSSHTimeout,
		"state_timeout":       &b.config.RawStateTimeout,
		"zone":                &b.config.Zone,
	}
	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}
	// Process required parameters.
	if b.config.BucketName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a bucket_name must be specified"))
	}
	if b.config.ClientSecretsFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a client_secrets_file must be specified"))
	}
	if b.config.PrivateKeyFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a private_key_file must be specified"))
	}
	if b.config.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}
	if b.config.SourceImage == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image must be specified"))
	}
	if b.config.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}
	// Process timeout settings.
	sshTimeout, err := time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	b.config.sshTimeout = sshTimeout
	stateTimeout, err := time.ParseDuration(b.config.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	b.config.stateTimeout = stateTimeout
	// Load the client secrets file.
	cs, err := loadClientSecrets(b.config.ClientSecretsFile)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing client secrets file: %s", err))
	}
	b.config.clientSecrets = cs
	// Load the private key.
	b.config.privateKeyBytes, err = processPrivateKeyFile(b.config.PrivateKeyFile, b.config.Passphrase)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed loading private key file: %s", err))
	}
	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}
	return nil, nil
}

// Run executes a googlecompute Packer build and returns a packer.Artifact
// representing a GCE machine image.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Initialize the Google Compute Engine API.
	client, err := New(b.config.ProjectId, b.config.Zone, b.config.clientSecrets, b.config.privateKeyBytes)
	if err != nil {
		log.Println("Failed to create the Google Compute Engine client.")
		return nil, err
	}
	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)
	// Build the steps.
	steps := []multistep.Step{
		new(stepCreateSSHKey),
		new(stepCreateInstance),
		new(stepInstanceInfo),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: 5 * time.Minute,
		},
		new(common.StepProvision),
		new(stepUpdateGsutil),
		new(stepCreateImage),
		new(stepUploadImage),
		new(stepRegisterImage),
	}
	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)
	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	if _, ok := state.GetOk("image_name"); !ok {
		log.Println("Failed to find image_name in state. Bug?")
		return nil, nil
	}
	artifact := &Artifact{
		imageName: state.Get("image_name").(string),
		client:    client,
	}
	return artifact, nil
}

// Cancel.
func (b *Builder) Cancel() {}
