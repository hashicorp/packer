package qemu

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// stepRun runs the virtual machine
type stepRun struct {
	DiskImage bool

	atLeastVersion2 bool
	ui              packersdk.Ui
}

func (s *stepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	s.ui = state.Get("ui").(packersdk.Ui)

	// Figure out version of qemu; store on step for later use
	rawVersion, err := driver.Version()
	if err != nil {
		err := fmt.Errorf("Error determining qemu version: %s", err)
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}
	qemuVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		err := fmt.Errorf("Error parsing qemu version: %s", err)
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}
	v2 := version.Must(version.NewVersion("2.0"))

	s.atLeastVersion2 = qemuVersion.GreaterThanOrEqual(v2)

	// Generate the qemu command
	command, err := s.getCommandArgs(config, state)
	if err != nil {
		err := fmt.Errorf("Error processing QemuArgs: %s", err)
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// run the qemu command
	if err := driver.Qemu(command...); err != nil {
		err := fmt.Errorf("Error launching VM: %s", err)
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if err := driver.Stop(); err != nil {
		ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
	}
}

func (s *stepRun) getDefaultArgs(config *Config, state multistep.StateBag) map[string]interface{} {

	defaultArgs := make(map[string]interface{})

	// Configure "boot" arguement
	// Run command is different depending whether we're booting from an
	// installation CD or a pre-baked image
	bootDrive := "once=d"
	message := "Starting VM, booting from CD-ROM"
	if s.DiskImage {
		bootDrive = "c"
		message = "Starting VM, booting disk image"
	}
	s.ui.Say(message)
	defaultArgs["-boot"] = bootDrive

	// configure "-qmp" arguments
	if config.QMPEnable {
		defaultArgs["-qmp"] = fmt.Sprintf("unix:%s,server,nowait", config.QMPSocketPath)
	}

	// configure "-name" arguments
	defaultArgs["-name"] = config.VMName

	// Configure "-machine" arguments
	if config.Accelerator == "none" {
		defaultArgs["-machine"] = fmt.Sprintf("type=%s", config.MachineType)
		s.ui.Message("WARNING: The VM will be started with no hardware acceleration.\n" +
			"The installation may take considerably longer to finish.\n")
	} else {
		defaultArgs["-machine"] = fmt.Sprintf("type=%s,accel=%s",
			config.MachineType, config.Accelerator)
	}

	// Firmware
	if config.Firmware != "" {
		defaultArgs["-bios"] = config.Firmware
	}

	// Configure "-netdev" arguments
	defaultArgs["-netdev"] = fmt.Sprintf("bridge,id=user.0,br=%s", config.NetBridge)
	if config.NetBridge == "" {
		defaultArgs["-netdev"] = fmt.Sprintf("user,id=user.0")
		if config.CommConfig.Comm.Type != "none" {
			commHostPort := state.Get("commHostPort").(int)
			defaultArgs["-netdev"] = fmt.Sprintf("user,id=user.0,hostfwd=tcp::%v-:%d", commHostPort, config.CommConfig.Comm.Port())
		}
	}

	// Configure "-vnc" arguments
	// vncPort is always set in stepConfigureVNC, so we don't need to
	// defensively assert
	vncPort := state.Get("vnc_port").(int)
	vncIP := config.VNCBindAddress

	vncRealAddress := fmt.Sprintf("%s:%d", vncIP, vncPort)
	vncPort = vncPort - 5900
	vncArgs := fmt.Sprintf("%s:%d", vncIP, vncPort)
	if config.VNCUsePassword {
		vncArgs = fmt.Sprintf("%s:%d,password", vncIP, vncPort)
	}
	defaultArgs["-vnc"] = vncArgs

	// Track the connection for the user
	vncPass, _ := state.Get("vnc_password").(string)

	message = getVncConnectionMessage(config.Headless, vncRealAddress, vncPass)
	if message != "" {
		s.ui.Message(message)
	}

	// Configure "-m" memory argument
	defaultArgs["-m"] = fmt.Sprintf("%dM", config.MemorySize)

	// Configure "-smp" processor hardware arguments
	if config.CpuCount > 1 {
		defaultArgs["-smp"] = fmt.Sprintf("cpus=%d,sockets=%d", config.CpuCount, config.CpuCount)
	}

	// Configure "-fda" floppy disk attachment
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		defaultArgs["-fda"] = floppyPathRaw.(string)
	} else {
		log.Println("Qemu Builder has no floppy files, not attaching a floppy.")
	}

	// Configure GUI display
	if !config.Headless {
		if s.atLeastVersion2 {
			// FIXME: "none" is a valid display option in qemu but we have
			// departed from the qemu usage here to instaed mean "let qemu
			// set a reasonable default". We need to deprecate this behavior
			// and let users just set "UseDefaultDisplay" if they want to let
			// qemu do its thing.
			if len(config.Display) > 0 && config.Display != "none" {
				defaultArgs["-display"] = config.Display
			} else if !config.UseDefaultDisplay {
				defaultArgs["-display"] = "gtk"
			}
		} else {
			s.ui.Message("WARNING: The version of qemu  on your host doesn't support display mode.\n" +
				"The display parameter will be ignored.")
		}
	}

	deviceArgs, driveArgs := s.getDeviceAndDriveArgs(config, state)
	defaultArgs["-device"] = deviceArgs
	defaultArgs["-drive"] = driveArgs

	return defaultArgs
}

