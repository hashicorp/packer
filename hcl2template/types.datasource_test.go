package hcl2template

import (
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestParse_datasource(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"two basic datasources",
			defaultParser,
			parseTestArgs{"testdata/datasources/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "datasources"),
				DataSources: DataSourcesMap{
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
		{"untyped datasource",
			defaultParser,
			parseTestArgs{"testdata/datasources/untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "datasources"),
			},
			true, true,
			nil,
			false,
		},
		{"unnamed source",
			defaultParser,
			parseTestArgs{"testdata/datasources/unnamed.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "datasources"),
			},
			true, true,
			nil,
			false,
		},
		{"inexistent source",
			defaultParser,
			parseTestArgs{"testdata/datasources/inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "datasources"),
			},
			true, true,
			nil,
			false,
		},
		{"duplicate source",
			defaultParser,
			parseTestArgs{"testdata/datasources/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "datasources"),
				DataSources: DataSourcesMap{
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
	}
	testParse(t, tests)
}
