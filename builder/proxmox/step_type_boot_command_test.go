package proxmox

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type commandTyperMock struct {
	monitorCmd func(*proxmox.VmRef, string) (map[string]interface{}, error)
}

func (m commandTyperMock) MonitorCmd(ref *proxmox.VmRef, cmd string) (map[string]interface{}, error) {
	return m.monitorCmd(ref, cmd)
}

var _ commandTyper = commandTyperMock{}

func TestTypeBootCommand(t *testing.T) {
	cs := []struct {
		name                 string
		builderConfig        *Config
		expectCallMonitorCmd bool
		monitorCmdErr        error
		monitorCmdRet        map[string]interface{}
		expectedKeysSent     string
		expectedAction       multistep.StepAction
	}{
		{
			name:                 "simple boot command is typed",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello"}}},
			expectCallMonitorCmd: true,
			expectedKeysSent:     "hello",
			expectedAction:       multistep.ActionContinue,
		},
		{
			name:                 "interpolated boot command",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello<enter>world"}}},
			expectCallMonitorCmd: true,
			expectedKeysSent:     "helloretworld",
			expectedAction:       multistep.ActionContinue,
		},
		{
			name:                 "merge multiple interpolated boot command",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"Hello World 2.0", "foo!bar@baz"}}},
			expectCallMonitorCmd: true,
			expectedKeysSent:     "shift-hellospcshift-worldspc2dot0fooshift-1barshift-2baz",
			expectedAction:       multistep.ActionContinue,
		},
		{
			name:                 "without boot command monitorcmd should not be called",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{}}},
			expectCallMonitorCmd: false,
			expectedAction:       multistep.ActionContinue,
		},
		{
			name:                 "invalid boot command template function",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"{{ foo }}"}}},
			expectCallMonitorCmd: false,
			expectedAction:       multistep.ActionHalt,
		},
		{
			// When proxmox (or Qemu, really) doesn't recognize the keycode we send, we get no error back, but
			// a map {"data": "invalid parameter: X"}, where X is the keycode.
			name:                 "invalid keys sent to proxmox",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"x"}}},
			expectCallMonitorCmd: true,
			monitorCmdRet:        map[string]interface{}{"data": "invalid parameter: x"},
			expectedKeysSent:     "x",
			expectedAction:       multistep.ActionHalt,
		},
		{
			name:                 "error in typing should return halt",
			builderConfig:        &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello"}}},
			expectCallMonitorCmd: true,
			monitorCmdErr:        fmt.Errorf("some error"),
			expectedKeysSent:     "h",
			expectedAction:       multistep.ActionHalt,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			accumulator := strings.Builder{}
			typer := commandTyperMock{
				monitorCmd: func(ref *proxmox.VmRef, cmd string) (map[string]interface{}, error) {
					if !c.expectCallMonitorCmd {
						t.Error("Did not expect MonitorCmd to be called")
					}
					if !strings.HasPrefix(cmd, "sendkey ") {
						t.Errorf("Expected all commands to be sendkey, got %s", cmd)
					}

					accumulator.WriteString(strings.TrimPrefix(cmd, "sendkey "))

					return c.monitorCmdRet, c.monitorCmdErr
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packer.TestUi(t))
			state.Put("config", c.builderConfig)
			state.Put("http_port", int(0))
			state.Put("vmRef", proxmox.NewVmRef(1))
			state.Put("proxmoxClient", typer)

			step := stepTypeBootCommand{
				c.builderConfig.BootConfig,
				c.builderConfig.ctx,
			}
			action := step.Run(context.TODO(), state)
			step.Cleanup(state)

			if action != c.expectedAction {
				t.Errorf("Expected action to be %v, got %v", c.expectedAction, action)
			}
			if c.expectedKeysSent != accumulator.String() {
				t.Errorf("Expected keystrokes to be %q, got %q", c.expectedKeysSent, accumulator.String())
			}
		})
	}
}
