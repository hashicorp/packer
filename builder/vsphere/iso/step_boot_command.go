package iso

import (
	"context"
	"fmt"
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"golang.org/x/mobile/event/key"
	"log"
	"net"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type BootConfig struct {
	BootCommand []string      `mapstructure:"boot_command"`
	BootWait    time.Duration `mapstructure:"boot_wait"` // example: "1m30s"; default: "10s"
	HTTPIP      string        `mapstructure:"http_ip"`
}

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

func (c *BootConfig) Prepare() []error {
	var errs []error

	if c.BootWait == 0 {
		c.BootWait = 10 * time.Second
	}

	return errs
}

type StepBootCommand struct {
	Config *BootConfig
	VMName string
	Ctx    interpolate.Context
}

var special = map[string]key.Code{
	"<enter>":    key.CodeReturnEnter,
	"<esc>":      key.CodeEscape,
	"<bs>":       key.CodeDeleteBackspace,
	"<del>":      key.CodeDeleteForward,
	"<tab>":      key.CodeTab,
	"<f1>":       key.CodeF1,
	"<f2>":       key.CodeF2,
	"<f3>":       key.CodeF3,
	"<f4>":       key.CodeF4,
	"<f5>":       key.CodeF5,
	"<f6>":       key.CodeF6,
	"<f7>":       key.CodeF7,
	"<f8>":       key.CodeF8,
	"<f9>":       key.CodeF9,
	"<f10>":      key.CodeF10,
	"<f11>":      key.CodeF11,
	"<f12>":      key.CodeF12,
	"<insert>":   key.CodeInsert,
	"<home>":     key.CodeHome,
	"<end>":      key.CodeEnd,
	"<pageUp>":   key.CodePageUp,
	"<pageDown>": key.CodePageDown,
	"<left>":     key.CodeLeftArrow,
	"<right>":    key.CodeRightArrow,
	"<up>":       key.CodeUpArrow,
	"<down>":     key.CodeDownArrow,
}

var keyInterval = packerCommon.PackerKeyDefault

func init() {
	if delay, err := time.ParseDuration(os.Getenv(packerCommon.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}
}

func (s *StepBootCommand) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if s.Config.BootCommand == nil {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Waiting %s for boot...", s.Config.BootWait))
	wait := time.After(s.Config.BootWait)
WAITLOOP:
	for {
		select {
		case <-wait:
			break WAITLOOP
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}
		}
	}

	port := state.Get("http_port").(int)
	if port > 0 {
		ip, err := getHostIP(s.Config.HTTPIP)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		err = packerCommon.SetHTTPIP(ip)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.Ctx.Data = &bootCommandTemplateData{
			ip,
			port,
			s.VMName,
		}
		ui.Say(fmt.Sprintf("HTTP server is working at http://%v:%v/", ip, port))
	}

	ui.Say("Typing boot command...")
	var keyAlt bool
	var keyCtrl bool
	var keyShift bool
	for _, command := range s.Config.BootCommand {
		message, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		for len(message) > 0 {
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}

			if strings.HasPrefix(message, "<wait>") {
				log.Printf("Waiting 1 second")
				time.Sleep(1 * time.Second)
				message = message[len("<wait>"):]
				continue
			}

			if strings.HasPrefix(message, "<wait5>") {
				log.Printf("Waiting 5 seconds")
				time.Sleep(5 * time.Second)
				message = message[len("<wait5>"):]
				continue
			}

			if strings.HasPrefix(message, "<wait10>") {
				log.Printf("Waiting 10 seconds")
				time.Sleep(10 * time.Second)
				message = message[len("<wait10>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftAltOn>") {
				keyAlt = true
				message = message[len("<leftAltOn>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftAltOff>") {
				keyAlt = false
				message = message[len("<leftAltOff>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftCtrlOn>") {
				keyCtrl = true
				message = message[len("<leftCtrlOn>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftCtrlOff>") {
				keyCtrl = false
				message = message[len("<leftCtrlOff>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftShiftOn>") {
				keyShift = true
				message = message[len("<leftShiftOn>"):]
				continue
			}

			if strings.HasPrefix(message, "<leftShiftOff>") {
				keyShift = false
				message = message[len("<leftShiftOff>"):]
				continue
			}

			var scancode key.Code
			for specialCode, specialValue := range special {
				if strings.HasPrefix(message, specialCode) {
					scancode = specialValue
					log.Printf("Special code '%s' found, replacing with: %s", specialCode, specialValue)
					message = message[len(specialCode):]
				}
			}

			var char rune
			if scancode == 0 {
				var size int
				char, size = utf8.DecodeRuneInString(message)
				message = message[size:]
			}

			_, err := vm.TypeOnKeyboard(driver.KeyInput{
				Message:  string(char),
				Scancode: scancode,
				Ctrl:     keyCtrl,
				Alt:      keyAlt,
				Shift:    keyShift,
			})
			if err != nil {
				state.Put("error", fmt.Errorf("error typing a boot command: %v", err))
				return multistep.ActionHalt
			}
			time.Sleep(keyInterval)
		}
	}

	return multistep.ActionContinue
}

func (s *StepBootCommand) Cleanup(state multistep.StateBag) {}

func getHostIP(s string) (string, error) {
	if s != "" {
		if net.ParseIP(s) != nil {
			return s, nil
		} else {
			return "", fmt.Errorf("invalid IP address")
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("IP not found")
}
