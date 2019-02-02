package hyperone

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/json"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
)

const (
	configPath = "~/.h1-cli/conf.json"
	tokenEnv   = "HYPERONE_TOKEN"

	defaultDiskType     = "ssd"
	defaultImageService = "564639bc052c084e2f2e3266"
	defaultStateTimeout = 5 * time.Minute
	defaultUserName     = "guru"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	APIURL     string `mapstructure:"api_url"`
	Token      string `mapstructure:"token"`
	Project    string `mapstructure:"project"`
	TokenLogin string `mapstructure:"token_login"`

	StateTimeout time.Duration `mapstructure:"state_timeout"`

	SourceImage      string                 `mapstructure:"source_image"`
	ImageName        string                 `mapstructure:"image_name"`
	ImageDescription string                 `mapstructure:"image_description"`
	ImageTags        map[string]interface{} `mapstructure:"image_tags"`
	ImageService     string                 `mapstructure:"image_service"`

	VmType string                 `mapstructure:"vm_type"`
	VmName string                 `mapstructure:"vm_name"`
	VmTags map[string]interface{} `mapstructure:"vm_tags"`

	DiskName string  `mapstructure:"disk_name"`
	DiskType string  `mapstructure:"disk_type"`
	DiskSize float32 `mapstructure:"disk_size"`

	Network             string `mapstructure:"network"`
	PrivateIP           string `mapstructure:"private_ip"`
	PublicIP            string `mapstructure:"public_ip"`
	PublicNetAdpService string `mapstructure:"public_netadp_service"`

	ChrootDisk           bool       `mapstructure:"chroot_disk"`
	ChrootDiskSize       float32    `mapstructure:"chroot_disk_size"`
	ChrootDiskType       string     `mapstructure:"chroot_disk_type"`
	ChrootMountPath      string     `mapstructure:"chroot_mount_path"`
	ChrootMounts         [][]string `mapstructure:"chroot_mounts"`
	ChrootCopyFiles      []string   `mapstructure:"chroot_copy_files"`
	ChrootCommandWrapper string     `mapstructure:"chroot_command_wrapper"`

	MountOptions      []string `mapstructure:"mount_options"`
	MountPartition    string   `mapstructure:"mount_partition"`
	PreMountCommands  []string `mapstructure:"pre_mount_commands"`
	PostMountCommands []string `mapstructure:"post_mount_commands"`

	SSHKeys  []string `mapstructure:"ssh_keys"`
	UserData string   `mapstructure:"user_data"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := &Config{}

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
				"chroot_command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	cliConfig, err := loadCLIConfig()
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = defaultUserName
	}

	if c.Comm.SSHTimeout == 0 {
		c.Comm.SSHTimeout = 10 * time.Minute
	}

	if c.APIURL == "" {
		c.APIURL = os.Getenv("HYPERONE_API_URL")
	}

	if c.Token == "" {
		c.Token = os.Getenv(tokenEnv)

		if c.Token == "" {
			c.Token = cliConfig.Profile.APIKey
		}

		// Fetching token by SSH is available only for the default API endpoint
		if c.TokenLogin != "" && c.APIURL == "" {
			c.Token, err = fetchTokenBySSH(c.TokenLogin)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if c.Project == "" {
		c.Project = cliConfig.Profile.Project.ID
	}

	if c.StateTimeout == 0 {
		c.StateTimeout = defaultStateTimeout
	}

	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			return nil, nil, err
		}
		c.ImageName = name
	}

	if c.ImageService == "" {
		c.ImageService = defaultImageService
	}

	if c.VmName == "" {
		c.VmName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskType == "" {
		c.DiskType = defaultDiskType
	}

	if c.PublicNetAdpService == "" {
		c.PublicNetAdpService = "public"
	}

	if c.ChrootCommandWrapper == "" {
		c.ChrootCommandWrapper = "{{.Command}}"
	}

	if c.ChrootDiskSize == 0 {
		c.ChrootDiskSize = c.DiskSize
	}

	if c.ChrootDiskType == "" {
		c.ChrootDiskType = c.DiskType
	}

	if c.ChrootMountPath == "" {
		path, err := interpolate.Render("/mnt/packer-hyperone-volumes/{{timestamp}}", nil)
		if err != nil {
			return nil, nil, err
		}
		c.ChrootMountPath = path
	}

	if c.ChrootMounts == nil {
		c.ChrootMounts = make([][]string, 0)
	}

	if len(c.ChrootMounts) == 0 {
		c.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	if c.ChrootCopyFiles == nil {
		c.ChrootCopyFiles = []string{"/etc/resolv.conf"}
	}

	if c.MountPartition == "" {
		c.MountPartition = "1"
	}

	// Validation
	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if c.Token == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("token is required"))
	}

	if c.VmType == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("vm type is required"))
	}

	if c.DiskSize == 0 {
		errs = packer.MultiErrorAppend(errs, errors.New("disk size is required"))
	}

	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("source image is required"))
	}

	if c.ChrootDisk {
		if len(c.PreMountCommands) == 0 {
			errs = packer.MultiErrorAppend(errs, errors.New("pre-mount commands are required for chroot disk"))
		}
	}

	for _, mounts := range c.ChrootMounts {
		if len(mounts) != 3 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("each chroot_mounts entry should have three elements"))
			break
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.Token)

	return c, nil, nil
}

type cliConfig struct {
	Profile struct {
		APIKey  string `json:"apiKey"`
		Project struct {
			ID string `json:"_id"`
		} `json:"project"`
	} `json:"profile"`
}

func loadCLIConfig() (cliConfig, error) {
	path, err := homedir.Expand(configPath)
	if err != nil {
		return cliConfig{}, err
	}

	_, err = os.Stat(path)
	if err != nil {
		// Config not found
		return cliConfig{}, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return cliConfig{}, err
	}

	var c cliConfig
	err = json.Unmarshal(content, &c)
	if err != nil {
		return cliConfig{}, err
	}

	return c, nil
}

func getPublicIP(state multistep.StateBag) (string, error) {
	return state.Get("public_ip").(string), nil
}
