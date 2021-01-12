package proxmox

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type commandTyperMock struct {
	sendkey func(*proxmox.VmRef, string) error
}

func (m commandTyperMock) Sendkey(ref *proxmox.VmRef, cmd string) error {
	return m.sendkey(ref, cmd)
}

var _ commandTyper = commandTyperMock{}

func TestTypeBootCommand(t *testing.T) {
	cs := []struct {
		name              string
		builderConfig     *Config
		expectCallSendkey bool
		sendkeyErr        error
		expectedKeysSent  string
		expectedAction    multistep.StepAction
	}{
		{
			name:              "simple boot command is typed",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "hello",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "interpolated boot command",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello<enter>world"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "helloretworld",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "merge multiple interpolated boot command",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"Hello World 2.0", "foo!bar@baz"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "shift-hellospcshift-worldspc2dot0fooshift-1barshift-2baz",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "holding and releasing keys",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"<leftShiftOn>hello<rightAltOn>world<leftShiftOff><rightAltOff>"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "shift-hshift-eshift-lshift-lshift-oshift-alt_r-wshift-alt_r-oshift-alt_r-rshift-alt_r-lshift-alt_r-d",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "holding multiple alphabetical keys and shift",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"<cOn><leftShiftOn>n<leftShiftOff><cOff>"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "shift-c-n",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "noop keystrokes",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"<cOn><leftShiftOn><cOff><leftAltOn><leftShiftOff><leftAltOff>"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "noop keystrokes mixed",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"<cOn><leftShiftOn><cOff>h<leftShiftOff>"}}},
			expectCallSendkey: true,
			expectedKeysSent:  "shift-h",
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "without boot command sendkey should not be called",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{}}},
			expectCallSendkey: false,
			expectedAction:    multistep.ActionContinue,
		},
		{
			name:              "invalid boot command template function",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"{{ foo }}"}}},
			expectCallSendkey: false,
			expectedAction:    multistep.ActionHalt,
		},
		{
			// When proxmox (or Qemu, really) doesn't recognize the keycode we send, we get no error back, but
			// a map {"data": "invalid parameter: X"}, where X is the keycode.
			name:              "invalid keys sent to proxmox",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"x"}}},
			expectCallSendkey: true,
			sendkeyErr:        fmt.Errorf("invalid parameter: x"),
			expectedKeysSent:  "x",
			expectedAction:    multistep.ActionHalt,
		},
		{
			name:              "error in typing should return halt",
			builderConfig:     &Config{BootConfig: bootcommand.BootConfig{BootCommand: []string{"hello"}}},
			expectCallSendkey: true,
			sendkeyErr:        fmt.Errorf("some error"),
			expectedKeysSent:  "h",
			expectedAction:    multistep.ActionHalt,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			accumulator := strings.Builder{}
			typer := commandTyperMock{
				sendkey: func(ref *proxmox.VmRef, cmd string) error {
					if !c.expectCallSendkey {
						t.Error("Did not expect sendkey to be called")
					}

					accumulator.WriteString(cmd)

					return c.sendkeyErr
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packersdk.TestUi(t))
			state.Put("config", c.builderConfig)
			state.Put("http_port", int(0))
			state.Put("vmRef", proxmox.NewVmRef(1))
			state.Put("proxmoxClient", typer)

			step := stepTypeBootCommand{
				c.builderConfig.BootConfig,
				c.builderConfig.Ctx,
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
