package hcl2template

import (
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestParse_datasource(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"two basic datasources",
			defaultParser,
			parseTestArgs{"testdata/datasources/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
				Datasources: Datasources{
					{
						Type: "amazon-ami",
						Name: "test",
					}: {
						Type: "amazon-ami",
						Name: "test",
					},
				},
			},
			false, false,
			[]packersdk.Build{},
			false,
		},
		{"recursive datasources",
			defaultParser,
			parseTestArgs{"testdata/datasources/recursive.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
				Datasources: Datasources{
					{
						Type: "null",
						Name: "foo",
					}: {
						Type: "null",
						Name: "foo",
					},
					{
						Type: "null",
						Name: "bar",
					}: {
						Type: "null",
						Name: "bar",
					},
					{
						Type: "null",
						Name: "baz",
					}: {
						Type: "null",
						Name: "baz",
					},
					{
						Type: "null",
						Name: "bang",
					}: {
						Type: "null",
						Name: "bang",
					},
					{
						Type: "null",
						Name: "yummy",
					}: {
						Type: "null",
						Name: "yummy",
					},
				},
			},
			false, false,
			[]packersdk.Build{},
			false,
		},
		{"untyped datasource",
			defaultParser,
			parseTestArgs{"testdata/datasources/untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
			},
			true, true,
			nil,
			false,
		},
		{"unnamed source",
			defaultParser,
			parseTestArgs{"testdata/datasources/unnamed.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
			},
			true, true,
			nil,
			false,
		},
		{"nonexistent source",
			defaultParser,
			parseTestArgs{"testdata/datasources/nonexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
				Datasources: Datasources{
					{
						Type: "nonexistent",
						Name: "test",
					}: {
						Type: "nonexistent",
						Name: "test",
					},
				},
			},
			true, true,
			nil,
			false,
		},
		{"duplicate source",
			defaultParser,
			parseTestArgs{"testdata/datasources/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
				Datasources: Datasources{
					{
						Type: "amazon-ami",
						Name: "test",
					}: {
						Type: "amazon-ami",
						Name: "test",
					},
				},
			},
			true, true,
			nil,
			false,
		},
		{"cyclic dependency between data sources",
			defaultParser,
			parseTestArgs{"testdata/datasources/dependency_cycle.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "datasources"),
				Datasources: Datasources{
					{
						Type: "null",
						Name: "gummy",
					}: {
						Type: "null",
						Name: "gummy",
					},
					{
						Type: "null",
						Name: "bear",
					}: {
						Type: "null",
						Name: "bear",
					},
				},
			},
			true, true,
			nil,
			false,
		},
	}

	testParse(t, tests)
}
