package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
)

// Config is the configuration structure for the GCE builder. It stores
// both the publicly settable state as well as the privately generated
// state of the config object.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	BucketName        string            `mapstructure:"bucket_name"`
	ClientSecretsFile string            `mapstructure:"client_secrets_file"`
	ImageName         string            `mapstructure:"image_name"`
	ImageDescription  string            `mapstructure:"image_description"`
	InstanceName      string            `mapstructure:"instance_name"`
	MachineType       string            `mapstructure:"machine_type"`
	Metadata          map[string]string `mapstructure:"metadata"`
	Network           string            `mapstructure:"network"`
	Passphrase        string            `mapstructure:"passphrase"`
	PrivateKeyFile    string            `mapstructure:"private_key_file"`
	ProjectId         string            `mapstructure:"project_id"`
	SourceImage       string            `mapstructure:"source_image"`
	SSHUsername       string            `mapstructure:"ssh_username"`
	SSHPort           uint              `mapstructure:"ssh_port"`
	RawSSHTimeout     string            `mapstructure:"ssh_timeout"`
	RawStateTimeout   string            `mapstructure:"state_timeout"`
	Tags              []string          `mapstructure:"tags"`
	Zone              string            `mapstructure:"zone"`

	clientSecrets   *clientSecrets
	instanceName    string
	privateKeyBytes []byte
	sshTimeout      time.Duration
	stateTimeout    time.Duration
	tpl             *packer.ConfigTemplate
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	md, err := common.DecodeConfig(c, raws...)
	if err != nil {
		return nil, nil, err
	}

	c.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, nil, err
	}
	c.tpl.UserVars = c.PackerUserVars

	// Prepare the errors
	errs := common.CheckUnusedConfig(md)

	// Set defaults.
	if c.Network == "" {
		c.Network = "default"
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		c.ImageName = "packer-{{timestamp}}"
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.MachineType == "" {
		c.MachineType = "n1-standard-1"
	}

	if c.RawSSHTimeout == "" {
		c.RawSSHTimeout = "5m"
	}

	if c.RawStateTimeout == "" {
		c.RawStateTimeout = "5m"
	}

	if c.SSHUsername == "" {
		c.SSHUsername = "root"
	}

	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	// Process Templates
	templates := map[string]*string{
		"bucket_name":         &c.BucketName,
		"client_secrets_file": &c.ClientSecretsFile,
		"image_name":          &c.ImageName,
		"image_description":   &c.ImageDescription,
		"instance_name":       &c.InstanceName,
		"machine_type":        &c.MachineType,
		"network":             &c.Network,
		"passphrase":          &c.Passphrase,
		"private_key_file":    &c.PrivateKeyFile,
		"project_id":          &c.ProjectId,
		"source_image":        &c.SourceImage,
		"ssh_username":        &c.SSHUsername,
		"ssh_timeout":         &c.RawSSHTimeout,
		"state_timeout":       &c.RawStateTimeout,
		"zone":                &c.Zone,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	// Process required parameters.
	if c.BucketName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a bucket_name must be specified"))
	}

	if c.ClientSecretsFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a client_secrets_file must be specified"))
	}

	if c.PrivateKeyFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a private_key_file must be specified"))
	}

	if c.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}

	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}

	// Process timeout settings.
	sshTimeout, err := time.ParseDuration(c.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	c.sshTimeout = sshTimeout

	stateTimeout, err := time.ParseDuration(c.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	c.stateTimeout = stateTimeout

	if c.ClientSecretsFile != "" {
		// Load the client secrets file.
		cs, err := loadClientSecrets(c.ClientSecretsFile)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed parsing client secrets file: %s", err))
		}
		c.clientSecrets = cs
	}

	if c.PrivateKeyFile != "" {
		// Load the private key.
		c.privateKeyBytes, err = processPrivateKeyFile(c.PrivateKeyFile, c.Passphrase)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed loading private key file: %s", err))
		}
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
