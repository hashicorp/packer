//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

// Package chroot is able to create an Azure managed image without requiring the
// launch of a new virtual machine for every build. It does this by attaching and
// mounting the root disk and chrooting into that directory.
// It then creates a managed image from that attached disk.
package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/chroot"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

// BuilderId is the unique ID for this builder
const BuilderId = "azure.chroot"

// Config is the configuration that is chained through the steps and settable
// from the template.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ClientConfig client.Config `mapstructure:",squash"`

	// When set to `true`, starts with an empty, unpartitioned disk. Defaults to `false`.
	FromScratch bool `mapstructure:"from_scratch"`
	// Either a managed disk resource ID or a publisher:offer:sku:version specifier for plaform image sources.
	Source     string `mapstructure:"source" required:"true"`
	sourceType sourceType

	// How to run shell commands. This may be useful to set environment variables or perhaps run
	// a command with sudo or so on. This is a configuration template where the `.Command` variable
	// is replaced with the command to be run. Defaults to `{{.Command}}`.
	CommandWrapper string `mapstructure:"command_wrapper"`
	// A series of commands to execute after attaching the root volume and before mounting the chroot.
	// This is not required unless using `from_scratch`. If so, this should include any partitioning
	// and filesystem creation commands. The path to the device is provided by `{{.Device}}`.
	PreMountCommands []string `mapstructure:"pre_mount_commands"`
	// Options to supply the `mount` command when mounting devices. Each option will be prefixed with
	// `-o` and supplied to the `mount` command ran by Packer. Because this command is ran in a shell,
	// user discretion is advised. See this manual page for the `mount` command for valid file system specific options.
	MountOptions []string `mapstructure:"mount_options"`
	// The partition number containing the / partition. By default this is the first partition of the volume.
	MountPartition string `mapstructure:"mount_partition"`
	// The path where the volume will be mounted. This is where the chroot environment will be. This defaults
	// to `/mnt/packer-amazon-chroot-volumes/{{.Device}}`. This is a configuration template where the `.Device`
	// variable is replaced with the name of the device where the volume is attached.
	MountPath string `mapstructure:"mount_path"`
	// As `pre_mount_commands`, but the commands are executed after mounting the root device and before the
	// extra mount and copy steps. The device and mount path are provided by `{{.Device}}` and `{{.MountPath}}`.
	PostMountCommands []string `mapstructure:"post_mount_commands"`
	// This is a list of devices to mount into the chroot environment. This configuration parameter requires
	// some additional documentation which is in the "Chroot Mounts" section below. Please read that section
	// for more information on how to use this.
	ChrootMounts [][]string `mapstructure:"chroot_mounts"`
	// Paths to files on the running Azure instance that will be copied into the chroot environment prior to
	// provisioning. Defaults to `/etc/resolv.conf` so that DNS lookups work. Pass an empty list to skip copying
	// `/etc/resolv.conf`. You may need to do this if you're building an image that uses systemd.
	CopyFiles []string `mapstructure:"copy_files"`

	// The name of the temporary disk that will be created in the resource group of the VM that Packer is
	// running on. Will be generated if not set.
	TemporaryOSDiskName string `mapstructure:"temporary_os_disk_name"`
	// Try to resize the OS disk to this size on the first copy. Disks can only be englarged. If not specified,
	// the disk will keep its original size. Required when using `from_scratch`
	OSDiskSizeGB int32 `mapstructure:"os_disk_size_gb"`
	// The [storage SKU](https://docs.microsoft.com/en-us/rest/api/compute/disks/createorupdate#diskstorageaccounttypes)
	// to use for the OS Disk. Defaults to `Standard_LRS`.
	OSDiskStorageAccountType string `mapstructure:"os_disk_storage_account_type"`
	// The [cache type](https://docs.microsoft.com/en-us/rest/api/compute/images/createorupdate#cachingtypes)
	// specified in the resulting image and for attaching it to the Packer VM. Defaults to `ReadOnly`
	OSDiskCacheType string `mapstructure:"os_disk_cache_type"`
	// If set to `true`, leaves the temporary disk behind in the Packer VM resource group. Defaults to `false`
	OSDiskSkipCleanup bool `mapstructure:"os_disk_skip_cleanup"`

	// The image to create using this build.
	ImageResourceID string `mapstructure:"image_resource_id" required:"true"`
	// The [Hyper-V generation type](https://docs.microsoft.com/en-us/rest/api/compute/images/createorupdate#hypervgenerationtypes).
	// Defaults to `V1`.
	ImageHyperVGeneration string `mapstructure:"image_hyperv_generation"`

	ctx interpolate.Context
}

type sourceType string

const (
	sourcePlatformImage sourceType = "PlatformImage"
	sourceDisk          sourceType = "Disk"
)

