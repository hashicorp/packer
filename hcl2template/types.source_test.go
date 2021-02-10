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
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
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
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
		},
		{"unnamed source",
			defaultParser,
			parseTestArgs{"testdata/sources/unnamed.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
		},
		{"unused source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "inexistant", Name: "ubuntu-1204"}: {Type: "inexistant", Name: "ubuntu-1204"},
				},
			},
			false, false,
			[]packersdk.Build{},
			false,
		},
		{"used source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/inexistent_used.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "inexistant", Name: "ubuntu-1204"}: {Type: "inexistant", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{Type: "inexistant", Name: "ubuntu-1204"},
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
		},
		{"duplicate source",
			defaultParser,
			parseTestArgs{"testdata/sources/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
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
