package common

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type DriverConfig struct {
	AdditionalDiskSize   []uint `mapstructure:"disk_additional_size"`
	Annotation           string `mapstructure:"annotation"`
	Insecure             bool   `mapstructure:"insecure"`
	RemoteCacheDatastore string `mapstructure:"remote_cache_datastore"`
	RemoteCacheDirectory string `mapstructure:"remote_cache_directory"`
	RemoteCluster        string `mapstructure:"remote_cluster"`
	RemoteDatacenter     string `mapstructure:"remote_datacenter"`
	RemoteDatastore      string `mapstructure:"remote_datastore"`
	RemoteFolder         string `mapstructure:"remote_folder"`
	RemoteHost           string `mapstructure:"remote_host"`
	RemotePassword       string `mapstructure:"remote_password"`
	RemoteResourcePool   string `mapstructure:"remote_resource_pool"`
	RemoteUser           string `mapstructure:"remote_username"`
	Vcenter              string `mapstructure:"vcenter"`
	VMName               string `mapstructure:"vm_name"`
}

func (c *DriverConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	var errs []error

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s", pc.PackerBuildName)
	}

	if c.RemoteHost == "" && c.RemoteCluster == "" {
		//TODO: If the host is empty, its seems that we can create the VM on the cluster (if provided)
		//vcenter will choose the host in the cluster ?
		errs = append(errs, fmt.Errorf("remote_host must be provided"))
	}

	if c.Vcenter == "" {
		if c.RemoteHost == "" && c.RemoteCluster != "" {
			errs = append(errs, fmt.Errorf("For clusters, vcenter must be provided"))
		} else {
			c.Vcenter = c.RemoteHost
		}
	}

	if c.RemoteUser == "" {
		c.RemoteUser = "root"
		user := os.Getenv("GOVC_USERNAME")
		if len(user) != 0 {
			c.RemoteUser = user
		}
	}

	if c.RemotePassword == "" {
		pass := os.Getenv("GOVC_PASSWORD")
		if len(pass) != 0 {
			c.RemotePassword = pass
		}
	}

	if c.RemoteDatacenter == "" {
		c.RemoteDatacenter = "ha-datacenter"
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

	return errs
}
