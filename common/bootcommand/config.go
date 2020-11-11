//go:generate struct-markdown

package bootcommand

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// The boot configuration is very important: `boot_command` specifies the keys
// to type when the virtual machine is first booted in order to start the OS
// installer. This command is typed after boot_wait, which gives the virtual
// machine some time to actually load.
//
// The boot_command is an array of strings. The strings are all typed in
// sequence. It is an array only to improve readability within the template.
//
// There are a set of special keys available. If these are in your boot
// command, they will be replaced by the proper key:
//
// -   `<bs>` - Backspace
//
// -   `<del>` - Delete
//
// -   `<enter> <return>` - Simulates an actual "enter" or "return" keypress.
//
// -   `<esc>` - Simulates pressing the escape key.
//
// -   `<tab>` - Simulates pressing the tab key.
//
// -   `<f1> - <f12>` - Simulates pressing a function key.
//
// -   `<up> <down> <left> <right>` - Simulates pressing an arrow key.
//
// -   `<spacebar>` - Simulates pressing the spacebar.
//
// -   `<insert>` - Simulates pressing the insert key.
//
// -   `<home> <end>` - Simulates pressing the home and end keys.
//
// -   `<pageUp> <pageDown>` - Simulates pressing the page up and page down
//     keys.
//
// -   `<menu>` - Simulates pressing the Menu key.
//
// -   `<leftAlt> <rightAlt>` - Simulates pressing the alt key.
//
// -   `<leftCtrl> <rightCtrl>` - Simulates pressing the ctrl key.
//
// -   `<leftShift> <rightShift>` - Simulates pressing the shift key.
//
// -   `<leftSuper> <rightSuper>` - Simulates pressing the ⌘ or Windows key.
//
// -   `<wait> <wait5> <wait10>` - Adds a 1, 5 or 10 second pause before
//     sending any additional keys. This is useful if you have to generally
//     wait for the UI to update before typing more.
//
// -   `<waitXX>` - Add an arbitrary pause before sending any additional keys.
//     The format of `XX` is a sequence of positive decimal numbers, each with
//     optional fraction and a unit suffix, such as `300ms`, `1.5h` or `2h45m`.
//     Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. For
//     example `<wait10m>` or `<wait1m20s>`.
//
// -   `<XXXOn> <XXXOff>` - Any printable keyboard character, and of these
//      "special" expressions, with the exception of the `<wait>` types, can
//      also be toggled on or off. For example, to simulate ctrl+c, use
//      `<leftCtrlOn>c<leftCtrlOff>`. Be sure to release them, otherwise they
//      will be held down until the machine reboots. To hold the `c` key down,
//      you would use `<cOn>`. Likewise, `<cOff>` to release.
//
// -   `{{ .HTTPIP }} {{ .HTTPPort }}` - The IP and port, respectively of an
//     HTTP server that is started serving the directory specified by the
//     `http_directory` configuration parameter. If `http_directory` isn't
//     specified, these will be blank!
//
// -   `{{ .Name }}` - The name of the VM.
//
// Example boot command. This is actually a working boot command used to start an
// CentOS 6.4 installer:
//
// In JSON:
//
// ```json
// "boot_command": [
//     "<tab><wait>",
//     " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
//  ]
// ```
//
// In HCL2:
//
// ```hcl
// boot_command = [
//     "<tab><wait>",
//     " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
//  ]
// ```
//
// The example shown below is a working boot command used to start an Ubuntu
// 12.04 installer:
//
// In JSON:
//
// ```json
// "boot_command": [
//   "<esc><esc><enter><wait>",
//   "/install/vmlinuz noapic ",
//   "preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
//   "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
//   "hostname={{ .Name }} ",
//   "fb=false debconf/frontend=noninteractive ",
//   "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
//   "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
//   "initrd=/install/initrd.gz -- <enter>"
// ]
// ```
//
// In HCL2:
//
// ```hcl
// boot_command = [
//   "<esc><esc><enter><wait>",
//   "/install/vmlinuz noapic ",
//   "preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
//   "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
//   "hostname={{ .Name }} ",
//   "fb=false debconf/frontend=noninteractive ",
//   "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
//   "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
//   "initrd=/install/initrd.gz -- <enter>"
// ]
// ```
//
// For more examples of various boot commands, see the sample projects from our
// [community templates page](/community-tools#templates).
type BootConfig struct {
	// Time to wait after sending a group of key pressses. The value of this
	// should be a duration. Examples are `5s` and `1m30s` which will cause
	// Packer to wait five seconds and one minute 30 seconds, respectively. If
	// this isn't specified, a sensible default value is picked depending on
	// the builder type.
	BootGroupInterval time.Duration `mapstructure:"boot_keygroup_interval"`
	// The time to wait after booting the initial virtual machine before typing
	// the `boot_command`. The value of this should be a duration. Examples are
	// `5s` and `1m30s` which will cause Packer to wait five seconds and one
	// minute 30 seconds, respectively. If this isn't specified, the default is
	// `10s` or 10 seconds. To set boot_wait to 0s, use a negative number, such
	// as "-1s"
	BootWait time.Duration `mapstructure:"boot_wait"`
	// This is an array of commands to type when the virtual machine is first
	// booted. The goal of these commands should be to type just enough to
	// initialize the operating system installer. Special keys can be typed as
	// well, and are covered in the section below on the boot command. If this
	// is not specified, it is assumed the installer will start itself.
	BootCommand []string `mapstructure:"boot_command"`
}

// The boot command "typed" character for character over a VNC connection to
// the machine, simulating a human actually typing the keyboard.
//
// Keystrokes are typed as separate key up/down events over VNC with a default
// 100ms delay. The delay alleviates issues with latency and CPU contention.
// You can tune this delay on a per-builder basis by specifying
// "boot_key_interval" in your Packer template.
type VNCConfig struct {
	BootConfig `mapstructure:",squash"`
	// Whether to create a VNC connection or not. A boot_command cannot be used
	// when this is true. Defaults to false.
	DisableVNC bool `mapstructure:"disable_vnc"`
	// Time in ms to wait between each key press
	BootKeyInterval time.Duration `mapstructure:"boot_key_interval"`
}

func (c *BootConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if c.BootWait == 0 {
		c.BootWait = 10 * time.Second
	}

	if c.BootCommand != nil {
		expSeq, err := GenerateExpressionSequence(c.FlatBootCommand())
		if err != nil {
			errs = append(errs, err)
		} else if vErrs := expSeq.Validate(); vErrs != nil {
			errs = append(errs, vErrs...)
		}
	}

	return
}

func (c *BootConfig) FlatBootCommand() string {
	return strings.Join(c.BootCommand, "")
}

func (c *VNCConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if len(c.BootCommand) > 0 && c.DisableVNC {
		errs = append(errs,
			fmt.Errorf("A boot command cannot be used when vnc is disabled."))
	}

	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}
