package common

import (
	"os"

	"github.com/mitchellh/packer/template/interpolate"
)

type DriverConfig struct {
	FusionAppPath        string `mapstructure:"fusion_app_path"`
	RemoteType           string `mapstructure:"remote_type"`
	RemoteDatastore      string `mapstructure:"remote_datastore"`
	RemoteCacheDatastore string `mapstructure:"remote_cache_datastore"`
	RemoteCacheDirectory string `mapstructure:"remote_cache_directory"`
	RemoteHost           string `mapstructure:"remote_host"`
	RemotePort           uint   `mapstructure:"remote_port"`
	RemoteUser           string `mapstructure:"remote_username"`
	RemotePassword       string `mapstructure:"remote_password"`
	RemotePrivateKey     string `mapstructure:"remote_private_key_file"`
}

func (c *DriverConfig) Prepare(ctx *interpolate.Context) []error {
	if c.FusionAppPath == "" {
		c.FusionAppPath = os.Getenv("FUSION_APP_PATH")
	}
	if c.FusionAppPath == "" {
		c.FusionAppPath = "/Applications/VMware Fusion.app"
	}

	return nil
}
