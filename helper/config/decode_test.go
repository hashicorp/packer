package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

func TestDecode(t *testing.T) {
	type Target struct {
		Name    string
		Address string
		Time    time.Duration
	}

	cases := map[string]struct {
		Input  []interface{}
		Output *Target
		Opts   *DecodeOpts
	}{
		"basic": {
			[]interface{}{
				map[string]interface{}{
					"name": "bar",
					"time": "5s",
				},
			},
			&Target{
				Name: "bar",
				Time: 5 * time.Second,
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
