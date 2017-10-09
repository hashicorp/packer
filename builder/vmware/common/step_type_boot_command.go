package common

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
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
//   http_port int
//   ui     packer.Ui
//   vnc_port uint
//
// Produces:
//   <nothing>
type StepTypeBootCommand struct {
	BootCommand []string
	VMName      string
	Ctx         interpolate.Context
	Skip        bool
}

func (s *StepTypeBootCommand) Run(state multistep.StateBag) multistep.StepAction {
	if s.Skip {
		log.Println("Skipping boot command step...")
		return multistep.ActionContinue
	}

	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vncIp := state.Get("vnc_ip").(string)
	vncPort := state.Get("vnc_port").(uint)
	vncPassword := state.Get("vnc_password")

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	// Connect to VNC
	ui.Say("Connecting to VM via VNC")
	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", vncIp, vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	var auth []vnc.ClientAuth

	if vncPassword != nil && len(vncPassword.(string)) > 0 {
		auth = []vnc.ClientAuth{&vnc.PasswordAuth{Password: vncPassword.(string)}}
	} else {
		auth = []vnc.ClientAuth{new(vnc.ClientAuthNone)}
	}

	c, err := vnc.Client(nc, &vnc.ClientConfig{Auth: auth, Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	// Determine the host IP
	var ipFinder HostIPFinder
	if finder, ok := driver.(HostIPFinder); ok {
		ipFinder = finder
	} else if runtime.GOOS == "windows" {
		ipFinder = new(VMnetNatConfIPFinder)
	} else {
		ipFinder = &IfconfigIPFinder{Device: "vmnet8"}
	}

	hostIP, err := ipFinder.HostIP()
	if err != nil {
		err := fmt.Errorf("Error detecting host IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Host IP for the VMware machine: %s", hostIP)
	common.SetHTTPIP(hostIP)

	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	ui.Say("Typing the boot command over VNC...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Check for interrupts between typing things so we can cancel
		// since this isn't the fastest thing.
		if _, ok := state.GetOk(multistep.StateCancelled); ok {
			return multistep.ActionHalt
		}

		if pauseFn != nil {
			pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command[%d]: %s", i, command), state)
		}

		vncSendString(c, command)
	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}

func vncSendString(c *vnc.ClientConn, original string) {
	// Scancodes reference: https://github.com/qemu/qemu/blob/master/ui/vnc_keysym.h
	special := make(map[string]uint32)
	special["<bs>"] = 0xFF08
	special["<del>"] = 0xFFFF
	special["<enter>"] = 0xFF0D
	special["<esc>"] = 0xFF1B
	special["<f1>"] = 0xFFBE
	special["<f2>"] = 0xFFBF
	special["<f3>"] = 0xFFC0
	special["<f4>"] = 0xFFC1
	special["<f5>"] = 0xFFC2
	special["<f6>"] = 0xFFC3
	special["<f7>"] = 0xFFC4
	special["<f8>"] = 0xFFC5
	special["<f9>"] = 0xFFC6
	special["<f10>"] = 0xFFC7
	special["<f11>"] = 0xFFC8
	special["<f12>"] = 0xFFC9
	special["<return>"] = 0xFF0D
	special["<tab>"] = 0xFF09
	special["<up>"] = 0xFF52
	special["<down>"] = 0xFF54
	special["<left>"] = 0xFF51
	special["<right>"] = 0xFF53
	special["<spacebar>"] = 0x020
	special["<insert>"] = 0xFF63
	special["<home>"] = 0xFF50
	special["<end>"] = 0xFF57
	special["<pageUp>"] = 0xFF55
	special["<pageDown>"] = 0xFF56
	special["<leftAlt>"] = 0xFFE9
	special["<leftCtrl>"] = 0xFFE3
	special["<leftShift>"] = 0xFFE1
	special["<rightAlt>"] = 0xFFEA
	special["<rightCtrl>"] = 0xFFE4
	special["<rightShift>"] = 0xFFE2

	shiftedChars := "~!@#$%^&*()_+{}|:\"<>?"

	// We delay (default 100ms) between each key event to allow for CPU or
	// network latency. See PackerKeyEnv for tuning.
	keyInterval := common.PackerKeyDefault
	if delay, err := time.ParseDuration(os.Getenv(common.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}

	// TODO(mitchellh): Ripe for optimizations of some point, perhaps.
	for len(original) > 0 {
		var keyCode uint32
		keyShift := false

		if strings.HasPrefix(original, "<leftAltOn>") {
			keyCode = special["<leftAlt>"]
			original = original[len("<leftAltOn>"):]
			log.Printf("Special code '<leftAltOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<leftCtrlOn>") {
			keyCode = special["<leftCtrl>"]
			original = original[len("<leftCtrlOn>"):]
			log.Printf("Special code '<leftCtrlOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<leftShiftOn>") {
			keyCode = special["<leftShift>"]
			original = original[len("<leftShiftOn>"):]
			log.Printf("Special code '<leftShiftOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<leftAltOff>") {
			keyCode = special["<leftAlt>"]
			original = original[len("<leftAltOff>"):]
			log.Printf("Special code '<leftAltOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<leftCtrlOff>") {
			keyCode = special["<leftCtrl>"]
			original = original[len("<leftCtrlOff>"):]
			log.Printf("Special code '<leftCtrlOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<leftShiftOff>") {
			keyCode = special["<leftShift>"]
			original = original[len("<leftShiftOff>"):]
			log.Printf("Special code '<leftShiftOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightAltOn>") {
			keyCode = special["<rightAlt>"]
			original = original[len("<rightAltOn>"):]
			log.Printf("Special code '<rightAltOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightCtrlOn>") {
			keyCode = special["<rightCtrl>"]
			original = original[len("<rightCtrlOn>"):]
			log.Printf("Special code '<rightCtrlOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightShiftOn>") {
			keyCode = special["<rightShift>"]
			original = original[len("<rightShiftOn>"):]
			log.Printf("Special code '<rightShiftOn>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, true)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightAltOff>") {
			keyCode = special["<rightAlt>"]
			original = original[len("<rightAltOff>"):]
			log.Printf("Special code '<rightAltOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightCtrlOff>") {
			keyCode = special["<rightCtrl>"]
			original = original[len("<rightCtrlOff>"):]
			log.Printf("Special code '<rightCtrlOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<rightShiftOff>") {
			keyCode = special["<rightShift>"]
			original = original[len("<rightShiftOff>"):]
			log.Printf("Special code '<rightShiftOff>' found, replacing with: %d", keyCode)

			c.KeyEvent(keyCode, false)
			time.Sleep(keyInterval)

			continue
		}

		if strings.HasPrefix(original, "<wait>") {
			log.Printf("Special code '<wait>' found, sleeping one second")
			time.Sleep(1 * time.Second)
			original = original[len("<wait>"):]
			continue
		}

		if strings.HasPrefix(original, "<wait5>") {
			log.Printf("Special code '<wait5>' found, sleeping 5 seconds")
			time.Sleep(5 * time.Second)
			original = original[len("<wait5>"):]
			continue
		}

		if strings.HasPrefix(original, "<wait10>") {
			log.Printf("Special code '<wait10>' found, sleeping 10 seconds")
			time.Sleep(10 * time.Second)
			original = original[len("<wait10>"):]
			continue
		}

		for specialCode, specialValue := range special {
			if strings.HasPrefix(original, specialCode) {
				log.Printf("Special code '%s' found, replacing with: %d", specialCode, specialValue)
				keyCode = specialValue
				original = original[len(specialCode):]
				break
			}
		}

		if keyCode == 0 {
			r, size := utf8.DecodeRuneInString(original)
			original = original[size:]
			keyCode = uint32(r)
			keyShift = unicode.IsUpper(r) || strings.ContainsRune(shiftedChars, r)

			log.Printf("Sending char '%c', code %d, shift %v", r, keyCode, keyShift)
		}

		if keyShift {
			c.KeyEvent(KeyLeftShift, true)
		}

		c.KeyEvent(keyCode, true)
		time.Sleep(keyInterval)
		c.KeyEvent(keyCode, false)
		time.Sleep(keyInterval)

		if keyShift {
			c.KeyEvent(KeyLeftShift, false)
		}
	}
}
