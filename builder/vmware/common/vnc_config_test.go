package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/template/interpolate"
)

func TestVNCConfigWrapper_Prepare(t *testing.T) {
	tc := []struct {
		name     string
		config   *VNCConfigWrapper
		driver   *DriverConfig
		errs     []error
		warnings []string
	}{
		{
			name: "VNC and boot command for local build",
			config: &VNCConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					BootConfig: bootcommand.BootConfig{
						BootCommand: []string{"<boot><command>"},
					},
					DisableVNC: true,
				},
			},
			driver:   new(DriverConfig),
			errs:     []error{fmt.Errorf("A boot command cannot be used when vnc is disabled.")},
			warnings: nil,
		},
		{
			name: "Disable VNC warning for remote build",
			config: &VNCConfigWrapper{
				VNCConfig: bootcommand.VNCConfig{
					BootConfig: bootcommand.BootConfig{
						BootCommand: []string{"<boot><command>"},
					},
					DisableVNC: false,
				},
			},
			driver: &DriverConfig{
				RemoteType: "esxi",
			},
			errs:     nil,
			warnings: []string{"[WARN] The vmware-esxi do not use VNC to connect to the host anymore. By default, the VNC is disabled."},
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
		})
	}
}
