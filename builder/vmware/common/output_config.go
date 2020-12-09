//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type OutputConfig struct {
	// This is the path on your local machine (the one running Packer) to the
	// directory where the resulting virtual machine will be created.
	// This may be relative or absolute. If relative, the path is relative to
	// the working directory when packer is executed.
	//
	// If you are running a remote esx build, the output_dir is the path on your
	// local machine (the machine running Packer) to which Packer will export
	// the vm if you have `"skip_export": false`. If you want to manage the
	// virtual machine's path on the remote datastore, use `remote_output_dir`.
	//
	// This directory must not exist or be empty prior to running
	// the builder.
	//
	// By default this is output-BUILDNAME where "BUILDNAME" is the name of the
	// build.
	OutputDir string `mapstructure:"output_directory" required:"false"`
	// This is the directoy on your remote esx host where you will save your
	// vm, relative to your remote_datastore.
	//
	// This option's default value is your `vm_name`, and the final path of your
	// vm will be vmfs/volumes/$remote_datastore/$vm_name/$vm_name.vmx where
	// `$remote_datastore` and `$vm_name` match their corresponding template
	// options
	//
	// For example, setting `"remote_output_directory": "path/to/subdir`
	// will create a directory `/vmfs/volumes/remote_datastore/path/to/subdir`.
	//
	// Packer will not create the remote datastore for you; it must already
	// exist. However, Packer will create all directories defined in the option
	// that do not currently exist.
	//
	// This option will be ignored unless you are building on a remote esx host.
	RemoteOutputDir string `mapstructure:"remote_output_directory" required:"false"`
}

func (c *OutputConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	return nil
}
