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
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/chroot"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/mitchellh/mapstructure"
)

// BuilderID is the unique ID for this builder
const BuilderID = "azure.chroot"

// Config is the configuration that is chained through the steps and settable
// from the template.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ClientConfig client.Config `mapstructure:",squash"`

	// When set to `true`, starts with an empty, unpartitioned disk. Defaults to `false`.
	FromScratch bool `mapstructure:"from_scratch"`
	// One of the following can be used as a source for an image:
	// - a shared image version resource ID
	// - a managed disk resource ID
	// - a publisher:offer:sku:version specifier for plaform image sources.
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

	// Try to resize the OS disk to this size on the first copy. Disks can only be englarged. If not specified,
	// the disk will keep its original size. Required when using `from_scratch`
	OSDiskSizeGB int32 `mapstructure:"os_disk_size_gb"`
	// The [storage SKU](https://docs.microsoft.com/en-us/rest/api/compute/disks/createorupdate#diskstorageaccounttypes)
	// to use for the OS Disk. Defaults to `Standard_LRS`.
	OSDiskStorageAccountType string `mapstructure:"os_disk_storage_account_type"`
	// The [cache type](https://docs.microsoft.com/en-us/rest/api/compute/images/createorupdate#cachingtypes)
	// specified in the resulting image and for attaching it to the Packer VM. Defaults to `ReadOnly`
	OSDiskCacheType string `mapstructure:"os_disk_cache_type"`

	// The [storage SKU](https://docs.microsoft.com/en-us/rest/api/compute/disks/createorupdate#diskstorageaccounttypes)
	// to use for datadisks. Defaults to `Standard_LRS`.
	DataDiskStorageAccountType string `mapstructure:"data_disk_storage_account_type"`
	// The [cache type](https://docs.microsoft.com/en-us/rest/api/compute/images/createorupdate#cachingtypes)
	// specified in the resulting image and for attaching it to the Packer VM. Defaults to `ReadOnly`
	DataDiskCacheType string `mapstructure:"data_disk_cache_type"`

	// The [Hyper-V generation type](https://docs.microsoft.com/en-us/rest/api/compute/images/createorupdate#hypervgenerationtypes) for Managed Image output.
	// Defaults to `V1`.
	ImageHyperVGeneration string `mapstructure:"image_hyperv_generation"`

	// The id of the temporary OS disk that will be created. Will be generated if not set.
	TemporaryOSDiskID string `mapstructure:"temporary_os_disk_id"`

	// The id of the temporary OS disk snapshot that will be created. Will be generated if not set.
	TemporaryOSDiskSnapshotID string `mapstructure:"temporary_os_disk_snapshot_id"`

	// The prefix for the resource ids of the temporary data disks that will be created. The disks will be suffixed with a number. Will be generated if not set.
	TemporaryDataDiskIDPrefix string `mapstructure:"temporary_data_disk_id_prefix"`

	// The prefix for the resource ids of the temporary data disk snapshots that will be created. The snapshots will be suffixed with a number. Will be generated if not set.
	TemporaryDataDiskSnapshotIDPrefix string `mapstructure:"temporary_data_disk_snapshot_id"`

	// If set to `true`, leaves the temporary disks and snapshots behind in the Packer VM resource group. Defaults to `false`
	SkipCleanup bool `mapstructure:"skip_cleanup"`

	// The managed image to create using this build.
	ImageResourceID string `mapstructure:"image_resource_id"`

	// The shared image to create using this build.
	SharedImageGalleryDestination SharedImageGalleryDestination `mapstructure:"shared_image_destination"`

	ctx interpolate.Context
}

type sourceType string

