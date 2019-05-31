//go:generate struct-markdown

package qemu

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "transcend.qemu"

var accels = map[string]struct{}{
	"none": {},
	"kvm":  {},
	"tcg":  {},
	"xen":  {},
	"hax":  {},
	"hvf":  {},
	"whpx": {},
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
	"ide":         true,
	"scsi":        true,
	"virtio":      true,
	"virtio-scsi": true,
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

var diskDZeroes = map[string]bool{
	"unmap": true,
	"on":    true,
	"off":   true,
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig   `mapstructure:",squash"`
	common.HTTPConfig     `mapstructure:",squash"`
	common.ISOConfig      `mapstructure:",squash"`
	bootcommand.VNCConfig `mapstructure:",squash"`
	Comm                  communicator.Config `mapstructure:",squash"`
	common.FloppyConfig   `mapstructure:",squash"`
	// Use iso from provided url. Qemu must support
    // curl block device. This defaults to false.
	ISOSkipCache      bool       `mapstructure:"iso_skip_cache" required:"false"`
	// The accelerator type to use when running the VM.
    // This may be none, kvm, tcg, hax, hvf, whpx, or xen. The appropriate
    // software must have already been installed on your build machine to use the
    // accelerator you specified. When no accelerator is specified, Packer will try
    // to use kvm if it is available but will default to tcg otherwise.
	Accelerator       string     `mapstructure:"accelerator" required:"false"`
	// The number of cpus to use when building the VM.
    //  The default is 1 CPU.
	CpuCount          int        `mapstructure:"cpus" required:"false"`
	// The interface to use for the disk. Allowed
    // values include any of ide, scsi, virtio or virtio-scsi*. Note
    // also that any boot commands or kickstart type scripts must have proper
    // adjustments for resulting device names. The Qemu builder uses virtio by
    // default.
	DiskInterface     string     `mapstructure:"disk_interface" required:"false"`
	// The size, in megabytes, of the hard disk to create
    // for the VM. By default, this is 40960 (40 GB).
	DiskSize          uint       `mapstructure:"disk_size" required:"false"`
	// The cache mode to use for disk. Allowed values
    // include any of writethrough, writeback, none, unsafe
    // or directsync. By default, this is set to writeback.
	DiskCache         string     `mapstructure:"disk_cache" required:"false"`
	// The discard mode to use for disk. Allowed values
    // include any of unmap or ignore. By default, this is set to ignore.
	DiskDiscard       string     `mapstructure:"disk_discard" required:"false"`
	// The detect-zeroes mode to use for disk.
    // Allowed values include any of unmap, on or off. Defaults to off.
    // When the value is "off" we don't set the flag in the qemu command, so that
    // Packer still works with old versions of QEMU that don't have this option.
	DetectZeroes      string     `mapstructure:"disk_detect_zeroes" required:"false"`
	// Packer compacts the QCOW2 image using
    // qemu-img convert.  Set this option to true to disable compacting.
    // Defaults to false.
	SkipCompaction    bool       `mapstructure:"skip_compaction" required:"false"`
	// Apply compression to the QCOW2 disk file
    // using qemu-img convert. Defaults to false.
	DiskCompression   bool       `mapstructure:"disk_compression" required:"false"`
	// Either qcow2 or raw, this specifies the output
    // format of the virtual machine image. This defaults to qcow2.
	Format            string     `mapstructure:"format" required:"false"`
	// Packer defaults to building QEMU virtual machines by
    // launching a GUI that shows the console of the machine being built. When this
    // value is set to true, the machine will start without a console.
	Headless          bool       `mapstructure:"headless" required:"false"`
	// Packer defaults to building from an ISO file, this
    // parameter controls whether the ISO URL supplied is actually a bootable
    // QEMU image. When this value is set to true, the machine will either clone
    // the source or use it as a backing file (if use_backing_file is true);
    // then, it will resize the image according to disk_size and boot it.
	DiskImage         bool       `mapstructure:"disk_image" required:"false"`
	// Only applicable when disk_image is true
    // and format is qcow2, set this option to true to create a new QCOW2
    // file that uses the file located at iso_url as a backing file. The new file
    // will only contain blocks that have changed compared to the backing file, so
    // enabling this option can significantly reduce disk usage.
	UseBackingFile    bool       `mapstructure:"use_backing_file" required:"false"`
	// The type of machine emulation to use. Run your
    // qemu binary with the flags -machine help to list available types for
    // your system. This defaults to pc.
	MachineType       string     `mapstructure:"machine_type" required:"false"`
	// The amount of memory to use when building the VM
    // in megabytes. This defaults to 512 megabytes.
	MemorySize        int        `mapstructure:"memory" required:"false"`
	// The driver to use for the network interface. Allowed
    // values ne2k_pci, i82551, i82557b, i82559er, rtl8139, e1000,
    // pcnet, virtio, virtio-net, virtio-net-pci, usb-net, i82559a,
    // i82559b, i82559c, i82550, i82562, i82557a, i82557c, i82801,
    // vmxnet3, i82558a or i82558b. The Qemu builder uses virtio-net by
    // default.
	NetDevice         string     `mapstructure:"net_device" required:"false"`
	// This is the path to the directory where the
    // resulting virtual machine will be created. This may be relative or absolute.
    // If relative, the path is relative to the working directory when packer
    // is executed. This directory must not exist or be empty prior to running
    // the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
    // name of the build.
	OutputDir         string     `mapstructure:"output_directory" required:"false"`
	// Allows complete control over the
    // qemu command line (though not, at this time, qemu-img). Each array of
    // strings makes up a command line switch that overrides matching default
    // switch/value pairs. Any value specified as an empty string is ignored. All
    // values after the switch are concatenated with no separator.
	QemuArgs          [][]string `mapstructure:"qemuargs" required:"false"`
	// The name of the Qemu binary to look for. This
    // defaults to qemu-system-x86_64, but may need to be changed for
    // some platforms. For example qemu-kvm, or qemu-system-i386 may be a
    // better choice for some systems.
	QemuBinary        string     `mapstructure:"qemu_binary" required:"false"`
	// The command to use to gracefully shut down the
    // machine once all the provisioning is done. By default this is an empty
    // string, which tells Packer to just forcefully shut down the machine unless a
    // shutdown command takes place inside script so this may safely be omitted. It
    // is important to add a shutdown_command. By default Packer halts the virtual
    // machine and the file system may not be sync'd. Thus, changes made in a
    // provisioner might not be saved. If one or more scripts require a reboot it is
    // suggested to leave this blank since reboots may fail and specify the final
    // shutdown command in your last script.
	ShutdownCommand   string     `mapstructure:"shutdown_command" required:"false"`
	// The minimum and
    // maximum port to use for the SSH port on the host machine which is forwarded
    // to the SSH port on the guest machine. Because Packer often runs in parallel,
    // Packer will choose a randomly available port in this range to use as the
    // host port. By default this is 2222 to 4444.
	SSHHostPortMin    int        `mapstructure:"ssh_host_port_min" required:"false"`
	SSHHostPortMax    int        `mapstructure:"ssh_host_port_max"`
	// If true, do not pass a -display option
    // to qemu, allowing it to choose the default. This may be needed when running
    // under macOS, and getting errors about sdl not being available.
	UseDefaultDisplay bool       `mapstructure:"use_default_display" required:"false"`
	// The IP address that should be
    // binded to for VNC. By default packer will use 127.0.0.1 for this. If you
    // wish to bind to all interfaces use 0.0.0.0.
	VNCBindAddress    string     `mapstructure:"vnc_bind_address" required:"false"`
	// The minimum and maximum port
    // to use for VNC access to the virtual machine. The builder uses VNC to type
    // the initial boot_command. Because Packer generally runs in parallel,
    // Packer uses a randomly chosen port in this range that appears available. By
    // default this is 5900 to 6000. The minimum and maximum ports are inclusive.
	VNCPortMin        int        `mapstructure:"vnc_port_min" required:"false"`
	VNCPortMax        int        `mapstructure:"vnc_port_max"`
	// This is the name of the image (QCOW2 or IMG) file for
    // the new virtual machine. By default this is packer-BUILDNAME, where
    // "BUILDNAME" is the name of the build. Currently, no file extension will be
    // used unless it is specified in this option.
	VMName            string     `mapstructure:"vm_name" required:"false"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`

	// TODO(mitchellh): deprecate
	RunOnce bool `mapstructure:"run_once"`
	// The amount of time to wait after executing the
    // shutdown_command for the virtual machine to actually shut down. If it
    // doesn't shut down in this time, it is an error. By default, the timeout is
    // 5m or five minutes.
	RawShutdownTimeout string `mapstructure:"shutdown_timeout" required:"false"`

	shutdownTimeout time.Duration ``
	ctx             interpolate.Context
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
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

	var errs *packer.MultiError
	warnings := make([]string, 0)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40960
	}

	if b.config.DiskCache == "" {
		b.config.DiskCache = "writeback"
	}

	if b.config.DiskDiscard == "" {
		b.config.DiskDiscard = "ignore"
	}

	if b.config.DetectZeroes == "" {
		b.config.DetectZeroes = "off"
	}

	if b.config.Accelerator == "" {
		if runtime.GOOS == "windows" {
			b.config.Accelerator = "tcg"
		} else {
			// /dev/kvm is a kernel module that may be loaded if kvm is
			// installed and the host supports VT-x extensions. To make sure
			// this will actually work we need to os.Open() it. If os.Open fails
			// the kernel module was not installed or loaded correctly.
			if fp, err := os.Open("/dev/kvm"); err != nil {
				b.config.Accelerator = "tcg"
			} else {
				fp.Close()
				b.config.Accelerator = "kvm"
			}
		}
		log.Printf("use detected accelerator: %s", b.config.Accelerator)
	} else {
		log.Printf("use specified accelerator: %s", b.config.Accelerator)
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

	if b.config.MemorySize < 10 {
		log.Printf("MemorySize %d is too small, using default: 512", b.config.MemorySize)
		b.config.MemorySize = 512
	}

	if b.config.CpuCount < 1 {
		log.Printf("CpuCount %d too small, using default: 1", b.config.CpuCount)
		b.config.CpuCount = 1
	}

	if b.config.SSHHostPortMin == 0 {
		b.config.SSHHostPortMin = 2222
	}

	if b.config.SSHHostPortMax == 0 {
		b.config.SSHHostPortMax = 4444
	}

	if b.config.VNCBindAddress == "" {
		b.config.VNCBindAddress = "127.0.0.1"
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

	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VNCConfig.Prepare(&b.config.ctx)...)

	if b.config.NetDevice == "" {
		b.config.NetDevice = "virtio-net"
	}

	if b.config.DiskInterface == "" {
		b.config.DiskInterface = "virtio"
	}

	// TODO: backwards compatibility, write fixer instead
	if b.config.SSHWaitTimeout != 0 {
		b.config.Comm.SSHTimeout = b.config.SSHWaitTimeout
	}

	if b.config.ISOSkipCache {
		b.config.ISOChecksumType = "none"
	}

	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	if es := b.config.Comm.Prepare(&b.config.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if !(b.config.Format == "qcow2" || b.config.Format == "raw") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid format, only 'qcow2' or 'raw' are allowed"))
	}

	if b.config.Format != "qcow2" {
		b.config.SkipCompaction = true
		b.config.DiskCompression = false
	}

	if b.config.UseBackingFile && !(b.config.DiskImage && b.config.Format == "qcow2") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("use_backing_file can only be enabled for QCOW2 images and when disk_image is true"))
	}

	if _, ok := accels[b.config.Accelerator]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid accelerator, only 'kvm', 'tcg', 'xen', 'hax', 'hvf', 'whpx', or 'none' are allowed"))
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
			errs, errors.New("unrecognized disk discard type"))
	}

	if _, ok := diskDZeroes[b.config.DetectZeroes]; !ok {
		errs = packer.MultiErrorAppend(
			errs, errors.New("unrecognized disk detect zeroes setting"))
	}

	if !b.config.PackerForce {
		if _, err := os.Stat(b.config.OutputDir); err == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Output directory '%s' already exists. It must not exist.", b.config.OutputDir))
		}
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
	if b.config.SSHHostPortMin < 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be positive"))
	}

	if b.config.VNCPortMin > b.config.VNCPortMax {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	if b.config.QemuArgs == nil {
		b.config.QemuArgs = make([][]string, 0)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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

	steps := []multistep.Step{}
	if !b.config.ISOSkipCache {
		steps = append(steps, &common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			Extension:    b.config.TargetExtension,
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
			Url:          b.config.ISOUrls,
		},
		)
	} else {
		steps = append(steps, &stepSetISO{
			ResultKey: "iso_path",
			Url:       b.config.ISOUrls,
		},
		)
	}

	steps = append(steps, new(stepPrepareOutputDir),
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
		},
		new(stepCreateDisk),
		new(stepCopyDisk),
		new(stepResizeDisk),
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
	)

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			new(stepForwardSSH),
		)
	}

	steps = append(steps,
		new(stepConfigureVNC),
		steprun,
		&stepTypeBootCommand{},
	)

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      commHost,
				SSHConfig: b.config.Comm.SSHConfigFunc(),
				SSHPort:   commPort,
				WinRMPort: commPort,
			},
		)
	}

	steps = append(steps,
		new(common.StepProvision),
	)

	steps = append(steps,
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
	)
	steps = append(steps,
		new(stepShutdown),
	)

	steps = append(steps,
		new(stepConvertDisk),
	)

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
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
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
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
