package qemu

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/mitchellh/go-vnc"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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

type keyComm interface{}

type qmpSender struct {
	c *bufio.ReadWriter
}

type vncSender struct {
	c *vnc.ClientConn
}

func (s *stepTypeBootCommand) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	var comm keyComm
	port := uint(0)

	switch config.SendKeyComm {
	case "vnc":
		port = state.Get("vnc_port").(uint)
	case "qmp":
		port = state.Get("qmp_port").(uint)
	}

	ui.Say("Connecting to VM via " + config.SendKeyComm)
	nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		err := fmt.Errorf("Error connecting to %s: %s", config.SendKeyComm, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()
	switch config.SendKeyComm {
	case "vnc":
		c, err := vnc.Client(nc, &vnc.ClientConfig{Exclusive: false})
		if err != nil {
			err := fmt.Errorf("Error handshaking with VNC: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		defer c.Close()
		comm = &vncSender{c: c}
		log.Printf("Connected to VNC desktop: %s", c.DesktopName)
	case "qmp":
		var banner []byte
		comm = &qmpSender{c: bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc))}
		n, err := comm.(*qmpSender).c.Read(banner)
		if err != nil {
			log.Printf("qemu qmp banner err: %s\n", err.Error())
			return multistep.ActionHalt
		}
		log.Printf("Connected to QMP monitor: 127.0.0.1:%d", port)
		log.Printf("qemu qmp banner: %s\n", banner[:n])
	}

	ctx := config.ctx
	ctx.Data = &bootCommandTemplateData{
		"10.0.2.2",
		httpPort,
		config.VMName,
	}

	ui.Say("Typing the boot command...")
	for _, command := range config.BootCommand {
		command, err := interpolate.Render(command, &ctx)
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

		sendString(comm, command)
	}

	return multistep.ActionContinue
}

func (*stepTypeBootCommand) Cleanup(multistep.StateBag) {}

func sendString(c keyComm, original string) {
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

		if strings.HasPrefix(original, "<wait") && strings.HasSuffix(original, ">") {
			re := regexp.MustCompile(`<wait([0-9hms]+)>$`)
			dstr := re.FindStringSubmatch(original)
			if len(dstr) > 1 {
				log.Printf("Special code %s found, sleeping", dstr[0])
				if dt, err := time.ParseDuration(dstr[1]); err == nil {
					time.Sleep(dt)
					original = original[len(dstr[0]):]
					continue
				}
			}
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

		switch c.(type) {
		case *vncSender:
			keyc := c.(*vncSender)
			if keyShift {
				keyc.c.KeyEvent(KeyLeftShift, true)
			}
			keyc.c.KeyEvent(keyCode, true)
			time.Sleep(time.Second / 10)
			keyc.c.KeyEvent(keyCode, false)
			time.Sleep(time.Second / 10)
			if keyShift {
				keyc.c.KeyEvent(KeyLeftShift, false)
			}
			// qemu is picky, so no matter what, wait a small period
			time.Sleep(100 * time.Millisecond)
		case *qmpSender:
			keyc := c.(*qmpSender)
			req := struct {
				Execute   string `json:"execute"`
				Arguments struct {
					Keys     []uint32 `json:"keys"`
					HoldTime uint     `json:"hold-time,omitempty"`
				} `json:"arguments"`
			}{Execute: "send-key"} //, Arguments: {HoldTime: 100}}
			if keyShift {
				req.Arguments.Keys = append(req.Arguments.Keys, KeyLeftShift)
			}
			req.Arguments.Keys = append(req.Arguments.Keys, keyCode)
			buf, _ := json.Marshal(req)
			buf = append(buf, byte('\n'))
			log.Printf("%s", buf)
			keyc.c.Write(buf)
			keyc.c.Flush()
		}
	}
}
