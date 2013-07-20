package virtualbox

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"
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
//   config *config
//   driver Driver
//   http_port int
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type stepTypeBootCommand struct{}

func (s *stepTypeBootCommand) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	httpPort := state["http_port"].(uint)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	tplData := &bootCommandTemplateData{
		"10.0.2.2",
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the boot command...")
	for _, command := range config.BootCommand {
		var buf bytes.Buffer
		t := template.Must(template.New("boot").Parse(command))
		t.Execute(&buf, tplData)

		for _, code := range scancodes(buf.String()) {
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
			if _, ok := state[multistep.StateCancelled]; ok {
				return multistep.ActionHalt
			}

			if err := driver.VBoxManage("controlvm", vmName, "keyboardputscancode", code); err != nil {
				err := fmt.Errorf("Error sending boot command: %s", err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (*stepTypeBootCommand) Cleanup(map[string]interface{}) {}

func scancodes(message string) []string {
	special := make(map[string][]string)
	special["<bs>"] = []string{"ff", "08"}
	special["<del>"] = []string{"ff", "ff"}
	special["<enter>"] = []string{"1c", "9c"}
	special["<esc>"] = []string{"01", "81"}
	special["<f1>"] = []string{"ff", "be"}
	special["<f2>"] = []string{"ff", "bf"}
	special["<f3>"] = []string{"ff", "c0"}
	special["<f4>"] = []string{"ff", "c1"}
	special["<f5>"] = []string{"ff", "c2"}
	special["<f6>"] = []string{"ff", "c3"}
	special["<f7>"] = []string{"ff", "c4"}
	special["<f8>"] = []string{"ff", "c5"}
	special["<f9>"] = []string{"ff", "c6"}
	special["<f10>"] = []string{"ff", "c7"}
	special["<f11>"] = []string{"ff", "c8"}
	special["<f12>"] = []string{"ff", "c9"}
	special["<return>"] = []string{"1c", "9c"}
	special["<tab>"] = []string{"0f", "8f"}

	shiftedChars := "~!@#$%^&*()_+{}|:\"<>?"

	// Scancodes reference: http://www.win.tue.nl/~aeb/linux/kbd/scancodes-1.html
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