// GetContext implements ContextProvider to allow steps to use the config context
// for template interpolation
func (c *Config) GetContext() interpolate.Context {
	return c.ctx
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = azcommon.TemplateFuncs
	b.config.ctx.Funcs["vm"] = CreateVMMetadataTemplateFunc()
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				// these fields are interpolated in the steps,
				// when more information is available
				"command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError
	var warns []string

	// Defaults
	err = b.config.ClientConfig.SetDefaultValues()
	if err != nil {
		return nil, err
	}

	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	// set default copy file if we're not giving our own
	if b.config.CopyFiles == nil {
		if !b.config.FromScratch {
			b.config.CopyFiles = []string{"/etc/resolv.conf"}
		}
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-azure-chroot-disks/{{.Device}}"
	}

	if b.config.MountPartition == "" {
		b.config.MountPartition = "1"
	}

	if b.config.TemporaryOSDiskName == "" {

		if def, err := interpolate.Render("PackerTemp-{{timestamp}}", &b.config.ctx); err == nil {
			b.config.TemporaryOSDiskName = def
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("unable to render temporary disk name: %s", err))
		}
	}

	if b.config.OSDiskStorageAccountType == "" {
		b.config.OSDiskStorageAccountType = string(compute.PremiumLRS)
	}

	if b.config.OSDiskCacheType == "" {
		b.config.OSDiskCacheType = string(compute.CachingTypesReadOnly)
	}

	if b.config.ImageHyperVGeneration == "" {
		b.config.ImageHyperVGeneration = string(compute.V1)
	}

	// checks, accumulate any errors or warnings

	if b.config.FromScratch {
		if b.config.Source != "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("source cannot be specified when building from_scratch"))
		}
		if b.config.OSDiskSizeGB == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("os_disk_size_gb is required with from_scratch"))
		}
		if len(b.config.PreMountCommands) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("pre_mount_commands is required with from_scratch"))
		}
	} else {
		if _, err := client.ParsePlatformImageURN(b.config.Source); err == nil {
			log.Println("Source is platform image:", b.config.Source)
			b.config.sourceType = sourcePlatformImage
		} else if id, err := azure.ParseResourceID(b.config.Source); err == nil &&
			strings.EqualFold(id.Provider, "Microsoft.Compute") && strings.EqualFold(id.ResourceType, "disks") {
			log.Println("Source is a disk resource ID:", b.config.Source)
			b.config.sourceType = sourceDisk
		} else {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("source: %q is not a valid platform image specifier, nor is it a disk resource ID", b.config.Source))
		}
	}

	if err := checkDiskCacheType(b.config.OSDiskCacheType); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("os_disk_cache_type: %v", err))
	}

	if err := checkStorageAccountType(b.config.OSDiskStorageAccountType); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("os_disk_storage_account_type: %v", err))
	}

	if b.config.ImageResourceID == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("image_resource_id is required"))
	} else {
		r, err := azure.ParseResourceID(b.config.ImageResourceID)
		if err != nil ||
			!strings.EqualFold(r.Provider, "Microsoft.Compute") ||
			!strings.EqualFold(r.ResourceType, "images") {
			errs = packer.MultiErrorAppend(fmt.Errorf(
				"image_resource_id: %q is not a valid image resource id", b.config.ImageResourceID))
		}
	}

	if err := checkHyperVGeneration(b.config.ImageHyperVGeneration); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("image_hyperv_generation: %v", err))
	}

	if errs != nil {
		return warns, errs
	}

	packer.LogSecretFilter.Set(b.config.ClientConfig.ClientSecret, b.config.ClientConfig.ClientJWT)
	return warns, nil
}

func checkDiskCacheType(s string) interface{} {
	for _, v := range compute.PossibleCachingTypesValues() {
		if compute.CachingTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value %v",
		s, compute.PossibleCachingTypesValues())
}

func checkStorageAccountType(s string) interface{} {
	for _, v := range compute.PossibleDiskStorageAccountTypesValues() {
		if compute.DiskStorageAccountTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value %v",
		s, compute.PossibleDiskStorageAccountTypesValues())
}

