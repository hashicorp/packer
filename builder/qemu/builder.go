package qemu

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "transcend.qemu"

var accels = map[string]struct{}{
	"none": struct{}{},
	"kvm":  struct{}{},
	"tcg":  struct{}{},
	"xen":  struct{}{},
}

var netDevice = map[string]bool{
	"ne2k_pci":       true,
	"i82551":         true,
	"i82557b":        true,
	"i82559er":       true,
	"rtl8139":        true,
	"e1000":          true,
	"pcnet":          true,
	"virtio":         true,
	"virtio-net":     true,
	"virtio-net-pci": true,
	"usb-net":        true,
	"i82559a":        true,
	"i82559b":        true,
	"i82559c":        true,
	"i82550":         true,
	"i82562":         true,
	"i82557a":        true,
	"i82557c":        true,
	"i82801":         true,
	"vmxnet3":        true,
	"i82558a":        true,
	"i82558b":        true,
}

var diskInterface = map[string]bool{
	"ide":    true,
	"scsi":   true,
	"virtio": true,
}

var diskCache = map[string]bool{
	"writethrough": true,
	"writeback":    true,
	"none":         true,
	"unsafe":       true,
	"directsync":   true,
}

var diskDiscard = map[string]bool{
	"unmap":  true,
	"ignore": true,
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Accelerator     string     `mapstructure:"accelerator"`
	BootCommand     []string   `mapstructure:"boot_command"`
	DiskInterface   string     `mapstructure:"disk_interface"`
	DiskSize        uint       `mapstructure:"disk_size"`
	DiskCache       string     `mapstructure:"disk_cache"`
	DiskDiscard     string     `mapstructure:"disk_discard"`
	FloppyFiles     []string   `mapstructure:"floppy_files"`
	Format          string     `mapstructure:"format"`
	Headless        bool       `mapstructure:"headless"`
	DiskImage       bool       `mapstructure:"disk_image"`
	HTTPDir         string     `mapstructure:"http_directory"`
	HTTPPortMin     uint       `mapstructure:"http_port_min"`
	HTTPPortMax     uint       `mapstructure:"http_port_max"`
	ISOChecksum     string     `mapstructure:"iso_checksum"`
	ISOChecksumType string     `mapstructure:"iso_checksum_type"`
	ISOUrls         []string   `mapstructure:"iso_urls"`
	MachineType     string     `mapstructure:"machine_type"`
	NetDevice       string     `mapstructure:"net_device"`
	OutputDir       string     `mapstructure:"output_directory"`
	QemuArgs        [][]string `mapstructure:"qemuargs"`
	QemuBinary      string     `mapstructure:"qemu_binary"`
	ShutdownCommand string     `mapstructure:"shutdown_command"`
	SSHHostPortMin  uint       `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax  uint       `mapstructure:"ssh_host_port_max"`
	VNCPortMin      uint       `mapstructure:"vnc_port_min"`
	VNCPortMax      uint       `mapstructure:"vnc_port_max"`
	VMName          string     `mapstructure:"vm_name"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHKeyPath     string        `mapstructure:"ssh_key_path"`
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`

	// TODO(mitchellh): deprecate
	RunOnce bool `mapstructure:"run_once"`

	RawBootWait        string `mapstructure:"boot_wait"`
	RawSingleISOUrl    string `mapstructure:"iso_url"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`

	bootWait        time.Duration ``
	shutdownTimeout time.Duration ``
	ctx             interpolate.Context
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"qemuargs",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.DiskCache == "" {
		b.config.DiskCache = "writeback"
	}

	if b.config.DiskDiscard == "" {
		b.config.DiskDiscard = "ignore"
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

	if b.config.MachineType == "" {
		b.config.MachineType = "pc"
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

	if b.config.VNCPortMin == 0 {
		b.config.VNCPortMin = 5900
	}

	if b.config.VNCPortMax == 0 {
		b.config.VNCPortMax = 6000
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

	// TODO: backwards compatibility, write fixer instead
	if b.config.SSHKeyPath != "" {
		b.config.Comm.SSHPrivateKey = b.config.SSHKeyPath
	}
	if b.config.SSHWaitTimeout != 0 {
		b.config.Comm.SSHTimeout = b.config.SSHWaitTimeout
	}

	var errs *packer.MultiError
	warnings := make([]string, 0)

	if es := b.config.Comm.Prepare(&b.config.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if !(b.config.Format == "qcow2" || b.config.Format == "raw") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid format, only 'qcow2' or 'raw' are allowed"))
	}

	if _, ok := accels[b.config.Accelerator]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid accelerator, only 'kvm', 'tcg', 'xen', or 'none' are allowed"))
	}

	if _, ok := netDevice[b.config.NetDevice]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized network device type"))
	}

	if _, ok := diskInterface[b.config.DiskInterface]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized disk interface type"))
	}

	if _, ok := diskCache[b.config.DiskCache]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized disk cache type"))
	}

	if _, ok := diskDiscard[b.config.DiskDiscard]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized disk cache type"))
	}

	if b.config.HTTPPortMin > b.config.HTTPPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("http_port_min must be less than http_port_max"))
	}

	if b.config.ISOChecksumType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		b.config.ISOChecksumType = strings.ToLower(b.config.ISOChecksumType)
		if b.config.ISOChecksumType != "none" {
			if b.config.ISOChecksum == "" {
				errs = packer.MultiErrorAppend(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
			} else {
				b.config.ISOChecksum = strings.ToLower(b.config.ISOChecksum)
			}

			if h := common.HashForType(b.config.ISOChecksumType); h == nil {
				errs = packer.MultiErrorAppend(
					errs,
					fmt.Errorf("Unsupported checksum type: %s", b.config.ISOChecksumType))
			}
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

	b.config.shutdownTimeout, err = time.ParseDuration(b.config.RawShutdownTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	if b.config.VNCPortMin > b.config.VNCPortMax {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	if b.config.QemuArgs == nil {
		b.config.QemuArgs = make([][]string, 0)
	}

	if b.config.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Qemu
	driver, err := b.newDriver(b.config.QemuBinary)
	if err != nil {
		return nil, fmt.Errorf("Failed creating Qemu driver: %s", err)
	}

	steprun := &stepRun{}
	if !b.config.DiskImage {
		steprun.BootDrive = "once=d"
		steprun.Message = "Starting VM, booting from CD-ROM"
	} else {
		steprun.BootDrive = "c"
		steprun.Message = "Starting VM, booting disk image"
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
		new(stepCopyDisk),
		new(stepResizeDisk),
		new(stepHTTPServer),
		new(stepForwardSSH),
		new(stepConfigureVNC),
		steprun,
		&stepBootWait{},
		&stepTypeBootCommand{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
			SSHPort:   commPort,
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
		dir:   b.config.OutputDir,
		f:     files,
		state: make(map[string]interface{}),
	}

	artifact.state["diskName"] = state.Get("disk_filename").(string)
	artifact.state["diskType"] = b.config.Format
	artifact.state["diskSize"] = uint64(b.config.DiskSize)
	artifact.state["domainType"] = b.config.Accelerator

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
