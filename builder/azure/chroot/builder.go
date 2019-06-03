package chroot

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"log"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/builder/amazon/chroot"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ClientConfig client.Config `mapstructure:",squash"`

	FromScratch bool   `mapstructure:"from_scratch"`
	Source      string `mapstructure:"source"`

	CommandWrapper    string     `mapstructure:"command_wrapper"`
	PreMountCommands  []string   `mapstructure:"pre_mount_commands"`
	MountOptions      []string   `mapstructure:"mount_options"`
	MountPartition    string     `mapstructure:"mount_partition"`
	MountPath         string     `mapstructure:"mount_path"`
	PostMountCommands []string   `mapstructure:"post_mount_commands"`
	ChrootMounts      [][]string `mapstructure:"chroot_mounts"`
	CopyFiles         []string   `mapstructure:"copy_files"`

	TemporaryOSDiskName      string `mapstructure:"temporary_os_disk_name"`
	OSDiskSizeGB             int32  `mapstructure:"os_disk_size_gb"`
	OSDiskStorageAccountType string `mapstructure:"os_disk_storage_account_type"`
	OSDiskCacheType          string `mapstructure:"os_disk_cache_type"`
	OSDiskSkipCleanup        bool   `mapstructure:"os_disk_skip_cleanup"`

	ImageResourceID       string `mapstructure:"image_resource_id"`
	ImageOSState          string `mapstructure:"image_os_state"`
	ImageHyperVGeneration string `mapstructure:"image_hyperv_generation"`

	ctx interpolate.Context
}

func (c *Config) GetContext() interpolate.Context {
	return c.ctx
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = azcommon.TemplateFuncs
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

	if b.config.ImageOSState == "" {
		b.config.ImageOSState = string(compute.Generalized)
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
		} else {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("source: %q is not a valid platform image specifier", b.config.Source))
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

	if err := checkOSState(b.config.ImageOSState); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("image_os_state: %v", err))
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

func checkOSState(s string) interface{} {
	for _, v := range compute.PossibleOperatingSystemStateTypesValues() {
		if compute.OperatingSystemStateTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value %v",
		s, compute.PossibleOperatingSystemStateTypesValues())
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
	state.Put("wrappedCommand", chroot.CommandWrapper(wrappedCommand))

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
			ImageOSState:             b.config.ImageOSState,
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

	return nil, nil
}

var _ packer.Builder = &Builder{}
