package qemu

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// stepRun runs the virtual machine
type stepRun struct {
	BootDrive string
	Message   string
}

type qemuArgsTemplateData struct {
	HTTPIP      string
	HTTPPort    uint
	HTTPDir     string
	OutputDir   string
	Name        string
	SSHHostPort uint
}

func (s *stepRun) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(s.Message)

	command, err := getCommandArgs(s.BootDrive, state)
	if err != nil {
		err := fmt.Errorf("Error processing QemuArggs: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := driver.Qemu(command...); err != nil {
		err := fmt.Errorf("Error launching VM: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if err := driver.Stop(); err != nil {
		ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
	}
}

func getCommandArgs(bootDrive string, state multistep.StateBag) ([]string, error) {
	config := state.Get("config").(*Config)
	isoPath := state.Get("iso_path").(string)
	vncIP := state.Get("vnc_ip").(string)
	vncPort := state.Get("vnc_port").(uint)
	sshHostPort := state.Get("sshHostPort").(uint)
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	vnc := fmt.Sprintf("%s:%d", vncIP, vncPort-5900)
	vmName := config.VMName
	imgPath := filepath.Join(config.OutputDir, vmName)

	defaultArgs := make(map[string]interface{})
	var deviceArgs []string
	var driveArgs []string

	defaultArgs["-name"] = vmName
	defaultArgs["-machine"] = fmt.Sprintf("type=%s", config.MachineType)
	defaultArgs["-netdev"] = fmt.Sprintf("user,id=user.0,hostfwd=tcp::%v-:%d", sshHostPort, config.Comm.Port())

	qemuVersion, err := driver.Version()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(qemuVersion, ".")
	qemuMajor, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	if qemuMajor >= 2 {
		if config.DiskInterface == "virtio-scsi" {
			deviceArgs = append(deviceArgs, "virtio-scsi-pci,id=scsi0", "scsi-hd,bus=scsi0.0,drive=drive0")
			driveArgs = append(driveArgs, fmt.Sprintf("if=none,file=%s,id=drive0,cache=%s,discard=%s", imgPath, config.DiskCache, config.DiskDiscard))
		} else {
			driveArgs = append(driveArgs, fmt.Sprintf("file=%s,if=%s,cache=%s,discard=%s", imgPath, config.DiskInterface, config.DiskCache, config.DiskDiscard))
		}
	} else {
		driveArgs = append(driveArgs, fmt.Sprintf("file=%s,if=%s,cache=%s", imgPath, config.DiskInterface, config.DiskCache))
	}
	deviceArgs = append(deviceArgs, fmt.Sprintf("%s,netdev=user.0", config.NetDevice))

	if config.Headless == true {
		vncIpRaw, vncIpOk := state.GetOk("vnc_ip")
		vncPortRaw, vncPortOk := state.GetOk("vnc_port")

		if vncIpOk && vncPortOk {
			vncIp := vncIpRaw.(string)
			vncPort := vncPortRaw.(uint)

			ui.Message(fmt.Sprintf(
				"The VM will be run headless, without a GUI. If you want to\n"+
					"view the screen of the VM, connect via VNC without a password to\n"+
					"%s:%d", vncIp, vncPort))
		} else {
			ui.Message("The VM will be run headless, without a GUI, as configured.\n" +
				"If the run isn't succeeding as you expect, please enable the GUI\n" +
				"to inspect the progress of the build.")
		}
	} else {
		if qemuMajor >= 2 {
			defaultArgs["-display"] = "sdl"
		} else {
			ui.Message("WARNING: The version of qemu  on your host doesn't support display mode.\n" +
				"The display parameter will be ignored.")
		}
	}

	defaultArgs["-device"] = deviceArgs
	defaultArgs["-drive"] = driveArgs

	if !config.DiskImage {
		defaultArgs["-cdrom"] = isoPath
	}
	defaultArgs["-boot"] = bootDrive
	defaultArgs["-m"] = "512M"
	defaultArgs["-vnc"] = vnc

	// Append the accelerator to the machine type if it is specified
	if config.Accelerator != "none" {
		defaultArgs["-machine"] = fmt.Sprintf("%s,accel=%s", defaultArgs["-machine"], config.Accelerator)
	} else {
		ui.Message("WARNING: The VM will be started with no hardware acceleration.\n" +
			"The installation may take considerably longer to finish.\n")
	}

	// Determine if we have a floppy disk to attach
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		defaultArgs["-fda"] = floppyPathRaw.(string)
	} else {
		log.Println("Qemu Builder has no floppy files, not attaching a floppy.")
	}

	inArgs := make(map[string][]string)
	if len(config.QemuArgs) > 0 {
		ui.Say("Overriding defaults Qemu arguments with QemuArgs...")

		httpPort := state.Get("http_port").(uint)
		ctx := config.ctx
		ctx.Data = qemuArgsTemplateData{
			"10.0.2.2",
			httpPort,
			config.HTTPDir,
			config.OutputDir,
			config.VMName,
			sshHostPort,
		}
		newQemuArgs, err := processArgs(config.QemuArgs, &ctx)
		if err != nil {
			return nil, err
		}

		// because qemu supports multiple appearances of the same
		// switch, just different values, each key in the args hash
		// will have an array of string values
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
