package common

import (
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
)

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
	Name     string
}

// This step "types" the boot command into the VM via the Hyper-V virtual keyboard
type StepTypeBootCommand struct {
	BootCommand []string
	SwitchName  string
	Ctx         interpolate.Context
}

func (s *StepTypeBootCommand) Run(state multistep.StateBag) multistep.StepAction {
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)

	hostIp, err := driver.GetHostAdapterIpAddressForSwitch(s.SwitchName)

	if err != nil {
		err := fmt.Errorf("Error getting host adapter ip address: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Host IP for the HyperV machine: %s", hostIp))

	common.SetHTTPIP(hostIp)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIp,
		httpPort,
		vmName,
	}

	ui.Say("Typing the boot command...")
	scanCodesToSend := []string{}

	for _, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)

		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		scanCodesToSend = append(scanCodesToSend, scancodes(command)...)
	}

	scanCodesToSendString := strings.Join(scanCodesToSend, " ")

	if err := driver.TypeScanCodes(vmName, scanCodesToSendString); err != nil {
		err := fmt.Errorf("Error sending boot command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}

func scancodes(message string) []string {
	// Scancodes reference: http://www.win.tue.nl/~aeb/linux/kbd/scancodes-1.html
	//
	// Scancodes represent raw keyboard output and are fed to the VM by using
	// powershell to use Msvm_Keyboard
	//
	// Scancodes are recorded here in pairs. The first entry represents
	// the key press and the second entry represents the key release and is
	// derived from the first by the addition of 0x80.
	special := make(map[string][]string)
	special["<bs>"] = []string{"0e", "8e"}
	special["<del>"] = []string{"53", "d3"}
	special["<enter>"] = []string{"1c", "9c"}
	special["<esc>"] = []string{"01", "81"}
	special["<f1>"] = []string{"3b", "bb"}
	special["<f2>"] = []string{"3c", "bc"}
	special["<f3>"] = []string{"3d", "bd"}
	special["<f4>"] = []string{"3e", "be"}
	special["<f5>"] = []string{"3f", "bf"}
	special["<f6>"] = []string{"40", "c0"}
	special["<f7>"] = []string{"41", "c1"}
	special["<f8>"] = []string{"42", "c2"}
	special["<f9>"] = []string{"43", "c3"}
	special["<f10>"] = []string{"44", "c4"}
	special["<return>"] = []string{"1c", "9c"}
	special["<tab>"] = []string{"0f", "8f"}
	special["<up>"] = []string{"48", "c8"}
	special["<down>"] = []string{"50", "d0"}
	special["<left>"] = []string{"4b", "cb"}
	special["<right>"] = []string{"4d", "cd"}
	special["<spacebar>"] = []string{"39", "b9"}
	special["<insert>"] = []string{"52", "d2"}
	special["<home>"] = []string{"47", "c7"}
	special["<end>"] = []string{"4f", "cf"}
	special["<pageUp>"] = []string{"49", "c9"}
	special["<pageDown>"] = []string{"51", "d1"}
	special["<leftAlt>"] = []string{"38", "b8"}
	special["<leftCtrl>"] = []string{"1d", "9d"}
	special["<leftShift>"] = []string{"2a", "aa"}
	special["<rightAlt>"] = []string{"e038", "e0b8"}
	special["<rightCtrl>"] = []string{"e01d", "e09d"}
	special["<rightShift>"] = []string{"36", "b6"}

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

	result := make([]string, 0, len(message)*2)
	for len(message) > 0 {
		var scancode []string

		if strings.HasPrefix(message, "<leftAltOn>") {
			scancode = []string{"38"}
			message = message[len("<leftAltOn>"):]
			log.Printf("Special code '<leftAltOn>' found, replacing with: 38")
		}

		if strings.HasPrefix(message, "<leftCtrlOn>") {
			scancode = []string{"1d"}
			message = message[len("<leftCtrlOn>"):]
			log.Printf("Special code '<leftCtrlOn>' found, replacing with: 1d")
		}

		if strings.HasPrefix(message, "<leftShiftOn>") {
			scancode = []string{"2a"}
			message = message[len("<leftShiftOn>"):]
			log.Printf("Special code '<leftShiftOn>' found, replacing with: 2a")
		}

		if strings.HasPrefix(message, "<leftAltOff>") {
			scancode = []string{"b8"}
			message = message[len("<leftAltOff>"):]
			log.Printf("Special code '<leftAltOff>' found, replacing with: b8")
		}

		if strings.HasPrefix(message, "<leftCtrlOff>") {
			scancode = []string{"9d"}
			message = message[len("<leftCtrlOff>"):]
			log.Printf("Special code '<leftCtrlOff>' found, replacing with: 9d")
		}

		if strings.HasPrefix(message, "<leftShiftOff>") {
			scancode = []string{"aa"}
			message = message[len("<leftShiftOff>"):]
			log.Printf("Special code '<leftShiftOff>' found, replacing with: aa")
		}

		if strings.HasPrefix(message, "<rightAltOn>") {
			scancode = []string{"e038"}
			message = message[len("<rightAltOn>"):]
			log.Printf("Special code '<rightAltOn>' found, replacing with: e038")
		}

		if strings.HasPrefix(message, "<rightCtrlOn>") {
			scancode = []string{"e01d"}
			message = message[len("<rightCtrlOn>"):]
			log.Printf("Special code '<rightCtrlOn>' found, replacing with: e01d")
		}

		if strings.HasPrefix(message, "<rightShiftOn>") {
			scancode = []string{"36"}
			message = message[len("<rightShiftOn>"):]
			log.Printf("Special code '<rightShiftOn>' found, replacing with: 36")
		}

		if strings.HasPrefix(message, "<rightAltOff>") {
			scancode = []string{"e0b8"}
			message = message[len("<rightAltOff>"):]
			log.Printf("Special code '<rightAltOff>' found, replacing with: e0b8")
		}

		if strings.HasPrefix(message, "<rightCtrlOff>") {
			scancode = []string{"e09d"}
			message = message[len("<rightCtrlOff>"):]
			log.Printf("Special code '<rightCtrlOff>' found, replacing with: e09d")
		}

		if strings.HasPrefix(message, "<rightShiftOff>") {
			scancode = []string{"b6"}
			message = message[len("<rightShiftOff>"):]
			log.Printf("Special code '<rightShiftOff>' found, replacing with: b6")
		}

		if strings.HasPrefix(message, "<wait>") {
			log.Printf("Special code <wait> found, will sleep 1 second at this point.")
			scancode = []string{"wait"}
			message = message[len("<wait>"):]
		}

		if strings.HasPrefix(message, "<wait5>") {
			log.Printf("Special code <wait5> found, will sleep 5 seconds at this point.")
			scancode = []string{"wait5"}
			message = message[len("<wait5>"):]
		}

		if strings.HasPrefix(message, "<wait10>") {
			log.Printf("Special code <wait10> found, will sleep 10 seconds at this point.")
			scancode = []string{"wait10"}
			message = message[len("<wait10>"):]
		}

		if scancode == nil {
			for specialCode, specialValue := range special {
				if strings.HasPrefix(message, specialCode) {
					log.Printf("Special code '%s' found, replacing with: %s", specialCode, specialValue)
					scancode = specialValue
					message = message[len(specialCode):]
					break
				}
			}
		}

		if scancode == nil {
			r, size := utf8.DecodeRuneInString(message)
			message = message[size:]
			scancodeInt := scancodeMap[r]
			keyShift := unicode.IsUpper(r) || strings.ContainsRune(shiftedChars, r)

			scancode = make([]string, 0, 4)
			if keyShift {
				scancode = append(scancode, "2a")
			}

			scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))

			if keyShift {
				scancode = append(scancode, "aa")
			}

			scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt+0x80))
			log.Printf("Sending char '%c', code '%v', shift %v", r, scancode, keyShift)
		}

		result = append(result, scancode...)
	}

	return result
}
