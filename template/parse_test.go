// +build !windows

package template

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		File   string
		Result *Template
		Err    bool
	}{
		/*
		 * Builders
		 */
		{
			"parse-basic.json",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-basic.hcl",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-builder-no-type.json",
			nil,
			true,
		},
		{
			"parse-builder-no-type.hcl",
			nil,
			true,
		},
		{
			"parse-builder-repeat.json",
			nil,
			true,
		},
		{
			"parse-builder-repeat.hcl",
			nil,
			true,
		},

		/*
		 * Provisioners
		 */
		{
			"parse-provisioner-basic.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-basic.hcl",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-pause-before.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type:        "something",
						PauseBefore: 1 * time.Second,
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-pause-before.hcl",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type:        "something",
						PauseBefore: 1 * time.Second,
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-only.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Only: []string{"foo"},
						},
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-only.hcl",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Only: []string{"foo"},
						},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-except.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Except: []string{"foo"},
						},
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-except.hcl",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Except: []string{"foo"},
						},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-override.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						Override: []map[string]interface{}{{
							"foo": []map[string]interface{}{{
								"bar": "baz",
							}},
						}},
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-override.hcl",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						Override: []map[string]interface{}{{
							"foo": []map[string]interface{}{{
								"bar": "baz",
							}},
						}},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-no-type.json",
			nil,
			true,
		},
		{
			"parse-provisioner-no-type.hcl",
			nil,
			true,
		},

		{
			"parse-variable-default.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Default: "foo",
					},
				},
			},
			false,
		},
		{
			"parse-variable-default.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Default: "foo",
					},
				},
			},
			false,
		},

		{
			"parse-variable-required.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Required: true,
					},
				},
			},
			false,
		},
		{
			"parse-variable-required.hcl",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Required: true,
					},
				},
			},
			false,
		},

		{
			"parse-pp-basic.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							Config: map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-basic.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							Config: map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-keep.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type:              "foo",
							KeepInputArtifact: true,
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-keep.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type:              "foo",
							KeepInputArtifact: true,
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-only.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Only: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-only.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Only: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-except.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Except: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-except.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Except: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-string.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-string.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-map.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-map.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-slice.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
					{
						{
							Type: "bar",
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-slice.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
					{
						{
							Type: "bar",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-multi.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
					{
						{
							Type: "bar",
						},
					},
				},
			},
			false,
		},
		{
			"parse-pp-multi.hcl",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Type: "foo",
						},
					},
					{
						{
							Type: "bar",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-no-type.json",
			nil,
			true,
		},
		{
			"parse-pp-no-type.hcl",
			nil,
			true,
		},

		{
			"parse-description.json",
			&Template{
				Description: "foo",
			},
			false,
		},
		{
			"parse-description.hcl",
			&Template{
				Description: "foo",
			},
			false,
		},

		{
			"parse-min-version.json",
			&Template{
				MinVersion: "1.2",
			},
			false,
		},
		{
			"parse-min-version.hcl",
			&Template{
				MinVersion: "1.2",
			},
			false,
		},

		{
			"parse-push.json",
			&Template{
				Push: Push{
					Name: "foo",
				},
			},
			false,
		},
		{
			"parse-push.hcl",
			&Template{
				Push: Push{
					Name: "foo",
				},
			},
			false,
		},

		{
			"parse-comment.json",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-comment.hcl",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
	}

	for i, tc := range cases {
		path, _ := filepath.Abs(fixtureDir(tc.File))
		tpl, err := ParseFile(fixtureDir(tc.File))
		if (err != nil) != tc.Err {
			t.Errorf("[%d]bad: %s. err: %s", i, tc.File, err)
			continue
		}

		if tc.Result != nil {
			tc.Result.Path = path
		}
		if tpl != nil {
			tpl.RawContents = nil
		}
		if !reflect.DeepEqual(tpl, tc.Result) {
			t.Errorf("[%d]bad: %s result,expected\n\n%#v\n\n%#v", i, tc.File, tpl, tc.Result)
		}
	}
}

func TestParse_contents(t *testing.T) {
	for _, f := range []string{"parse-contents.json", "parse-contents.hcl"} {
		tpl, err := ParseFile(fixtureDir(f))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := strings.TrimSpace(string(tpl.RawContents))
		expected := `{"builders":[{"type":"test"}]}`
		if actual != expected {
			t.Fatalf("bad: %s\n\n%s", actual, expected)
		}
	}
}

func TestParse_bad(t *testing.T) {
	cases := []struct {
		File     string
		Expected string
	}{
		{"error-beginning.json", "At 1:1: illegal char"},
		{"error-end.json", "1:30: illegal char: *"},
	}
	for _, tc := range cases {
		_, err := ParseFile(fixtureDir(tc.File))
		if err == nil {
			t.Errorf("file: %s\nexpected error", tc.File)
			continue
		}
		if !strings.Contains(err.Error(), tc.Expected) {
			t.Errorf("file: %s\nExpected: %s\n%s\n", tc.File, tc.Expected, err.Error())
		}
	}
}
