package vmware

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
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
//   http_port int
//   ui     packer.Ui
//   vnc_port uint
//
// Produces:
//   <nothing>
type stepTypeBootCommand struct{}

func (s *stepTypeBootCommand) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	httpPort := state["http_port"].(uint)
	ui := state["ui"].(packer.Ui)
	vncPort := state["vnc_port"].(uint)

	// Connect to VNC
	ui.Say("Connecting to VM via VNC")
	nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	c, err := vnc.Client(nc, &vnc.ClientConfig{Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	// Determine the host IP
	ipFinder := &IfconfigIPFinder{"vmnet8"}
	hostIp, err := ipFinder.HostIP()
	if err != nil {
		err := fmt.Errorf("Error detecting host IP: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	tplData := &bootCommandTemplateData{
		hostIp,
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the boot command over VNC...")
	for _, command := range config.BootCommand {
		var buf bytes.Buffer
		t := template.Must(template.New("boot").Parse(command))
		t.Execute(&buf, tplData)

		vncSendString(c, buf.String())
	}

	return multistep.ActionContinue
}

func (*stepTypeBootCommand) Cleanup(map[string]interface{}) {}

func vncSendString(c *vnc.ClientConn, original string) {
	special := make(map[string]uint32)
	special["<enter>"] = 0xFF0D
	special["<esc>"] = 0xFF1B
	special["<return>"] = 0xFF0D
	special["<tab>"] = 0xFF09

	shiftedChars := "~!@#$%^&*()_+{}|:\"<>?"

	// TODO(mitchellh): Ripe for optimizations of some point, perhaps.
	for len(original) > 0 {
		var keyCode uint32
		keyShift := false

		if strings.HasPrefix(original, "<wait>") {
			log.Printf("Special code '<wait>' found, sleeping one second")
			time.Sleep(1 * time.Second)
			original = original[len("<wait>"):]
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
		c.KeyEvent(keyCode, false)

		if keyShift {
			c.KeyEvent(KeyLeftShift, false)
		}
	}
}
