//go:generate struct-markdown

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
	// Custom API endpoint URL, compatible with HyperOne.
    // It can also be specified via environment variable HYPERONE_API_URL.
	APIURL     string `mapstructure:"api_url" required:"false"`
	// The authentication token used to access your account.
    // This can be either a session token or a service account token.
    // If not defined, the builder will attempt to find it in the following order:
	Token      string `mapstructure:"token" required:"true"`
	// The id or name of the project. This field is required
    // only if using session tokens. It should be skipped when using service
    // account authentication.
	Project    string `mapstructure:"project" required:"true"`
	// Login (an e-mail) on HyperOne platform. Set this
    // if you want to fetch the token by SSH authentication.
	TokenLogin string `mapstructure:"token_login" required:"false"`
	// Timeout for waiting on the API to complete
    // a request. Defaults to 5m.
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
	// ID or name of the image to launch server from.
	SourceImage      string                 `mapstructure:"source_image" required:"true"`
	// The name of the resulting image. Defaults to
    // "packer-{{timestamp}}"
    // (see configuration templates for more info).
	ImageName        string                 `mapstructure:"image_name" required:"false"`
	// The description of the resulting image.
	ImageDescription string                 `mapstructure:"image_description" required:"false"`
	// Key/value pair tags to
    // add to the created image.
	ImageTags        map[string]interface{} `mapstructure:"image_tags" required:"false"`
	// The service of the resulting image.
	ImageService     string                 `mapstructure:"image_service" required:"false"`
	// ID or name of the type this server should be created with.
	VmType string                 `mapstructure:"vm_type" required:"true"`
	// The name of the created server.
	VmName string                 `mapstructure:"vm_name" required:"false"`
	// Key/value pair tags to
    // add to the created server.
	VmTags map[string]interface{} `mapstructure:"vm_tags" required:"false"`
	// The name of the created disk.
	DiskName string  `mapstructure:"disk_name" required:"false"`
	// The type of the created disk. Defaults to ssd.
	DiskType string  `mapstructure:"disk_type" required:"false"`
	// Size of the created disk, in GiB.
	DiskSize float32 `mapstructure:"disk_size" required:"true"`
	// The ID of the network to attach to the created server.
	Network             string `mapstructure:"network" required:"false"`
	// The ID of the private IP within chosen network
    // that should be assigned to the created server.
	PrivateIP           string `mapstructure:"private_ip" required:"false"`
	// The ID of the public IP that should be assigned to
    // the created server. If network is chosen, the public IP will be associated
    // with server's private IP.
	PublicIP            string `mapstructure:"public_ip" required:"false"`
	// Custom service of public network adapter.
    // Can be useful when using custom api_url. Defaults to public.
	PublicNetAdpService string `mapstructure:"public_netadp_service" required:"false"`

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
	// List of SSH keys by name or id to be added
    // to the server on launch.
	SSHKeys  []string `mapstructure:"ssh_keys" required:"false"`
	// User data to launch with the server. Packer will not
    // automatically wait for a user script to finish before shutting down the
    // instance, this must be handled in a provisioner.
	UserData string   `mapstructure:"user_data" required:"false"`

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
