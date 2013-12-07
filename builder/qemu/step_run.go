package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"strings"
)

// stepRun runs the virtual machine
type stepRun struct {
	BootDrive string
	Message   string
}

type qemuArgsTemplateData struct {
	HTTPIP    string
	HTTPPort  uint
	HTTPDir   string
	OutputDir string
	Name      string
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
	config := state.Get("config").(*config)
	isoPath := state.Get("iso_path").(string)
	vncPort := state.Get("vnc_port").(uint)
	sshHostPort := state.Get("sshHostPort").(uint)
	ui := state.Get("ui").(packer.Ui)

	guiArgument := "sdl"
	vnc := fmt.Sprintf("0.0.0.0:%d", vncPort-5900)
	vmName := config.VMName
	imgPath := filepath.Join(config.OutputDir,
		fmt.Sprintf("%s.%s", vmName, strings.ToLower(config.Format)))

	if config.Headless == true {
		ui.Message("WARNING: The VM will be started in headless mode, as configured.\n" +
			"In headless mode, errors during the boot sequence or OS setup\n" +
			"won't be easily visible. Use at your own discretion.")
		guiArgument = "none"
	}

	defaultArgs := make(map[string]string)
	defaultArgs["-name"] = vmName
	defaultArgs["-machine"] = fmt.Sprintf("type=pc-1.0,accel=%s", config.Accelerator)
	defaultArgs["-display"] = guiArgument
	defaultArgs["-netdev"] = "user,id=user.0"
	defaultArgs["-device"] = fmt.Sprintf("%s,netdev=user.0", config.NetDevice)
	defaultArgs["-drive"] = fmt.Sprintf("file=%s,if=%s", imgPath, config.DiskInterface)
	defaultArgs["-cdrom"] = isoPath
	defaultArgs["-boot"] = bootDrive
	defaultArgs["-m"] = "512m"
	defaultArgs["-redir"] = fmt.Sprintf("tcp:%v::22", sshHostPort)
	defaultArgs["-vnc"] = vnc

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
		tplData := qemuArgsTemplateData{
			"10.0.2.2",
			httpPort,
			config.HTTPDir,
			config.OutputDir,
			config.VMName,
		}
		newQemuArgs, err := processArgs(config.QemuArgs, config.tpl, &tplData)
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
			arg[0] = defaultArgs[key]
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

func processArgs(args [][]string, tpl *packer.ConfigTemplate, tplData *qemuArgsTemplateData) ([][]string, error) {
	var err error

	if args == nil {
		return make([][]string, 0), err
	}

	newArgs := make([][]string, len(args))
	for argsIdx, rowArgs := range args {
		parms := make([]string, len(rowArgs))
		newArgs[argsIdx] = parms
		for i, parm := range rowArgs {
			parms[i], err = tpl.Process(parm, &tplData)
			if err != nil {
				return nil, err
			}
		}
	}

	return newArgs, err
}
