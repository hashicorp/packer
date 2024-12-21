package hcp_sbom

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestConfigPrepare(t *testing.T) {
	tests := []struct {
		name               string
		inputConfig        map[string]interface{}
		interpolateContext interpolate.Context
		expectConfig       *Config
		expectError        bool
	}{
		{
			"empty config, should error without a source",
			map[string]interface{}{},
			interpolate.Context{},
			nil,
			true,
		},
		{
			"config with full context for interpolation: success",
			map[string]interface{}{
				"source": "{{ .Name }}",
			},
			interpolate.Context{
				Data: &struct {
					Name string
				}{
					Name: "testInterpolate",
				},
			},
			&Config{
				Source: "testInterpolate",
			},
			false,
		},
		{
			// Note: this will look weird to reviewers, but is actually
			// expected for the moment.
			// Refer to the comment in `Prepare` for context as to WHY
			// this cannot be considered an error.
			"config with sbom name as interpolated value, without it in context, replace with a placeholder",
			map[string]interface{}{
				"source":    "test",
				"sbom_name": "{{ .Name }}",
			},
			interpolate.Context{},
			&Config{
				Source:   "test",
				SbomName: "<no value>",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov := &Provisioner{}
			prov.config.ctx = tt.interpolateContext
			err := prov.Prepare(tt.inputConfig)
			if err != nil && !tt.expectError {
				t.Fatalf("configuration unexpectedly failed to prepare: %s", err)
			}

			if err == nil && tt.expectError {
				t.Fatalf("configuration succeeded to prepare, but should have failed")
			}

			if err != nil {
				t.Logf("config had error %q", err)
				return
			}

			diff := cmp.Diff(prov.config, *tt.expectConfig, cmpopts.IgnoreUnexported(Config{}))
			if diff != "" {
				t.Errorf("configuration returned by `Prepare` is different from what was expected: %s", diff)
			}
		})
	}
}
