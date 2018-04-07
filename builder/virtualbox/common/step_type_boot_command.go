package common

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/packer/common"
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
	VMName      string
	Ctx         interpolate.Context
}

func (s *StepTypeBootCommand) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

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

	ui.Say("Typing the boot command...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		for _, code := range scancodes(command) {
			if code == "wait" {
				time.Sleep(1 * time.Second)
				continue
			}

			if code == "wait5" {
				time.Sleep(5 * time.Second)
				continue
			}

			if code == "wait10" {
				time.Sleep(10 * time.Second)
				continue
			}

			// Since typing is sometimes so slow, we check for an interrupt
			// in between each character.
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}

			var codes []string

			for i := 0; i < len(code)/2; i++ {
				codes = append(codes, code[i*2:i*2+2])
			}

			args := []string{"controlvm", vmName, "keyboardputscancode"}
			args = append(args, codes...)

			if err := driver.VBoxManage(args...); err != nil {
				err := fmt.Errorf("Error sending boot command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		if pauseFn != nil {
			pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command[%d]: %s", i, command), state)
		}

	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}

func scancodes(message string) []string {
	// Scancodes reference: https://www.win.tue.nl/~aeb/linux/kbd/scancodes-10.html
	//
	// Scancodes represent raw keyboard output and are fed to the VM by the
	// VBoxManage controlvm keyboardputscancode program.
	//
	// Scancodes are recorded here in pairs. The first entry represents
	// the key press and the second entry represents the key release and is
	// derived from the first by the addition of 0x80.
	special := make(map[string][]string)
	special["<bs>"] = []string{"0e", "8e"}
	special["<del>"] = []string{"e053", "e0d3"}
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
	special["<f11>"] = []string{"57", "d7"}
	special["<f12>"] = []string{"58", "d8"}
	special["<return>"] = []string{"1c", "9c"}
	special["<tab>"] = []string{"0f", "8f"}
	special["<up>"] = []string{"e048", "e0c8"}
	special["<down>"] = []string{"e050", "e0d0"}
	special["<left>"] = []string{"e04b", "e0cb"}
	special["<right>"] = []string{"e04d", "e0cd"}
	special["<spacebar>"] = []string{"39", "b9"}
	special["<insert>"] = []string{"e052", "e0d2"}
	special["<home>"] = []string{"e047", "e0c7"}
	special["<end>"] = []string{"e04f", "e0cf"}
	special["<pageUp>"] = []string{"e049", "e0c9"}
	special["<pageDown>"] = []string{"e051", "e0d1"}
	special["<leftAlt>"] = []string{"38", "b8"}
	special["<leftCtrl>"] = []string{"1d", "9d"}
	special["<leftShift>"] = []string{"2a", "aa"}
	special["<rightAlt>"] = []string{"e038", "e0b8"}
	special["<rightCtrl>"] = []string{"e01d", "e09d"}
	special["<rightShift>"] = []string{"36", "b6"}
	special["<leftSuper>"] = []string{"e05b", "e0db"}
	special["<rightSuper>"] = []string{"e05c", "e0dc"}

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

	azOnRegex := regexp.MustCompile("^<(?P<ordinary>[a-zA-Z])On>")
	azOffRegex := regexp.MustCompile("^<(?P<ordinary>[a-zA-Z])Off>")

	result := make([]string, 0, len(message)*2)
	for len(message) > 0 {
		var scancode []string

		if azOnRegex.MatchString(message) {
			m := azOnRegex.FindStringSubmatch(message)
			r, _ := utf8.DecodeRuneInString(m[1])
			message = message[len("<aOn>"):]
			scancodeInt := scancodeMap[r]
			keyShift := unicode.IsUpper(r) || strings.ContainsRune(shiftedChars, r)

			if keyShift {
				scancode = append(scancode, "2a")
			}

			scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))

			log.Printf("Sending char '%c', code '%v', shift %v", r, scancodeInt, keyShift)
		}

		if azOffRegex.MatchString(message) {
			m := azOffRegex.FindStringSubmatch(message)
			r, _ := utf8.DecodeRuneInString(m[1])
			message = message[len("<aOff>"):]
			scancodeInt := scancodeMap[r] + 0x80
			keyShift := unicode.IsUpper(r) || strings.ContainsRune(shiftedChars, r)

			if keyShift {
				scancode = append(scancode, "aa")
			}

			scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))

			log.Printf("Sending char '%c', code '%v', shift %v", r, scancodeInt, keyShift)
		}

		if strings.HasPrefix(message, "<f12On>") {
			scancode = append(scancode, "58")
			message = message[len("<f12On>"):]
			log.Printf("Special code '<f12On>', replacing with: 58")
		}

		if strings.HasPrefix(message, "<leftAltOn>") {
			scancode = append(scancode, "38")
			message = message[len("<leftAltOn>"):]
			log.Printf("Special code '<leftAltOn>' found, replacing with: 38")
		}

		if strings.HasPrefix(message, "<leftCtrlOn>") {
			scancode = append(scancode, "1d")
			message = message[len("<leftCtrlOn>"):]
			log.Printf("Special code '<leftCtrlOn>' found, replacing with: 1d")
		}

		if strings.HasPrefix(message, "<leftShiftOn>") {
			scancode = append(scancode, "2a")
			message = message[len("<leftShiftOn>"):]
			log.Printf("Special code '<leftShiftOn>' found, replacing with: 2a")
		}

		if strings.HasPrefix(message, "<leftSuperOn>") {
			scancode = append(scancode, "e05b")
			message = message[len("<leftSuperOn>"):]
			log.Printf("Special code '<leftSuperOn>' found, replacing with: e05b")
		}

		if strings.HasPrefix(message, "<f12Off>") {
			scancode = append(scancode, "d8")
			message = message[len("<f12Off>"):]
			log.Printf("Special code '<f12Off>' found, replacing with: d8")
		}

		if strings.HasPrefix(message, "<leftAltOff>") {
			scancode = append(scancode, "b8")
			message = message[len("<leftAltOff>"):]
			log.Printf("Special code '<leftAltOff>' found, replacing with: b8")
		}

		if strings.HasPrefix(message, "<leftCtrlOff>") {
			scancode = append(scancode, "9d")
			message = message[len("<leftCtrlOff>"):]
			log.Printf("Special code '<leftCtrlOff>' found, replacing with: 9d")
		}

		if strings.HasPrefix(message, "<leftShiftOff>") {
			scancode = append(scancode, "aa")
			message = message[len("<leftShiftOff>"):]
			log.Printf("Special code '<leftShiftOff>' found, replacing with: aa")
		}

		if strings.HasPrefix(message, "<leftSuperOff>") {
			scancode = append(scancode, "e0db")
			message = message[len("<leftSuperOff>"):]
			log.Printf("Special code '<leftSuperOff>' found, replacing with: e0db")
		}

		if strings.HasPrefix(message, "<rightAltOn>") {
			scancode = append(scancode, "e038")
			message = message[len("<rightAltOn>"):]
			log.Printf("Special code '<rightAltOn>' found, replacing with: e038")
		}

		if strings.HasPrefix(message, "<rightCtrlOn>") {
			scancode = append(scancode, "e01d")
			message = message[len("<rightCtrlOn>"):]
			log.Printf("Special code '<rightCtrlOn>' found, replacing with: e01d")
		}

		if strings.HasPrefix(message, "<rightShiftOn>") {
			scancode = append(scancode, "36")
			message = message[len("<rightShiftOn>"):]
			log.Printf("Special code '<rightShiftOn>' found, replacing with: 36")
		}

		if strings.HasPrefix(message, "<rightSuperOn>") {
			scancode = append(scancode, "e05c")
			message = message[len("<rightSuperOn>"):]
			log.Printf("Special code '<rightSuperOn>' found, replacing with: e05c")
		}

		if strings.HasPrefix(message, "<rightAltOff>") {
			scancode = append(scancode, "e0b8")
			message = message[len("<rightAltOff>"):]
			log.Printf("Special code '<rightAltOff>' found, replacing with: e0b8")
		}

		if strings.HasPrefix(message, "<rightCtrlOff>") {
			scancode = append(scancode, "e09d")
			message = message[len("<rightCtrlOff>"):]
			log.Printf("Special code '<rightCtrlOff>' found, replacing with: e09d")
		}

		if strings.HasPrefix(message, "<rightShiftOff>") {
			scancode = append(scancode, "b6")
			message = message[len("<rightShiftOff>"):]
			log.Printf("Special code '<rightShiftOff>' found, replacing with: b6")
		}

		if strings.HasPrefix(message, "<rightSuperOff>") {
			scancode = append(scancode, "e0dc")
			message = message[len("<rightSuperOff>"):]
			log.Printf("Special code '<rightSuperOff>' found, replacing with: e0dc")
		}

		if strings.HasPrefix(message, "<wait>") {
			log.Printf("Special code <wait> found, will sleep 1 second at this point.")
			scancode = append(scancode, "wait")
			message = message[len("<wait>"):]
		}

		if strings.HasPrefix(message, "<wait5>") {
			log.Printf("Special code <wait5> found, will sleep 5 seconds at this point.")
			scancode = append(scancode, "wait5")
			message = message[len("<wait5>"):]
		}

		if strings.HasPrefix(message, "<wait10>") {
			log.Printf("Special code <wait10> found, will sleep 10 seconds at this point.")
			scancode = append(scancode, "wait10")
			message = message[len("<wait10>"):]
		}

		if scancode == nil {
			for specialCode, specialValue := range special {
				if strings.HasPrefix(message, specialCode) {
					log.Printf("Special code '%s' found, replacing with: %s", specialCode, specialValue)
					scancode = append(scancode, specialValue...)
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
