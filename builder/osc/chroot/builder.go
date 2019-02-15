// The chroot package is able to create an Outscale OMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another OMI and chrooting into that
// directory. It then creates an OMI from that attached drive.
package chroot

import (
	"errors"
	"log"
	"runtime"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "oapi.outscale.chroot"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig       `mapstructure:",squash"`
	osccommon.OMIBlockDevices `mapstructure:",squash"`
	osccommon.OMIConfig       `mapstructure:",squash"`
	osccommon.AccessConfig    `mapstructure:",squash"`

	ChrootMounts      [][]string                 `mapstructure:"chroot_mounts"`
	CommandWrapper    string                     `mapstructure:"command_wrapper"`
	CopyFiles         []string                   `mapstructure:"copy_files"`
	DevicePath        string                     `mapstructure:"device_path"`
	NVMEDevicePath    string                     `mapstructure:"nvme_device_path"`
	FromScratch       bool                       `mapstructure:"from_scratch"`
	MountOptions      []string                   `mapstructure:"mount_options"`
	MountPartition    string                     `mapstructure:"mount_partition"`
	MountPath         string                     `mapstructure:"mount_path"`
	PostMountCommands []string                   `mapstructure:"post_mount_commands"`
	PreMountCommands  []string                   `mapstructure:"pre_mount_commands"`
	RootDeviceName    string                     `mapstructure:"root_device_name"`
	RootVolumeSize    int64                      `mapstructure:"root_volume_size"`
	RootVolumeType    string                     `mapstructure:"root_volume_type"`
	SourceOMI         string                     `mapstructure:"source_omi"`
	SourceOMIFilter   osccommon.OmiFilterOptions `mapstructure:"source_omi_filter"`
	RootVolumeTags    osccommon.TagMap           `mapstructure:"root_volume_tags"`

	ctx interpolate.Context
}

type wrappedCommandTemplate struct {
	Command string
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The outscale-chroot builder only works on Linux environments.")
	}
	return nil, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
