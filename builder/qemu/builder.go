package qemu

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const BuilderId = "transcend.qemu"

var netDevice = map[string]bool{
	"ne2k_pci":   true,
	"i82551":     true,
	"i82557b":    true,
	"i82559er":   true,
	"rtl8139":    true,
	"e1000":      true,
	"pcnet":      true,
	"virtio":     true,
	"virtio-net": true,
	"usb-net":    true,
	"i82559a":    true,
	"i82559b":    true,
	"i82559c":    true,
	"i82550":     true,
	"i82562":     true,
	"i82557a":    true,
	"i82557c":    true,
	"i82801":     true,
	"vmxnet3":    true,
	"i82558a":    true,
	"i82558b":    true,
}

var diskInterface = map[string]bool{
	"ide":    true,
	"scsi":   true,
	"virtio": true,
}

type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	Accelerator     string     `mapstructure:"accelerator"`
	BootCommand     []string   `mapstructure:"boot_command"`
	DiskInterface   string     `mapstructure:"disk_interface"`
	DiskSize        uint       `mapstructure:"disk_size"`
	FloppyFiles     []string   `mapstructure:"floppy_files"`
	Format          string     `mapstructure:"format"`
	Headless        bool       `mapstructure:"headless"`
	HTTPDir         string     `mapstructure:"http_directory"`
	HTTPPortMin     uint       `mapstructure:"http_port_min"`
	HTTPPortMax     uint       `mapstructure:"http_port_max"`
	ISOChecksum     string     `mapstructure:"iso_checksum"`
	ISOChecksumType string     `mapstructure:"iso_checksum_type"`
	ISOUrls         []string   `mapstructure:"iso_urls"`
	NetDevice       string     `mapstructure:"net_device"`
	OutputDir       string     `mapstructure:"output_directory"`
	QemuArgs        [][]string `mapstructure:"qemuargs"`
	QemuBinary      string     `mapstructure:"qemu_binary"`
	ShutdownCommand string     `mapstructure:"shutdown_command"`
	SSHHostPortMin  uint       `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax  uint       `mapstructure:"ssh_host_port_max"`
	SSHPassword     string     `mapstructure:"ssh_password"`
	SSHPort         uint       `mapstructure:"ssh_port"`
	SSHUser         string     `mapstructure:"ssh_username"`
	SSHKeyPath      string     `mapstructure:"ssh_key_path"`
	VNCPortMin      uint       `mapstructure:"vnc_port_min"`
	VNCPortMax      uint       `mapstructure:"vnc_port_max"`
	VMName          string     `mapstructure:"vm_name"`

	// TODO(mitchellh): deprecate
	RunOnce bool `mapstructure:"run_once"`

	RawBootWait        string `mapstructure:"boot_wait"`
	RawSingleISOUrl    string `mapstructure:"iso_url"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	RawSSHWaitTimeout  string `mapstructure:"ssh_wait_timeout"`

	bootWait        time.Duration ``
	shutdownTimeout time.Duration ``
	sshWaitTimeout  time.Duration ``
	tpl             *packer.ConfigTemplate
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.Accelerator == "" {
		b.config.Accelerator = "kvm"
	}

	if b.config.HTTPPortMin == 0 {
		b.config.HTTPPortMin = 8000
	}

	if b.config.HTTPPortMax == 0 {
		b.config.HTTPPortMax = 9000
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.QemuBinary == "" {
		b.config.QemuBinary = "qemu-system-x86_64"
	}

	if b.config.RawBootWait == "" {
		b.config.RawBootWait = "10s"
	}

	if b.config.SSHHostPortMin == 0 {
		b.config.SSHHostPortMin = 2222
	}

	if b.config.SSHHostPortMax == 0 {
		b.config.SSHHostPortMax = 4444
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.VNCPortMin == 0 {
		b.config.VNCPortMin = 5900
	}

	if b.config.VNCPortMax == 0 {
		b.config.VNCPortMax = 6000
	}

	for i, args := range b.config.QemuArgs {
		for j, arg := range args {
			if err := b.config.tpl.Validate(arg); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Error processing qemu-system_x86-64[%d][%d]: %s", i, j, err))
			}
		}
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.Format == "" {
		b.config.Format = "qcow2"
	}

	if b.config.FloppyFiles == nil {
		b.config.FloppyFiles = make([]string, 0)
	}

	if b.config.NetDevice == "" {
		b.config.NetDevice = "virtio-net"
	}

	if b.config.DiskInterface == "" {
		b.config.DiskInterface = "virtio"
	}

	// Errors
	templates := map[string]*string{
		"http_directory":    &b.config.HTTPDir,
		"iso_checksum":      &b.config.ISOChecksum,
		"iso_checksum_type": &b.config.ISOChecksumType,
		"iso_url":           &b.config.RawSingleISOUrl,
		"output_directory":  &b.config.OutputDir,
		"shutdown_command":  &b.config.ShutdownCommand,
		"ssh_key_path":      &b.config.SSHKeyPath,
		"ssh_password":      &b.config.SSHPassword,
		"ssh_username":      &b.config.SSHUser,
		"vm_name":           &b.config.VMName,
		"format":            &b.config.Format,
		"boot_wait":         &b.config.RawBootWait,
		"shutdown_timeout":  &b.config.RawShutdownTimeout,
		"ssh_wait_timeout":  &b.config.RawSSHWaitTimeout,
		"accelerator":       &b.config.Accelerator,
		"net_device":        &b.config.NetDevice,
		"disk_interface":    &b.config.DiskInterface,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	for i, url := range b.config.ISOUrls {
		var err error
		b.config.ISOUrls[i], err = b.config.tpl.Process(url, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing iso_urls[%d]: %s", i, err))
		}
	}

	for i, command := range b.config.BootCommand {
		if err := b.config.tpl.Validate(command); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing boot_command[%d]: %s", i, err))
		}
	}

	for i, file := range b.config.FloppyFiles {
		var err error
		b.config.FloppyFiles[i], err = b.config.tpl.Process(file, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing floppy_files[%d]: %s",
					i, err))
		}
	}

	if !(b.config.Format == "qcow2" || b.config.Format == "raw") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid format, only 'qcow2' or 'raw' are allowed"))
	}

	if !(b.config.Accelerator == "kvm" || b.config.Accelerator == "xen") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid format, only 'kvm' or 'xen' are allowed"))
	}

	if _, ok := netDevice[b.config.NetDevice]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized network device type"))
	}

	if _, ok := diskInterface[b.config.DiskInterface]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized disk interface type"))
	}

	if b.config.HTTPPortMin > b.config.HTTPPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("http_port_min must be less than http_port_max"))
	}

	if b.config.ISOChecksum == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Due to large file sizes, an iso_checksum is required"))
	} else {
		b.config.ISOChecksum = strings.ToLower(b.config.ISOChecksum)
	}

	if b.config.ISOChecksumType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		b.config.ISOChecksumType = strings.ToLower(b.config.ISOChecksumType)
		if h := common.HashForType(b.config.ISOChecksumType); h == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Unsupported checksum type: %s", b.config.ISOChecksumType))
		}
	}

	if b.config.RawSingleISOUrl == "" && len(b.config.ISOUrls) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
	} else if b.config.RawSingleISOUrl != "" && len(b.config.ISOUrls) > 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
	} else if b.config.RawSingleISOUrl != "" {
		b.config.ISOUrls = []string{b.config.RawSingleISOUrl}
	}

	for i, url := range b.config.ISOUrls {
		b.config.ISOUrls[i], err = common.DownloadableURL(url)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed to parse iso_url %d: %s", i+1, err))
		}
	}

	if !b.config.PackerForce {
		if _, err := os.Stat(b.config.OutputDir); err == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Output directory '%s' already exists. It must not exist.", b.config.OutputDir))
		}
	}

	b.config.bootWait, err = time.ParseDuration(b.config.RawBootWait)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
	}

	if b.config.RawShutdownTimeout == "" {
		b.config.RawShutdownTimeout = "5m"
	}

	if b.config.RawSSHWaitTimeout == "" {
		b.config.RawSSHWaitTimeout = "20m"
	}

	b.config.shutdownTimeout, err = time.ParseDuration(b.config.RawShutdownTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if b.config.SSHKeyPath != "" {
		if _, err := os.Stat(b.config.SSHKeyPath); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		} else if _, err := sshKeyToSigner(b.config.SSHKeyPath); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		}
	}

	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	if b.config.SSHUser == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("An ssh_username must be specified."))
	}

	b.config.sshWaitTimeout, err = time.ParseDuration(b.config.RawSSHWaitTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_wait_timeout: %s", err))
	}

	if b.config.VNCPortMin > b.config.VNCPortMax {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	if b.config.QemuArgs == nil {
		b.config.QemuArgs = make([][]string, 0)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Qemu
	driver, err := b.newDriver(b.config.QemuBinary)
	if err != nil {
		return nil, fmt.Errorf("Failed creating Qemu driver: %s", err)
	}

	steps := []multistep.Step{
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
		},
		new(stepPrepareOutputDir),
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		new(stepCreateDisk),
		new(stepHTTPServer),
		new(stepForwardSSH),
		new(stepConfigureVNC),
		&stepRun{
			BootDrive: "once=d",
			Message:   "Starting VM, booting from CD-ROM",
		},
		&stepBootWait{},
		&stepTypeBootCommand{},
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: b.config.sshWaitTimeout,
		},
		new(common.StepProvision),
		new(stepShutdown),
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

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
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(b.config.OutputDir, visit); err != nil {
		return nil, err
	}

	artifact := &Artifact{
		dir: b.config.OutputDir,
		f:   files,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) newDriver(qemuBinary string) (Driver, error) {
	qemuPath, err := exec.LookPath(qemuBinary)
	if err != nil {
		return nil, err
	}

	qemuImgPath, err := exec.LookPath("qemu-img")
	if err != nil {
		return nil, err
	}

	log.Printf("Qemu path: %s, Qemu Image page: %s", qemuPath, qemuImgPath)
	driver := &QemuDriver{
		QemuPath:    qemuPath,
		QemuImgPath: qemuImgPath,
	}

	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
