package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/template/interpolate"
)

func TestVNCConfigWrapper_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *BootConfigWrapper
		expectedConfig *BootConfigWrapper
		driver         *DriverConfig
		errs           []error
		warnings       []string
	}{
		{
			name: "VNC and boot command for local build",
			config: &BootConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					BootConfig: bootcommand.BootConfig{
						BootCommand: []string{"<boot><command>"},
					},
					DisableVNC: true,
				},
				USBScanCode: false,
			},
			expectedConfig: nil,
			driver:         new(DriverConfig),
			errs:           []error{fmt.Errorf("A boot command cannot be used when vnc is disabled.")},
			warnings:       nil,
		},
		{
			name: "Disable VNC warning for remote build",
			config: &BootConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					BootConfig: bootcommand.BootConfig{
						BootCommand: []string{"<boot><command>"},
					},
					DisableVNC: false,
				},
				USBScanCode: true,
			},
			expectedConfig: &BootConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					DisableVNC: true,
				},
				USBScanCode: true,
			},
			driver: &DriverConfig{
				RemoteType: "esxi",
			},
			errs: nil,
			warnings: []string{"[WARN] `usb_scan_codes` is set to true then the remote VMWare builds " +
				"will not use VNC to connect to the host. The `disable_vnc` option will be ignored and automatically set to true."},
		},
		{
			name: "Disable USBScanCode warning for local build",
			config: &BootConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					BootConfig: bootcommand.BootConfig{
						BootCommand: []string{"<boot><command>"},
					},
					DisableVNC: false,
				},
				USBScanCode: true,
			},
			expectedConfig: &BootConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					DisableVNC: false,
				},
				USBScanCode: false,
			},
			driver: &DriverConfig{},
			errs:   nil,
			warnings: []string{"[WARN] `usb_scan_codes` can only be used with remote VMWare builds. " +
				"The `usb_scan_codes` option will be ignored and automatically set to false."},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			warnings, errs := c.config.Prepare(interpolate.NewContext(), c.driver)
			if !reflect.DeepEqual(errs, c.errs) {
				t.Fatalf("bad: \n expected '%v' \nactual '%v'", c.errs, errs)
			}
			if diff := cmp.Diff(warnings, c.warnings); diff != "" {
				t.Fatalf("unexpected warnings: %s", diff)
			}
			if len(c.errs) == 0 {
				if diff := cmp.Diff(c.config, c.expectedConfig,
					cmpopts.IgnoreFields(bootcommand.VNCConfig{},
						"BootConfig",
					)); diff != "" {
					t.Fatalf("unexpected config: %s", diff)
				}
			}
		})
	}
}
