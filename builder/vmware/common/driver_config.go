//go:generate struct-markdown

package common

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type DriverConfig struct {
	// When set to true, Packer will cleanup the cache folder where the ISO file is stored during the build on the remote machine.
	// By default, this is set to false.
	CleanUpRemoteCache bool `mapstructure:"cleanup_remote_cache" required:"false"`
	// Path to "VMware Fusion.app". By default this is
	// /Applications/VMware Fusion.app but this setting allows you to
	// customize this.
	FusionAppPath string `mapstructure:"fusion_app_path" required:"false"`
	// The type of remote machine that will be used to
	// build this VM rather than a local desktop product. The only value accepted
	// for this currently is esx5. If this is not set, a desktop product will
	// be used. By default, this is not set.
	RemoteType string `mapstructure:"remote_type" required:"false"`
	// The path to the datastore where the VM will be stored
	// on the ESXi machine.
	RemoteDatastore string `mapstructure:"remote_datastore" required:"false"`
	// The path to the datastore where supporting files
	// will be stored during the build on the remote machine.
	RemoteCacheDatastore string `mapstructure:"remote_cache_datastore" required:"false"`
	// The path where the ISO and/or floppy files will
	// be stored during the build on the remote machine. The path is relative to
	// the remote_cache_datastore on the remote machine.
	RemoteCacheDirectory string `mapstructure:"remote_cache_directory" required:"false"`
	// The host of the remote machine used for access.
	// This is only required if remote_type is enabled.
	RemoteHost string `mapstructure:"remote_host" required:"false"`
	// The SSH port of the remote machine
	RemotePort int `mapstructure:"remote_port" required:"false"`
	// The SSH username used to access the remote machine.
	RemoteUser string `mapstructure:"remote_username" required:"false"`
	// The SSH password for access to the remote machine.
	RemotePassword string `mapstructure:"remote_password" required:"false"`
	// The SSH key for access to the remote machine.
	RemotePrivateKey string `mapstructure:"remote_private_key_file" required:"false"`
	// When Packer is preparing to run a
	// remote esxi build, and export is not disable, by default it runs a no-op
	// ovftool command to make sure that the remote_username and remote_password
	// given are valid. If you set this flag to true, Packer will skip this
	// validation. Default: false.
	SkipValidateCredentials bool `mapstructure:"skip_validate_credentials" required:"false"`
}

func (c *DriverConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.FusionAppPath == "" {
		c.FusionAppPath = os.Getenv("FUSION_APP_PATH")
	}
	if c.FusionAppPath == "" {
		c.FusionAppPath = "/Applications/VMware Fusion.app"
	}
	if c.RemoteUser == "" {
		c.RemoteUser = "root"
	}
	if c.RemoteDatastore == "" {
		c.RemoteDatastore = "datastore1"
	}
	if c.RemoteCacheDatastore == "" {
		c.RemoteCacheDatastore = c.RemoteDatastore
	}
	if c.RemoteCacheDirectory == "" {
		c.RemoteCacheDirectory = "packer_cache"
	}
	if c.RemotePort == 0 {
		c.RemotePort = 22
	}

	if c.RemoteType != "" {
		if c.RemoteHost == "" {
			errs = append(errs,
				fmt.Errorf("remote_host must be specified"))
		}

		if c.RemoteType != "esx5" {
			errs = append(errs,
				fmt.Errorf("Only 'esx5' value is accepted for remote_type"))
		}
	}

	return errs
}

func (c *DriverConfig) Validate(SkipExport bool) error {
	if SkipExport {
		return nil
	}

	if c.RemoteType != "" && c.RemotePassword == "" {
		return fmt.Errorf("exporting the vm from esxi with ovftool requires " +
			"that you set a value for remote_password")
	}

	return nil
}
