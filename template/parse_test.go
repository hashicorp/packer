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
					"something": &Builder{
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
			"parse-builder-repeat.json",
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
					&Provisioner{
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
					&Provisioner{
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
					&Provisioner{
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
					&Provisioner{
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
					&Provisioner{
						Type: "something",
						Override: map[string]interface{}{
							"foo": map[string]interface{}{},
						},
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
			"parse-variable-default.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": &Variable{
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
					"foo": &Variable{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
							Type: "foo",
						},
					},
					[]*PostProcessor{
						&PostProcessor{
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
					[]*PostProcessor{
						&PostProcessor{
							Type: "foo",
						},
						&PostProcessor{
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
			"parse-description.json",
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
			"parse-push.json",
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
					"something": &Builder{
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		path, _ := filepath.Abs(fixtureDir(tc.File))
		tpl, err := ParseFile(fixtureDir(tc.File))
		if (err != nil) != tc.Err {
			t.Fatalf("err: %s", err)
		}

		if tc.Result != nil {
			tc.Result.Path = path
		}
		if tpl != nil {
			tpl.RawContents = nil
		}
		if !reflect.DeepEqual(tpl, tc.Result) {
			t.Fatalf("bad: %s\n\n%#v\n\n%#v", tc.File, tpl, tc.Result)
		}
	}
}

func TestParse_contents(t *testing.T) {
	tpl, err := ParseFile(fixtureDir("parse-contents.json"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(string(tpl.RawContents))
	expected := `{"builders":[{"type":"test"}]}`
	if actual != expected {
		t.Fatalf("bad: %s\n\n%s", actual, expected)
	}
}

func TestParse_bad(t *testing.T) {
	cases := []struct {
		File     string
		Expected string
	}{
		{"error-beginning.json", "line 1, column 1 (offset 1)"},
		{"error-middle.json", "line 5, column 6 (offset 50)"},
		{"error-end.json", "line 1, column 30 (offset 30)"},
	}
	for _, tc := range cases {
		_, err := ParseFile(fixtureDir(tc.File))
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), tc.Expected) {
			t.Fatalf("file: %s\nExpected: %s\n%s\n", tc.File, tc.Expected, err.Error())
		}
	}
}
