package config

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

func TestDecode(t *testing.T) {
	type Target struct {
		Name    string
		Address string
		Time    time.Duration
		Trilean Trilean
	}

	cases := map[string]struct {
		Input  []interface{}
		Output *Target
		Opts   *DecodeOpts
	}{
		"basic": {
			[]interface{}{
				map[string]interface{}{
					"name":    "bar",
					"time":    "5s",
					"trilean": "true",
				},
			},
			&Target{
				Name:    "bar",
				Time:    5 * time.Second,
				Trilean: TriTrue,
			},
			nil,
		},

		"empty-string-trilean": {
			[]interface{}{
				map[string]interface{}{
					"trilean": "",
				},
			},
			&Target{
				Trilean: TriUnset,
			},
			nil,
		},

		"variables": {
			[]interface{}{
				map[string]interface{}{
					"name": "{{user `name`}}",
				},
				map[string]interface{}{
					"packer_user_variables": map[string]string{
						"name": "bar",
					},
				},
			},
			&Target{
				Name: "bar",
			},
			nil,
		},

		"filter": {
			[]interface{}{
				map[string]interface{}{
					"name":    "{{user `name`}}",
					"address": "{{user `name`}}",
				},
				map[string]interface{}{
					"packer_user_variables": map[string]string{
						"name": "bar",
					},
				},
			},
			&Target{
				Name:    "bar",
				Address: "{{user `name`}}",
			},
			&DecodeOpts{
				Interpolate: true,
				InterpolateFilter: &interpolate.RenderFilter{
					Include: []string{"name"},
				},
			},
		},

		"build name": {
			[]interface{}{
				map[string]interface{}{
					"name": "{{build_name}}",
				},
				map[string]interface{}{
					"packer_build_name": "foo",
				},
			},
			&Target{
				Name: "foo",
			},
			nil,
		},

		"build type": {
			[]interface{}{
				map[string]interface{}{
					"name": "{{build_type}}",
				},
				map[string]interface{}{
					"packer_builder_type": "foo",
				},
			},
			&Target{
				Name: "foo",
			},
			nil,
		},
	}

	for k, tc := range cases {
		var result Target
		err := Decode(&result, tc.Opts, tc.Input...)
		if err != nil {
			t.Fatalf("err: %s\n\n%s", k, err)
		}

		if !reflect.DeepEqual(&result, tc.Output) {
			t.Fatalf("bad:\n\n%#v\n\n%#v", &result, tc.Output)
		}
	}
}

func TestDecode_fixerRecommendations(t *testing.T) {
	type TestConfig struct {
		Name string
	}

	cases := []struct {
		Reason   string
		Input    []interface{}
		Opts     *DecodeOpts
		Expected string
	}{
		{
			Reason: "If no plugin type is provided, don't try to match fixer options",
			Input: []interface{}{
				map[string]interface{}{
					"name":    "bar",
					"iso_md5": "13123412341234",
				},
			},
			Opts:     &DecodeOpts{},
			Expected: `unknown configuration key: '"iso_md5"'`,
		},
		{
			Reason: "iso_md5 should always recommend packer fix regardless of plugin type",
			Input: []interface{}{
				map[string]interface{}{
					"name":    "bar",
					"iso_md5": "13123412341234",
				},
			},
			Opts:     &DecodeOpts{PluginType: "someplugin"},
			Expected: `Deprecated configuration key: 'iso_md5'`,
		},
		{
			Reason: "filename option should generate a fixer recommendation for the manifest postprocessor",
			Input: []interface{}{
				map[string]interface{}{
					"name":     "bar",
					"filename": "fakefilename",
				},
			},
			Opts:     &DecodeOpts{PluginType: "packer.post-processor.manifest"},
			Expected: `Deprecated configuration key: 'filename'`,
		},
		{
			Reason: "filename option should generate an unknown key error for other plugins",
			Input: []interface{}{
				map[string]interface{}{
					"name":     "bar",
					"filename": "fakefilename",
				},
			},
			Opts:     &DecodeOpts{PluginType: "randomplugin"},
			Expected: `unknown configuration key: '"filename"'`,
		},
	}

	for _, tc := range cases {
		var result TestConfig
		err := Decode(&result, tc.Opts, tc.Input...)
		if err == nil {
			t.Fatalf("Should have had an error: %s", tc.Reason)
		}

		if !strings.Contains(err.Error(), tc.Expected) {
			t.Fatalf("Expected: %s\nActual: %s\n; Reason: %s", tc.Expected, err.Error(), tc.Reason)
		}
	}
}