func getVncConnectionMessage(headless bool, vnc string, vncPass string) string {
	// Configure GUI display
	if headless {
		if vnc == "" {
			return "The VM will be run headless, without a GUI, as configured.\n" +
				"If the run isn't succeeding as you expect, please enable the GUI\n" +
				"to inspect the progress of the build."
		}

		if vncPass != "" {
			return fmt.Sprintf(
				"The VM will be run headless, without a GUI. If you want to\n"+
					"view the screen of the VM, connect via VNC to vnc://%s\n"+
					"with the password: %s", vnc, vncPass)
		}

		return fmt.Sprintf(
			"The VM will be run headless, without a GUI. If you want to\n"+
				"view the screen of the VM, connect via VNC without a password to\n"+
				"vnc://%s", vnc)
	}
	return ""
}

func (s *stepRun) getDeviceAndDriveArgs(config *Config, state multistep.StateBag) ([]string, []string) {
	var deviceArgs []string
	var driveArgs []string

	vmName := config.VMName
	imgPath := filepath.Join(config.OutputDir, vmName)

	// Configure virtual hard drives
	if s.atLeastVersion2 {
		drivesToAttach := []string{}

		if v, ok := state.GetOk("qemu_disk_paths"); ok {
			diskFullPaths := v.([]string)
			drivesToAttach = append(drivesToAttach, diskFullPaths...)
		}

		for i, drivePath := range drivesToAttach {
			driveArgumentString := fmt.Sprintf("file=%s,if=%s,cache=%s,discard=%s,format=%s", drivePath, config.DiskInterface, config.DiskCache, config.DiskDiscard, config.Format)
			if config.DiskInterface == "virtio-scsi" {
				// TODO: Megan: Remove this conditional. This, and the code
				// under the TODO below, reproduce the old behavior. While it
				// may be broken, the goal of this commit is to refactor in a way
				// that creates a result that is testably the same as the old
				// code. A pr will follow fixing this broken behavior.
				if i == 0 {
					deviceArgs = append(deviceArgs, fmt.Sprintf("virtio-scsi-pci,id=scsi%d", i))
				}
				// TODO: Megan: When you remove above conditional,
				// set deviceArgs = append(deviceArgs, fmt.Sprintf("scsi-hd,bus=scsi%d.0,drive=drive%d", i, i))
				deviceArgs = append(deviceArgs, fmt.Sprintf("scsi-hd,bus=scsi0.0,drive=drive%d", i))
				driveArgumentString = fmt.Sprintf("if=none,file=%s,id=drive%d,cache=%s,discard=%s,format=%s", drivePath, i, config.DiskCache, config.DiskDiscard, config.Format)
			}
			if config.DetectZeroes != "off" {
				driveArgumentString = fmt.Sprintf("%s,detect-zeroes=%s", driveArgumentString, config.DetectZeroes)
			}
			driveArgs = append(driveArgs, driveArgumentString)
		}
	} else {
		driveArgs = append(driveArgs, fmt.Sprintf("file=%s,if=%s,cache=%s,format=%s", imgPath, config.DiskInterface, config.DiskCache, config.Format))
	}

	deviceArgs = append(deviceArgs, fmt.Sprintf("%s,netdev=user.0", config.NetDevice))

	// Configure virtual CDs
	cdPaths := []string{}
	// Add the installation CD to the run command
	if !config.DiskImage {
		isoPath := state.Get("iso_path").(string)
		cdPaths = append(cdPaths, isoPath)
	}
	// Add our custom CD created from cd_files, if it exists
	cdFilesPath, ok := state.Get("cd_path").(string)
	if ok {
		if cdFilesPath != "" {
			cdPaths = append(cdPaths, cdFilesPath)
		}
	}
	for i, cdPath := range cdPaths {
		if config.CDROMInterface == "" {
			driveArgs = append(driveArgs, fmt.Sprintf("file=%s,media=cdrom", cdPath))
		} else if config.CDROMInterface == "virtio-scsi" {
			driveArgs = append(driveArgs, fmt.Sprintf("file=%s,if=none,index=%d,id=cdrom%d,media=cdrom", cdPath, i, i))
			deviceArgs = append(deviceArgs, "virtio-scsi-device", fmt.Sprintf("scsi-cd,drive=cdrom%d", i))
		} else {
			driveArgs = append(driveArgs, fmt.Sprintf("file=%s,if=%s,index=%d,id=cdrom%d,media=cdrom", cdPath, config.CDROMInterface, i, i))
		}
	}

	return deviceArgs, driveArgs
}

