package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestRunConfig_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *RunConfig
		expectedConfig *RunConfig
		driver         *DriverConfig
		errs           []error
		warnings       []string
	}{
		{
			name:   "VNC dafaults",
			config: &RunConfig{},
			expectedConfig: &RunConfig{
				VNCPortMin:     5900,
				VNCPortMax:     6000,
				VNCBindAddress: "127.0.0.1",
			},
			driver:   new(DriverConfig),
			errs:     nil,
			warnings: nil,
		},
		{
			name: "VNC port min less than vnc port max",
			config: &RunConfig{
				VNCPortMin: 5000,
				VNCPortMax: 5900,
			},
			expectedConfig: &RunConfig{
				VNCPortMin:     5000,
				VNCPortMax:     5900,
				VNCBindAddress: "127.0.0.1",
			},
			driver:   new(DriverConfig),
			errs:     nil,
			warnings: nil,
		},
		{
			name: "VNC port min bigger than vnc port max",
			config: &RunConfig{
				VNCPortMin: 5900,
				VNCPortMax: 5000,
			},
			expectedConfig: nil,
			driver:         new(DriverConfig),
			errs:           []error{fmt.Errorf("vnc_port_min must be less than vnc_port_max")},
			warnings:       nil,
		},
		{
			name: "VNC port min must be positive",
			config: &RunConfig{
				VNCPortMin: -1,
			},
			expectedConfig: nil,
			driver:         new(DriverConfig),
			errs:           []error{fmt.Errorf("vnc_port_min must be positive")},
			warnings:       nil,
		},
		{
			name: "fail when vnc_over_websocket set when remote_type is not set",
			config: &RunConfig{
				VNCOverWebsocket: true,
			},
			expectedConfig: nil,
			driver:         new(DriverConfig),
			errs:           []error{fmt.Errorf("'vnc_over_websocket' can only be used with remote VMWare builds.")},
			warnings:       nil,
		},
		{
			name: "warn about ignored vnc configuration",
			config: &RunConfig{
				VNCOverWebsocket: true,
				VNCPortMin:       5000,
				VNCPortMax:       5900,
			},
			expectedConfig: &RunConfig{
				VNCOverWebsocket: true,
				VNCPortMin:       5000,
				VNCPortMax:       5900,
			},
			driver: &DriverConfig{RemoteType: "esxi"},
			errs:   nil,
			warnings: []string{"[WARN] When 'vnc_over_websocket' is set " +
				"any other VNC configuration will be ignored."},
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
