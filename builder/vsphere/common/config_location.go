//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type LocationConfig

package common

import (
	"fmt"
	"path"
	"strings"
)

type LocationConfig struct {
	// Name of the new VM to create.
	VMName string `mapstructure:"vm_name"`
	// VM folder to create the VM in.
	Folder string `mapstructure:"folder"`
	// ESXi cluster where target VM is created. See
	// [Working with Clusters](#working-with-clusters).
	Cluster string `mapstructure:"cluster"`
	// ESXi host where target VM is created. A full path must be specified if
	// the host is in a folder. For example `folder/host`. See the
	// `Specifying Clusters and Hosts` section above for more details.
	Host string `mapstructure:"host"`
	// VMWare resource pool. Defaults to the root resource pool of the
	// `host` or `cluster`.
	ResourcePool string `mapstructure:"resource_pool"`
	// VMWare datastore. Required if `host` is a cluster, or if `host` has
	// multiple datastores.
	Datastore string `mapstructure:"datastore"`
	// Set this to true if packer should the host for uploading files
	// to the datastore. Defaults to false.
	SetHostForDatastoreUploads bool `mapstructure:"set_host_for_datastore_uploads"`
}

func (c *LocationConfig) Prepare() []error {
	var errs []error

	if c.VMName == "" {
		errs = append(errs, fmt.Errorf("'vm_name' is required"))
	}
	if c.Cluster == "" && c.Host == "" {
		errs = append(errs, fmt.Errorf("'host' or 'cluster' is required"))
	}

	// clean Folder path and remove leading slash as folders are relative within vsphere
	c.Folder = path.Clean(c.Folder)
	c.Folder = strings.TrimLeft(c.Folder, "/")

	return errs
}
