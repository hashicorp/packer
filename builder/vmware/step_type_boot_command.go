package vmware

import (
	"fmt"
	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const KeyLeftShift uint32 = 0xFFE1

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   vnc_port uint
//
// Produces:
//   <nothing>
type stepTypeBootCommand struct{}

func (s *stepTypeBootCommand) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	vncPort := state["vnc_port"].(uint)

	ui.Say("Connecting to VM via VNC")
	nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", vncPort))
	if err != nil {
		ui.Error(fmt.Sprintf("Error connecting to VNC: %s", err))
		return multistep.ActionHalt
	}
	defer nc.Close()

	c, err := vnc.Client(nc, &vnc.ClientConfig{Exclusive: true})
	if err != nil {
		ui.Error(fmt.Sprintf("Error handshaking with VNC: %s", err))
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connecting to VNC desktop: %s", c.DesktopName)
	ui.Say("Typing the boot command over VNC...")
	for _, command := range config.BootCommand {
		vncSendString(c, command)
	}

	return multistep.ActionContinue
}

func (*stepTypeBootCommand) Cleanup(map[string]interface{}) {}

func vncSendString(c *vnc.ClientConn, original string) {
	special := make(map[string]uint32)
	special["<enter>"] = 0xFF0D
	special["<return>"] = 0xFF0D
	special["<esc>"] = 0xFF1B

	// TODO(mitchellh): Ripe for optimizations of some point, perhaps.
	for len(original) > 0 {
		var keyCode uint32
		keyShift := false

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
			keyShift = unicode.IsUpper(r)

			log.Printf("Sending char '%c', code %d, shift %v", r, keyCode, keyShift)
		}

		if keyShift {
			c.KeyEvent(KeyLeftShift, true)
		}

		c.KeyEvent(keyCode, true)
		c.KeyEvent(keyCode, false)

		if keyShift {
			c.KeyEvent(KeyLeftShift, false)
		}
	}
}
