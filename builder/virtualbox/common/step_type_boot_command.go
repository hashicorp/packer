package common

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/boot_command"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const KeyLeftShift uint32 = 0xFFE1

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
	Name     string
}

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   driver Driver
//   http_port int
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type StepTypeBootCommand struct {
	BootCommand []string
	BootWait    time.Duration
	VMName      string
	Ctx         interpolate.Context
}

func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Wait the for the vm to boot.
	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.BootWait.String()))
		select {
		case <-time.After(s.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	hostIP := "10.0.2.2"
	common.SetHTTPIP(hostIP)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	d := &VBoxBCDriver{
		driver,
		vmName,
	}

	ui.Say("Typing the boot command...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		seq, err := bootcommand.GenerateExpressionSequence(command)
		if err != nil {
			err := fmt.Errorf("Error generating boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// This executes vboxmanage once for each character code. This seems
		// fine for now, but changes the prior behavior. If this becomes
		// a problem, we can always have the driver cache scancodes, and then
		// add a `Flush` method which we can call after this.
		if err := seq.Do(ctx, d); err != nil {
			err := fmt.Errorf("Error running boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if pauseFn != nil {
			pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command[%d]: %s", i, command), state)
		}

	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}

type VBoxBCDriver struct {
	driver Driver
	vmName string
}

func (d *VBoxBCDriver) sendCode(codes []string) error {
	args := []string{"controlvm", d.vmName, "keyboardputscancode"}
	args = append(args, codes...)

	if err := d.driver.VBoxManage(args...); err != nil {
		return err
	}
	return nil

}
func (d *VBoxBCDriver) SendKey(key rune, action bootcommand.KeyAction) error {
	shiftedChars := "~!@#$%^&*()_+{}|:\"<>?"

	scancodeIndex := make(map[string]uint)
	scancodeIndex["1234567890-="] = 0x02
	scancodeIndex["!@#$%^&*()_+"] = 0x02
	scancodeIndex["qwertyuiop[]"] = 0x10
	scancodeIndex["QWERTYUIOP{}"] = 0x10
	scancodeIndex["asdfghjkl;'`"] = 0x1e
	scancodeIndex[`ASDFGHJKL:"~`] = 0x1e
	scancodeIndex[`\zxcvbnm,./`] = 0x2b
	scancodeIndex["|ZXCVBNM<>?"] = 0x2b
	scancodeIndex[" "] = 0x39

	scancodeMap := make(map[rune]uint)
	for chars, start := range scancodeIndex {
		var i uint = 0
		for len(chars) > 0 {
			r, size := utf8.DecodeRuneInString(chars)
			chars = chars[size:]
			scancodeMap[r] = start + i
			i += 1
		}
	}

	keyShift := unicode.IsUpper(key) || strings.ContainsRune(shiftedChars, key)

	var scancode []string

	if action&(bootcommand.KeyOn|bootcommand.KeyPress) != 0 {
		scancodeInt := scancodeMap[key]
		if keyShift {
			scancode = append(scancode, "2a")
		}
		scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))
	}

	if action&(bootcommand.KeyOff|bootcommand.KeyPress) != 0 {
		scancodeInt := scancodeMap[key] + 0x80
		if keyShift {
			scancode = append(scancode, "aa")
		}
		scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))
	}

	for _, sc := range scancode {
		log.Printf("Sending char '%c', code '%s', shift %v", key, sc, keyShift)
	}

	d.sendCode(scancode)

	return nil
}

func (d *VBoxBCDriver) SendSpecial(special string, action bootcommand.KeyAction) error {
	// special contains on/off tuples
	sMap := make(map[string][]string)
	sMap["bs"] = []string{"0e", "8e"}
	sMap["del"] = []string{"e053", "e0d3"}
	sMap["enter"] = []string{"1c", "9c"}
	sMap["esc"] = []string{"01", "81"}
	sMap["f1"] = []string{"3b", "bb"}
	sMap["f2"] = []string{"3c", "bc"}
	sMap["f3"] = []string{"3d", "bd"}
	sMap["f4"] = []string{"3e", "be"}
	sMap["f5"] = []string{"3f", "bf"}
	sMap["f6"] = []string{"40", "c0"}
	sMap["f7"] = []string{"41", "c1"}
	sMap["f8"] = []string{"42", "c2"}
	sMap["f9"] = []string{"43", "c3"}
	sMap["f10"] = []string{"44", "c4"}
	sMap["f11"] = []string{"57", "d7"}
	sMap["f12"] = []string{"58", "d8"}
	sMap["return"] = []string{"1c", "9c"}
	sMap["tab"] = []string{"0f", "8f"}
	sMap["up"] = []string{"e048", "e0c8"}
	sMap["down"] = []string{"e050", "e0d0"}
	sMap["left"] = []string{"e04b", "e0cb"}
	sMap["right"] = []string{"e04d", "e0cd"}
	sMap["spacebar"] = []string{"39", "b9"}
	sMap["insert"] = []string{"e052", "e0d2"}
	sMap["home"] = []string{"e047", "e0c7"}
	sMap["end"] = []string{"e04f", "e0cf"}
	sMap["pageUp"] = []string{"e049", "e0c9"}
	sMap["pageDown"] = []string{"e051", "e0d1"}
	sMap["leftAlt"] = []string{"38", "b8"}
	sMap["leftCtrl"] = []string{"1d", "9d"}
	sMap["leftShift"] = []string{"2a", "aa"}
	sMap["rightAlt"] = []string{"e038", "e0b8"}
	sMap["rightCtrl"] = []string{"e01d", "e09d"}
	sMap["rightShift"] = []string{"36", "b6"}
	sMap["leftSuper"] = []string{"e05b", "e0db"}
	sMap["rightSuper"] = []string{"e05c", "e0dc"}

	keyCode, ok := sMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}

	switch action {
	case bootcommand.KeyOn:
		d.sendCode([]string{keyCode[0]})
	case bootcommand.KeyOff:
		d.sendCode([]string{keyCode[1]})
	case bootcommand.KeyPress:
		d.sendCode(keyCode)
	}
	return nil
}