func (s *stepRun) applyUserOverrides(defaultArgs map[string]interface{}, config *Config, state multistep.StateBag) ([]string, error) {
	// Done setting up defaults; time to process user args and defaults together
	// and generate output args

	inArgs := make(map[string][]string)
	if len(config.QemuArgs) > 0 {
		s.ui.Say("Overriding default Qemu arguments with qemuargs template option...")

		commHostPort := 0
		if config.CommConfig.Comm.Type != "none" {
			if v, ok := state.GetOk("commHostPort"); ok {
				commHostPort = v.(int)
			}
		}
		httpIp := state.Get("http_ip").(string)
		httpPort := state.Get("http_port").(int)

		type qemuArgsTemplateData struct {
			HTTPIP      string
			HTTPPort    int
			HTTPDir     string
			OutputDir   string
			Name        string
			SSHHostPort int
		}

		ictx := config.ctx
		ictx.Data = qemuArgsTemplateData{
			HTTPIP:      httpIp,
			HTTPPort:    httpPort,
			HTTPDir:     config.HTTPDir,
			OutputDir:   config.OutputDir,
			Name:        config.VMName,
			SSHHostPort: commHostPort,
		}

		// Interpolate each string in qemuargs
		newQemuArgs, err := processArgs(config.QemuArgs, &ictx)
		if err != nil {
			return nil, err
		}

		// Qemu supports multiple appearances of the same switch. This means
		// each key in the args hash will have an array of string values
		for _, qemuArgs := range newQemuArgs {
			key := qemuArgs[0]
			val := strings.Join(qemuArgs[1:], "")
			if _, ok := inArgs[key]; !ok {
				inArgs[key] = make([]string, 0)
			}
			if len(val) > 0 {
				inArgs[key] = append(inArgs[key], val)
			}
		}
	}

	// get any remaining missing default args from the default settings
	for key := range defaultArgs {
		if _, ok := inArgs[key]; !ok {
			arg := make([]string, 1)
			switch defaultArgs[key].(type) {
			case string:
				arg[0] = defaultArgs[key].(string)
			case []string:
				arg = defaultArgs[key].([]string)
			}
			inArgs[key] = arg
		}
	}

	// Check if we are missing the netDevice #6804
	if x, ok := inArgs["-device"]; ok {
		if !strings.Contains(strings.Join(x, ""), config.NetDevice) {
			inArgs["-device"] = append(inArgs["-device"], fmt.Sprintf("%s,netdev=user.0", config.NetDevice))
		}
	}

	// Flatten to array of strings
	outArgs := make([]string, 0)
	for key, values := range inArgs {
		if len(values) > 0 {
			for idx := range values {
				outArgs = append(outArgs, key, values[idx])
			}
		} else {
			outArgs = append(outArgs, key)
		}
	}

	return outArgs, nil
}

func (s *stepRun) getCommandArgs(config *Config, state multistep.StateBag) ([]string, error) {
	defaultArgs := s.getDefaultArgs(config, state)

	return s.applyUserOverrides(defaultArgs, config, state)
}

func processArgs(args [][]string, ctx *interpolate.Context) ([][]string, error) {
	var err error

	if args == nil {
		return make([][]string, 0), err
	}

	newArgs := make([][]string, len(args))
	for argsIdx, rowArgs := range args {
		parms := make([]string, len(rowArgs))
		newArgs[argsIdx] = parms
		for i, parm := range rowArgs {
			parms[i], err = interpolate.Render(parm, ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return newArgs, err
}
