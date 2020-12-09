package hcl2template

import (
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestParse_source(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"two basic sources",
			defaultParser,
			parseTestArgs{"testdata/sources/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					},
				},
			},
			false, false,
			[]packersdk.Build{},
			false,
		},
		{"untyped source",
			defaultParser,
			parseTestArgs{"testdata/sources/untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
		},
		{"unnamed source",
			defaultParser,
			parseTestArgs{"testdata/sources/unnamed.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
		},
		{"inexistent source",
			defaultParser,
			parseTestArgs{"testdata/sources/inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
		},
		{"duplicate source",
			defaultParser,
			parseTestArgs{"testdata/sources/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
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
