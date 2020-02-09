//go:generate mapstructure-to-hcl2 -type MockConfig,NestedMockConfig

package hcl2template

import (
	"context"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
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
}

type MockConfig struct {
	NotSquashed      string `mapstructure:"not_squashed"`
	NestedMockConfig `mapstructure:",squash"`
	Nested           NestedMockConfig   `mapstructure:"nested"`
	NestedSlice      []NestedMockConfig `mapstructure:"nested_slice"`
}

//////
// MockBuilder
//////

type MockBuilder struct {
	Config MockConfig
}

var _ packer.Builder = new(MockBuilder)

func (b *MockBuilder) ConfigSpec() hcldec.ObjectSpec { return b.Config.FlatMapstructure().HCL2Spec() }

func (b *MockBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return nil, nil, config.Decode(&b.Config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
}

func (b *MockBuilder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	return nil, nil
}

//////
// MockProvisioner
//////

type MockProvisioner struct {
	Config MockConfig
}

var _ packer.Provisioner = new(MockProvisioner)

func (b *MockProvisioner) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockProvisioner) Prepare(raws ...interface{}) error {
	return config.Decode(&b.Config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
}

func (b *MockProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, _ map[string]interface{}) error {
	return nil
}

//////
// MockPostProcessor
//////

type MockPostProcessor struct {
	Config MockConfig
}

var _ packer.PostProcessor = new(MockPostProcessor)

func (b *MockPostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockPostProcessor) Configure(raws ...interface{}) error {
	return config.Decode(&b.Config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
}

func (b *MockPostProcessor) PostProcess(ctx context.Context, ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, bool, error) {
	return nil, false, false, nil
}

//////
// MockCommunicator
//////

type MockCommunicator struct {
	Config MockConfig
	packer.Communicator
}

var _ packer.ConfigurableCommunicator = new(MockCommunicator)

func (b *MockCommunicator) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockCommunicator) Configure(raws ...interface{}) ([]string, error) {
	return nil, config.Decode(&b.Config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
}

//////
// Utils
//////

type NamedMapStringString map[string]string
type NamedString string
