package interpolate

import (
	"reflect"
	"testing"
)

func TestRenderInterface(t *testing.T) {
	type Test struct {
		Foo string
	}

	cases := map[string]struct {
		Input  interface{}
		Output interface{}
	}{
		"basic": {
			map[string]interface{}{
				"foo": "{{upper `bar`}}",
			},
			map[string]interface{}{
				"foo": "BAR",
			},
		},

		"struct": {
			&Test{
				Foo: "{{upper `bar`}}",
			},
			&Test{
				Foo: "BAR",
			},
		},
	}

	ctx := &Context{}
	for k, tc := range cases {
		actual, err := RenderInterface(tc.Input, ctx)
		if err != nil {
			t.Fatalf("err: %s\n\n%s", k, err)
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("err: %s\n\n%#v\n\n%#v", k, actual, tc.Output)
		}
	}
}

func TestRenderMap(t *testing.T) {
	cases := map[string]struct {
		Input  interface{}
		Output interface{}
		Filter *RenderFilter
	}{
		"basic": {
			map[string]interface{}{
				"foo": "{{upper `bar`}}",
			},
			map[string]interface{}{
				"foo": "BAR",
			},
			nil,
		},

		"map keys shouldn't be interpolated": {
			map[string]interface{}{
				"{{foo}}": "{{upper `bar`}}",
			},
			map[string]interface{}{
				"{{foo}}": "BAR",
			},
			nil,
		},

		"nested values": {
			map[string]interface{}{
				"foo": map[string]string{
					"bar": "{{upper `baz`}}",
				},
			},
			map[string]interface{}{
				"foo": map[string]string{
					"bar": "BAZ",
				},
			},
			nil,
		},

		// this test fails if you get github.com/mitchellh/reflectwalk@v1.0.1
		// the fail is caused by
		// https://github.com/mitchellh/reflectwalk/pull/22/commits/51d4c99fad9e9aa269e874bc3ad60313a574799f
		// TODO: open a PR to fix it.
		"nested value keys": {
			map[string]interface{}{
				"foo": map[string]string{
					"{{upper `bar`}}": "{{upper `baz`}}",
				},
			},
			map[string]interface{}{
				"foo": map[string]string{
					"BAR": "BAZ",
				},
			},
			nil,
		},

		"filter": {
			map[string]interface{}{
				"bar": "{{upper `baz`}}",
				"foo": map[string]string{
					"{{upper `bar`}}": "{{upper `baz`}}",
				},
			},
			map[string]interface{}{
				"bar": "BAZ",
				"foo": map[string]string{
					"{{upper `bar`}}": "{{upper `baz`}}",
				},
			},
			&RenderFilter{
				Include: []string{"bar"},
			},
		},

		"filter case-insensitive": {
			map[string]interface{}{
				"bar": "{{upper `baz`}}",
				"foo": map[string]string{
					"{{upper `bar`}}": "{{upper `baz`}}",
				},
			},
			map[string]interface{}{
				"bar": "BAZ",
				"foo": map[string]string{
					"{{upper `bar`}}": "{{upper `baz`}}",
				},
			},
			&RenderFilter{
				Include: []string{"baR"},
			},
		},
	}

	ctx := &Context{}
	for k, tc := range cases {
		actual, err := RenderMap(tc.Input, ctx, tc.Filter)
		if err != nil {
			t.Fatalf("err: %s\n\n%s", k, err)
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("err: %s\n\n%#v\n\n%#v", k, actual, tc.Output)
		}
	}
}
