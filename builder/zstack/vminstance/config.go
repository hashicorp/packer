package vminstance

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the ZStack builder. It stores
// both the publicly settable state as well as the privately generated
// state of the config object.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	BaseUrl             string `mapstructure:"base_url"`
	AccessKey           string `mapstructure:"access_key"`
	KeySecret           string `mapstructure:"key_secret"`
	ShowSecret          bool   `mapstructure:"show_secret"`
	SkipDeleteVm        bool   `mapstructure:"skip_delete_vminstance"`
	SkipProvisionMod    bool   `mapstructure:"skip_provision_mod"`
	SkipPackerSystemTag bool   `mapstructure:"skip_packer_systemtag"`
	ExportImage         bool   `mapstructure:"export_image"`

	ImageName        string   `mapstructure:"image_name"`
	Image            string   `mapstructure:"image_uuid"`
	ImageDescription string   `mapstructure:"image_description"`
	Zone             string   `mapstructure:"zone_uuid"`
	InstanceName     string   `mapstructure:"instance_name"`
	InstanceOffering string   `mapstructure:"instance_offering"`
	L3Network        string   `mapstructure:"l3network_uuid"`
	RawStateTimeout  string   `mapstructure:"state_timeout"`
	RawCreateTimeout string   `mapstructure:"create_timeout"`
	SSHPublicKeyFile string   `mapstructure:"ssh_public_key_file"`
	UserData         string   `mapstructure:"user_data"`
	UserDataFile     string   `mapstructure:"user_data_file"`
	PollReplaceStr   []string `mapstructure:"poll_replace_str"`

	DataVolumeImage string `mapstructure:"datavolume_image_uuid"`
	DataVolumeSize  string `mapstructure:"datavolume_size"`
	CreateWithRoot  bool   `mapstructure:"create_with_root"`
	MountPath       string `mapstructure:"mount_path"`
	FileSystemType  string `mapstructure:"filesystem"`

	Comm          communicator.Config `mapstructure:",squash"`
	stateTimeout  time.Duration
	createTimeout time.Duration
	ctx           interpolate.Context
}

func (c *Config) Check() []error {
	var errs *packer.MultiError

	// Set defaults.
	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a image_uuid must be specified"))
	}

	if c.L3Network == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a l3network_uuid must be specified"))
	}

	if c.InstanceOffering == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a instance_offering must be specified"))
	}

	if c.BaseUrl == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a base_url must be specified"))
	}

	if c.AccessKey == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a access_key must be specified"))
	}

	if c.KeySecret == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a key_secret must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone_uuid must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("user_data and user_data_file mustn't both be specified"))
	}

	if c.DataVolumeImage != "" && c.DataVolumeSize != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("datavolume_image_uuid and datavolume_size mustn't both be specified"))
	}

	if c.DataVolumeImage != "" || c.DataVolumeSize != "" {
		if c.Comm.SSHUsername != "root" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("ssh_username should be root to fdisk / mount datavolume"))
		}
	}

	if c.DataVolumeSize != "" {
		if _, err := getSizeFromStr(c.DataVolumeSize); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.MountPath == "" {
		c.MountPath = MountPath
	}

	if c.FileSystemType == "" {
		c.FileSystemType = "xfs"
	}

	if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		} else {
			data, err := ioutil.ReadFile(c.UserDataFile)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, err)
			} else {
				c.UserData = string(data)
			}
		}
	}

	if len(c.PollReplaceStr) > 0 && len(c.PollReplaceStr) != 2 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("poll_replace_str must be a pair"))
	}

	err := c.CalcCreateTimeout()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}
	err = c.CalcStateTimeout()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if c.Comm.SSHPassword == "" && c.Comm.SSHPrivateKeyFile != "" && c.SSHPublicKeyFile == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("no ssh_public_key_file or ssh_password found, but ssh_private_key_file existed"))
	}

	if errs != nil {
		return errs.Errors
	}
	return nil
}

func (c *Config) CalcStateTimeout() error {
	raw := "120s"
	if c.RawStateTimeout != "" {
		raw = c.RawStateTimeout
	}
	stateTimeout, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("Failed parsing state_timeout: %s", err)
	}
	c.stateTimeout = stateTimeout
	return nil
}

func (c *Config) CalcCreateTimeout() error {
	raw := "60s"
	if c.RawCreateTimeout != "" {
		raw = c.RawCreateTimeout
	}
	createTimeout, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("Failed parsing create_timeout: %s", err)
	}
	c.createTimeout = createTimeout
	return nil
}
