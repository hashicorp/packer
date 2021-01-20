//go:generate mapstructure-to-hcl2 -type MockConfig,NestedMockConfig,MockTag

package hcl2shim

import (
	"time"

	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

type NestedMockConfig struct {
	String               string               `mapstructure:"string"`
	Int                  int                  `mapstructure:"int"`
	Int64                int64                `mapstructure:"int64"`
	Bool                 bool                 `mapstructure:"bool"`
	Trilean              config.Trilean       `mapstructure:"trilean"`
	Duration             time.Duration        `mapstructure:"duration"`
	MapStringString      map[string]string    `mapstructure:"map_string_string"`
	SliceString          []string             `mapstructure:"slice_string"`
	SliceSliceString     [][]string           `mapstructure:"slice_slice_string"`
	NamedMapStringString NamedMapStringString `mapstructure:"named_map_string_string"`
	NamedString          NamedString          `mapstructure:"named_string"`
	Tags                 []MockTag            `mapstructure:"tag"`
	Datasource           string               `mapstructure:"data_source"`
}

type MockTag struct {
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}

type MockConfig struct {
	NotSquashed      string `mapstructure:"not_squashed"`
	NestedMockConfig `mapstructure:",squash"`
	Nested           NestedMockConfig   `mapstructure:"nested"`
	NestedSlice      []NestedMockConfig `mapstructure:"nested_slice"`
}

type NamedMapStringString map[string]string
type NamedString string