const (
	sourcePlatformImage sourceType = "PlatformImage"
	sourceDisk          sourceType = "Disk"
	sourceSharedImage   sourceType = "SharedImage"
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

// verify interface implementation
var _ packer.Builder = &Builder{}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	b.config.ctx.Funcs = azcommon.TemplateFuncs
	b.config.ctx.Funcs["vm"] = CreateVMMetadataTemplateFunc()
	md := &mapstructure.Metadata{}
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderID,
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
		Metadata: md,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	var errs *packersdk.MultiError
	var warns []string

	// Defaults
	err = b.config.ClientConfig.SetDefaultValues()
	if err != nil {
		return nil, nil, err
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

	if b.config.TemporaryOSDiskID == "" {
		if def, err := interpolate.Render(
			"/subscriptions/{{ vm `subscription_id` }}/resourceGroups/{{ vm `resource_group` }}/providers/Microsoft.Compute/disks/PackerTemp-osdisk-{{timestamp}}",
			&b.config.ctx); err == nil {
			b.config.TemporaryOSDiskID = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render temporary disk id: %s", err))
		}
	}

	if b.config.TemporaryOSDiskSnapshotID == "" {
		if def, err := interpolate.Render(
			"/subscriptions/{{ vm `subscription_id` }}/resourceGroups/{{ vm `resource_group` }}/providers/Microsoft.Compute/snapshots/PackerTemp-osdisk-snapshot-{{timestamp}}",
			&b.config.ctx); err == nil {
			b.config.TemporaryOSDiskSnapshotID = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render temporary snapshot id: %s", err))
		}
	}

	if b.config.TemporaryDataDiskIDPrefix == "" {
		if def, err := interpolate.Render(
			"/subscriptions/{{ vm `subscription_id` }}/resourceGroups/{{ vm `resource_group` }}/providers/Microsoft.Compute/disks/PackerTemp-datadisk-{{timestamp}}-",
			&b.config.ctx); err == nil {
			b.config.TemporaryDataDiskIDPrefix = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render temporary data disk id prefix: %s", err))
		}
	}

	if b.config.TemporaryDataDiskSnapshotIDPrefix == "" {
		if def, err := interpolate.Render(
			"/subscriptions/{{ vm `subscription_id` }}/resourceGroups/{{ vm `resource_group` }}/providers/Microsoft.Compute/snapshots/PackerTemp-datadisk-snapshot-{{timestamp}}-",
			&b.config.ctx); err == nil {
			b.config.TemporaryDataDiskSnapshotIDPrefix = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render temporary data disk snapshot id prefix: %s", err))
		}
	}

	if b.config.OSDiskStorageAccountType == "" {
		b.config.OSDiskStorageAccountType = string(compute.PremiumLRS)
	}

	if b.config.OSDiskCacheType == "" {
		b.config.OSDiskCacheType = string(compute.CachingTypesReadOnly)
	}

	if b.config.DataDiskStorageAccountType == "" {
		b.config.DataDiskStorageAccountType = string(compute.PremiumLRS)
	}

	if b.config.DataDiskCacheType == "" {
		b.config.DataDiskCacheType = string(compute.CachingTypesReadOnly)
	}

	if b.config.ImageHyperVGeneration == "" {
		b.config.ImageHyperVGeneration = string(compute.V1)
	}

	// checks, accumulate any errors or warnings

	if b.config.FromScratch {
		if b.config.Source != "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("source cannot be specified when building from_scratch"))
		}
		if b.config.OSDiskSizeGB == 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("os_disk_size_gb is required with from_scratch"))
		}
		if len(b.config.PreMountCommands) == 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("pre_mount_commands is required with from_scratch"))
		}
	} else {
		if _, err := client.ParsePlatformImageURN(b.config.Source); err == nil {
			log.Println("Source is platform image:", b.config.Source)
			b.config.sourceType = sourcePlatformImage
		} else if id, err := client.ParseResourceID(b.config.Source); err == nil &&
			strings.EqualFold(id.Provider, "Microsoft.Compute") &&
			strings.EqualFold(id.ResourceType.String(), "disks") {
			log.Println("Source is a disk resource ID:", b.config.Source)
			b.config.sourceType = sourceDisk
		} else if id, err := client.ParseResourceID(b.config.Source); err == nil &&
			strings.EqualFold(id.Provider, "Microsoft.Compute") &&
			strings.EqualFold(id.ResourceType.String(), "galleries/images/versions") {
			log.Println("Source is a shared image ID:", b.config.Source)
			b.config.sourceType = sourceSharedImage
		} else {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("source: %q is not a valid platform image specifier, nor is it a disk resource ID", b.config.Source))
		}
	}

	if err := checkDiskCacheType(b.config.OSDiskCacheType); err != nil {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("os_disk_cache_type: %v", err))
	}

	if err := checkStorageAccountType(b.config.OSDiskStorageAccountType); err != nil {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("os_disk_storage_account_type: %v", err))
	}

	if err := checkDiskCacheType(b.config.DataDiskCacheType); err != nil {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("data_disk_cache_type: %v", err))
	}

	if err := checkStorageAccountType(b.config.DataDiskStorageAccountType); err != nil {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("data_disk_storage_account_type: %v", err))
	}

	if b.config.ImageResourceID != "" {
		r, err := azure.ParseResourceID(b.config.ImageResourceID)
		if err != nil ||
			!strings.EqualFold(r.Provider, "Microsoft.Compute") ||
			!strings.EqualFold(r.ResourceType, "images") {
			errs = packersdk.MultiErrorAppend(fmt.Errorf(
				"image_resource_id: %q is not a valid image resource id", b.config.ImageResourceID))
		}
	}

	if azcommon.StringsContains(md.Keys, "shared_image_destination") {
		e, w := b.config.SharedImageGalleryDestination.Validate("shared_image_destination")
		if len(e) > 0 {
			errs = packersdk.MultiErrorAppend(errs, e...)
		}
		if len(w) > 0 {
			warns = append(warns, w...)
		}
	}

	if !azcommon.StringsContains(md.Keys, "shared_image_destination") && b.config.ImageResourceID == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("image_resource_id or shared_image_destination is required"))
	}

	if err := checkHyperVGeneration(b.config.ImageHyperVGeneration); err != nil {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("image_hyperv_generation: %v", err))
	}

	if errs != nil {
		return nil, warns, errs
	}

	packersdk.LogSecretFilter.Set(b.config.ClientConfig.ClientSecret, b.config.ClientConfig.ClientJWT)
	return nil, warns, nil
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

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	switch runtime.GOOS {
	case "linux", "freebsd":
		break
	default:
		return nil, errors.New("the azure-chroot builder only works on Linux and FreeBSD environments")
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

	// Build the step array from the config
	steps := buildsteps(b.config, info)

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &azcommon.Artifact{
		BuilderIdValue: BuilderID,
		StateData:      map[string]interface{}{"generated_data": state.Get("generated_data")},
		AzureClientSet: azcli,
	}
	if b.config.ImageResourceID != "" {
		artifact.Resources = append(artifact.Resources, b.config.ImageResourceID)
	}
	if e, _ := b.config.SharedImageGalleryDestination.Validate(""); len(e) == 0 {
		artifact.Resources = append(artifact.Resources, b.config.SharedImageGalleryDestination.ResourceID(info.SubscriptionID))
	}
	if b.config.SkipCleanup {
		if d, ok := state.GetOk(stateBagKey_Diskset); ok {
			for _, disk := range d.(Diskset) {
				artifact.Resources = append(artifact.Resources, disk.String())
			}
		}
		if d, ok := state.GetOk(stateBagKey_Snapshotset); ok {
			for _, snapshot := range d.(Diskset) {
				artifact.Resources = append(artifact.Resources, snapshot.String())
			}
		}
	}

	return artifact, nil
}

