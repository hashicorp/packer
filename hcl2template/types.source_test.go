// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
)

func TestParse_source(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"two basic sources",
			defaultParser,
			parseTestArgs{"testdata/sources/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Basedir: filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					},
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					SensitiveVars:  []string{},
				},
			},
			false,
			nil,
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
			nil,
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
			nil,
		},
		{"unused source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/nonexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds:                  nil,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "nonexistent", Name: "ubuntu-1204"}: {Type: "nonexistent", Name: "ubuntu-1204"},
				},
			},
			true, true,
			[]*packer.CoreBuild{},
			false,
			nil,
		},
		{"used source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/nonexistent_used.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "nonexistent", Name: "ubuntu-1204"}: {Type: "nonexistent", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{Type: "nonexistent", Name: "ubuntu-1204"},
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			nil,
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
			nil,
		},
	}
	testParse(t, tests)
}