func checkHyperVGeneration(s string) interface{} {
	for _, v := range compute.PossibleHyperVGenerationValues() {
		if compute.HyperVGeneration(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value %v",
		s, compute.PossibleHyperVGenerationValues())
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("the azure-chroot builder only works on Linux environments")
	}

	err := b.config.ClientConfig.FillParameters()
	if err != nil {
		return nil, fmt.Errorf("error setting Azure client defaults: %v", err)
	}
	azcli, err := client.New(b.config.ClientConfig, ui.Say)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure client: %v", err)
	}

	wrappedCommand := func(command string) (string, error) {
		ictx := b.config.ctx
		ictx.Data = &struct{ Command string }{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ictx)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("azureclient", azcli)
	state.Put("wrappedCommand", common.CommandWrapper(wrappedCommand))

	info, err := azcli.MetadataClient().GetComputeInfo()
	if err != nil {
		log.Printf("MetadataClient().GetComputeInfo(): error: %+v", err)
		err := fmt.Errorf(
			"Error retrieving information ARM resource ID and location" +
				"of the VM that Packer is running on.\n" +
				"Please verify that Packer is running on a proper Azure VM.")
		ui.Error(err.Error())
		return nil, err
	}

	state.Put("instance", info)

	// Build the steps
	var steps []multistep.Step

	if b.config.FromScratch {
		steps = append(steps,
			&StepCreateNewDisk{
				SubscriptionID:         info.SubscriptionID,
				ResourceGroup:          info.ResourceGroupName,
				DiskName:               b.config.TemporaryOSDiskName,
				DiskSizeGB:             b.config.OSDiskSizeGB,
				DiskStorageAccountType: b.config.OSDiskStorageAccountType,
				HyperVGeneration:       b.config.ImageHyperVGeneration,
				Location:               info.Location,
			})
	} else {
		switch b.config.sourceType {
		case sourcePlatformImage:

			if pi, err := client.ParsePlatformImageURN(b.config.Source); err == nil {
				if strings.EqualFold(pi.Version, "latest") {

					vmi, err := azcli.VirtualMachineImagesClient().GetLatest(ctx, pi.Publisher, pi.Offer, pi.Sku, info.Location)
					if err != nil {
						return nil, fmt.Errorf("error retieving latest version of %q: %v", b.config.Source, err)
					}
					pi.Version = to.String(vmi.Name)
					log.Println("Resolved latest version of source image:", pi.Version)
				}
				steps = append(steps,

					&StepCreateNewDisk{
						SubscriptionID:         info.SubscriptionID,
						ResourceGroup:          info.ResourceGroupName,
						DiskName:               b.config.TemporaryOSDiskName,
						DiskSizeGB:             b.config.OSDiskSizeGB,
						DiskStorageAccountType: b.config.OSDiskStorageAccountType,
						HyperVGeneration:       b.config.ImageHyperVGeneration,
						Location:               info.Location,
						PlatformImage:          pi,

						SkipCleanup: b.config.OSDiskSkipCleanup,
					})
			} else {
				panic("Unknown image source: " + b.config.Source)
			}
		case sourceDisk:
			steps = append(steps,
				&StepVerifySourceDisk{
					SourceDiskResourceID: b.config.Source,
					SubscriptionID:       info.SubscriptionID,
					Location:             info.Location,
				},
				&StepCreateNewDisk{
					SubscriptionID:         info.SubscriptionID,
					ResourceGroup:          info.ResourceGroupName,
					DiskName:               b.config.TemporaryOSDiskName,
					DiskSizeGB:             b.config.OSDiskSizeGB,
					DiskStorageAccountType: b.config.OSDiskStorageAccountType,
					HyperVGeneration:       b.config.ImageHyperVGeneration,
					SourceDiskResourceID:   b.config.Source,
					Location:               info.Location,

					SkipCleanup: b.config.OSDiskSkipCleanup,
				})
		default:
			panic(fmt.Errorf("Unknown source type: %+q", b.config.sourceType))
		}
	}

	steps = append(steps,
		&StepAttachDisk{}, // uses os_disk_resource_id and sets 'device' in stateBag
		&chroot.StepPreMountCommands{
			Commands: b.config.PreMountCommands,
		},
		&StepMountDevice{
			MountOptions:   b.config.MountOptions,
			MountPartition: b.config.MountPartition,
			MountPath:      b.config.MountPath,
		},
		&chroot.StepPostMountCommands{
			Commands: b.config.PostMountCommands,
		},
		&chroot.StepMountExtra{
			ChrootMounts: b.config.ChrootMounts,
		},
		&chroot.StepCopyFiles{
			Files: b.config.CopyFiles,
		},
		&chroot.StepChrootProvision{},
		&chroot.StepEarlyCleanup{},
		&StepCreateImage{
			ImageResourceID:          b.config.ImageResourceID,
			ImageOSState:             string(compute.Generalized),
			OSDiskCacheType:          b.config.OSDiskCacheType,
			OSDiskStorageAccountType: b.config.OSDiskStorageAccountType,
			Location:                 info.Location,
		},
	)

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &azcommon.Artifact{
		Resources:      []string{b.config.ImageResourceID},
		BuilderIdValue: BuilderId,
	}

	return artifact, nil
}

var _ packer.Builder = &Builder{}