func buildsteps(config Config, info *client.ComputeInfo) []multistep.Step {
	// Build the steps
	var steps []multistep.Step
	addSteps := func(s ...multistep.Step) { // convenience function
		steps = append(steps, s...)
	}

	e, _ := config.SharedImageGalleryDestination.Validate("")
	hasValidSharedImage := len(e) == 0

	if hasValidSharedImage {
		// validate destination early
		addSteps(
			&StepVerifySharedImageDestination{
				Image:    config.SharedImageGalleryDestination,
				Location: info.Location,
			},
		)
	}

	if config.FromScratch {
		addSteps(&StepCreateNewDiskset{
			OSDiskID:                 config.TemporaryOSDiskID,
			OSDiskSizeGB:             config.OSDiskSizeGB,
			OSDiskStorageAccountType: config.OSDiskStorageAccountType,
			HyperVGeneration:         config.ImageHyperVGeneration,
			Location:                 info.Location})
	} else {
		switch config.sourceType {
		case sourcePlatformImage:
			if pi, err := client.ParsePlatformImageURN(config.Source); err == nil {
				if strings.EqualFold(pi.Version, "latest") {
					addSteps(
						&StepResolvePlatformImageVersion{
							PlatformImage: pi,
							Location:      info.Location,
						})
				}
				addSteps(
					&StepCreateNewDiskset{
						OSDiskID:                 config.TemporaryOSDiskID,
						OSDiskSizeGB:             config.OSDiskSizeGB,
						OSDiskStorageAccountType: config.OSDiskStorageAccountType,
						HyperVGeneration:         config.ImageHyperVGeneration,
						Location:                 info.Location,
						SourcePlatformImage:      pi,

						SkipCleanup: config.SkipCleanup,
					})
			} else {
				panic("Couldn't parse platfrom image urn: " + config.Source + " err: " + err.Error())
			}

		case sourceDisk:
			addSteps(
				&StepVerifySourceDisk{
					SourceDiskResourceID: config.Source,
					Location:             info.Location,
				},
				&StepCreateNewDiskset{
					OSDiskID:                 config.TemporaryOSDiskID,
					OSDiskSizeGB:             config.OSDiskSizeGB,
					OSDiskStorageAccountType: config.OSDiskStorageAccountType,
					HyperVGeneration:         config.ImageHyperVGeneration,
					SourceOSDiskResourceID:   config.Source,
					Location:                 info.Location,

					SkipCleanup: config.SkipCleanup,
				})

		case sourceSharedImage:
			addSteps(
				&StepVerifySharedImageSource{
					SharedImageID:  config.Source,
					SubscriptionID: info.SubscriptionID,
					Location:       info.Location,
				},
				&StepCreateNewDiskset{
					OSDiskID:                   config.TemporaryOSDiskID,
					DataDiskIDPrefix:           config.TemporaryDataDiskIDPrefix,
					OSDiskSizeGB:               config.OSDiskSizeGB,
					OSDiskStorageAccountType:   config.OSDiskStorageAccountType,
					DataDiskStorageAccountType: config.DataDiskStorageAccountType,
					SourceImageResourceID:      config.Source,
					Location:                   info.Location,

					SkipCleanup: config.SkipCleanup,
				})

		default:
			panic(fmt.Errorf("Unknown source type: %+q", config.sourceType))
		}
	}

	addSteps(
		&StepAttachDisk{}, // uses os_disk_resource_id and sets 'device' in stateBag
		&chroot.StepPreMountCommands{
			Commands: config.PreMountCommands,
		},
		&StepMountDevice{
			MountOptions:   config.MountOptions,
			MountPartition: config.MountPartition,
			MountPath:      config.MountPath,
		},
		&chroot.StepPostMountCommands{
			Commands: config.PostMountCommands,
		},
		&chroot.StepMountExtra{
			ChrootMounts: config.ChrootMounts,
		},
		&chroot.StepCopyFiles{
			Files: config.CopyFiles,
		},
		&chroot.StepChrootProvision{},
		&chroot.StepEarlyCleanup{},
	)

	if config.ImageResourceID != "" {
		addSteps(&StepCreateImage{
			ImageResourceID:          config.ImageResourceID,
			ImageOSState:             string(compute.Generalized),
			OSDiskCacheType:          config.OSDiskCacheType,
			OSDiskStorageAccountType: config.OSDiskStorageAccountType,
			Location:                 info.Location,
		})
	}
	if hasValidSharedImage {
		addSteps(
			&StepCreateSnapshotset{
				OSDiskSnapshotID:         config.TemporaryOSDiskSnapshotID,
				DataDiskSnapshotIDPrefix: config.TemporaryDataDiskSnapshotIDPrefix,
				Location:                 info.Location,
				SkipCleanup:              config.SkipCleanup,
			},
			&StepCreateSharedImageVersion{
				Destination:     config.SharedImageGalleryDestination,
				OSDiskCacheType: config.OSDiskCacheType,
				Location:        info.Location,
			},
		)
	}

	return steps
}
