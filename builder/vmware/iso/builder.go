package iso

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	driver, err := vmwcommon.NewDriver(&b.config.DriverConfig, &b.config.SSHConfig, b.config.VMName)
	if err != nil {
		return nil, fmt.Errorf("Failed creating VMware driver: %s", err)
	}
	// Before we get deep into the build, make sure ovftool is present and
	// credentials are valid, if we're going to use ovftool.
	if err := driver.VerifyOvfTool(b.config.SkipExport, b.config.SkipValidateCredentials); err != nil {
		return nil, err
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("sshConfig", &b.config.SSHConfig)
	state.Put("driverConfig", &b.config.DriverConfig)
	state.Put("temporaryDevices", []string{}) // Devices (in .vmx) created by packer during building

	steps := []multistep.Step{
		&vmwcommon.StepPrepareTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
		},
		&commonsteps.StepDownload{
			Checksum:    b.config.ISOChecksum,
			Description: "ISO",
			Extension:   b.config.TargetExtension,
			ResultKey:   "iso_path",
			TargetPath:  b.config.TargetPath,
			Url:         b.config.ISOUrls,
		},
		&vmwcommon.StepOutputDir{
			Force:        b.config.PackerForce,
			OutputConfig: &b.config.OutputConfig,
			RemoteType:   b.config.RemoteType,
			VMName:       b.config.VMName,
		},
		&commonsteps.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
			Label:       b.config.FloppyConfig.FloppyLabel,
		},
		&commonsteps.StepCreateCD{
			Files: b.config.CDConfig.CDFiles,
			Label: b.config.CDConfig.CDLabel,
		},
		&vmwcommon.StepRemoteUpload{
			Key:       "floppy_path",
			Message:   "Uploading Floppy to remote machine...",
			DoCleanup: true,
			Checksum:  "none",
		},
		&vmwcommon.StepRemoteUpload{
			Key:       "cd_path",
			Message:   "Uploading CD to remote machine...",
			DoCleanup: true,
			Checksum:  "none",
		},
		&vmwcommon.StepRemoteUpload{
			Key:       "iso_path",
			Message:   "Uploading ISO to remote machine...",
			DoCleanup: b.config.DriverConfig.CleanUpRemoteCache,
			Checksum:  b.config.ISOChecksum,
		},
		&vmwcommon.StepCreateDisks{
			OutputDir:          &b.config.OutputDir,
			CreateMainDisk:     true,
			DiskName:           b.config.DiskName,
			MainDiskSize:       b.config.DiskSize,
			AdditionalDiskSize: b.config.AdditionalDiskSize,
			DiskAdapterType:    b.config.DiskAdapterType,
			DiskTypeId:         b.config.DiskTypeId,
		},
		&stepCreateVMX{},
		&vmwcommon.StepConfigureVMX{
			CustomData:       b.config.VMXData,
			VMName:           b.config.VMName,
			DisplayName:      b.config.VMXDisplayName,
			DiskAdapterType:  b.config.DiskAdapterType,
			CDROMAdapterType: b.config.CdromAdapterType,
		},
		&vmwcommon.StepSuppressMessages{},
		&vmwcommon.StepHTTPIPDiscover{},
		&commonsteps.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
			HTTPAddress: b.config.HTTPAddress,
		},
		&vmwcommon.StepConfigureVNC{
			Enabled:            !b.config.DisableVNC && !b.config.VNCOverWebsocket,
			VNCBindAddress:     b.config.VNCBindAddress,
			VNCPortMin:         b.config.VNCPortMin,
			VNCPortMax:         b.config.VNCPortMax,
			VNCDisablePassword: b.config.VNCDisablePassword,
		},
		&vmwcommon.StepRegister{
			Format:         b.config.Format,
			KeepRegistered: b.config.KeepRegistered,
			SkipExport:     b.config.SkipExport,
		},
		&vmwcommon.StepRun{
			DurationBeforeStop: 5 * time.Second,
			Headless:           b.config.Headless,
		},
		&vmwcommon.StepVNCConnect{
			VNCEnabled:         !b.config.DisableVNC,
			VNCOverWebsocket:   b.config.VNCOverWebsocket,
			InsecureConnection: b.config.InsecureConnection,
			DriverConfig:       &b.config.DriverConfig,
		},
		&vmwcommon.StepVNCBootCommand{
			Config: b.config.VNCConfig,
			VMName: b.config.VMName,
			Ctx:    b.config.ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      driver.CommHost,
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
		&vmwcommon.StepUploadTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
			ToolsUploadPath:   b.config.ToolsUploadPath,
			Ctx:               b.config.ctx,
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},
		&vmwcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&vmwcommon.StepCleanFiles{},
		&vmwcommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
		},
		&vmwcommon.StepConfigureVMX{
			CustomData:  b.config.VMXDataPost,
			SkipFloppy:  true,
			VMName:      b.config.VMName,
			DisplayName: b.config.VMXDisplayName,
		},
		&vmwcommon.StepCleanVMX{
			RemoveEthernetInterfaces: b.config.VMXConfig.VMXRemoveEthernet,
			VNCEnabled:               !b.config.DisableVNC,
		},
		&vmwcommon.StepUploadVMX{
			RemoteType: b.config.RemoteType,
		},
		&vmwcommon.StepExport{
			Format:         b.config.Format,
			SkipExport:     b.config.SkipExport,
			VMName:         b.config.VMName,
			OVFToolOptions: b.config.OVFToolOptions,
			OutputDir:      &b.config.OutputConfig.OutputDir,
		},
	}

	// Run!
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	// Compile the artifact list
	exportOutputPath := state.Get("export_output_path").(string) // set in StepOutputDir
	return vmwcommon.NewArtifact(b.config.RemoteType, b.config.Format, exportOutputPath,
		b.config.VMName, b.config.SkipExport, b.config.KeepRegistered, state)
}
