package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestDriverConfigPrepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *DriverConfig
		expectedConfig *DriverConfig
		errs           []error
	}{
		{
			name:   "Set default values",
			config: new(DriverConfig),
			expectedConfig: &DriverConfig{
				FusionAppPath:        "/Applications/VMware Fusion.app",
				RemoteDatastore:      "datastore1",
				RemoteCacheDatastore: "datastore1",
				RemoteCacheDirectory: "packer_cache",
				RemotePort:           22,
				RemoteUser:           "root",
			},
			errs: nil,
		},
		{
			name: "Override default values",
			config: &DriverConfig{
				FusionAppPath:        "foo",
				RemoteDatastore:      "set-datastore1",
				RemoteCacheDatastore: "set-datastore1",
				RemoteCacheDirectory: "set_packer_cache",
				RemotePort:           443,
				RemoteUser:           "admin",
			},
			expectedConfig: &DriverConfig{
				FusionAppPath:        "foo",
				RemoteDatastore:      "set-datastore1",
				RemoteCacheDatastore: "set-datastore1",
				RemoteCacheDirectory: "set_packer_cache",
				RemotePort:           443,
				RemoteUser:           "admin",
			},
			errs: nil,
		},
		{
			name: "Invalid remote type",
			config: &DriverConfig{
				RemoteType: "invalid",
				RemoteHost: "host",
			},
			expectedConfig: nil,
			errs:           []error{fmt.Errorf("Only 'esx5' value is accepted for remote_type")},
		},
		{
			name: "Remote host not set",
			config: &DriverConfig{
				RemoteType: "esx5",
			},
			expectedConfig: nil,
			errs:           []error{fmt.Errorf("remote_host must be specified")},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			errs := c.config.Prepare(interpolate.NewContext())
			if !reflect.DeepEqual(errs, c.errs) {
				t.Fatalf("bad: \n expected '%v' \nactual '%v'", c.errs, errs)
			}
			if len(c.errs) == 0 {
				if diff := cmp.Diff(c.config, c.expectedConfig); diff != "" {
					t.Fatalf("bad value: %s", diff)
				}
			}
		})
	}
}
