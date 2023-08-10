// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type MockConfig,NestedMockConfig,MockTag

package hcl2template

import (
	"context"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
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

func (b *MockConfig) Prepare(raws ...interface{}) error {
	for i, raw := range raws {
		cval, ok := raw.(cty.Value)
		if !ok {
			continue
		}
		b, err := json.Marshal(cval, cty.DynamicPseudoType)
		if err != nil {
			return err
		}
		ccval, err := json.Unmarshal(b, cty.DynamicPseudoType)
		if err != nil {
			return err
		}
		raws[i] = ccval
	}
	return config.Decode(b, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
}

//////
// MockBuilder
//////

type MockBuilder struct {
	Config MockConfig
}

var _ packersdk.Builder = new(MockBuilder)

func (b *MockBuilder) ConfigSpec() hcldec.ObjectSpec { return b.Config.FlatMapstructure().HCL2Spec() }

func (b *MockBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return []string{"ID"}, nil, b.Config.Prepare(raws...)
}

func (b *MockBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	return nil, nil
}

//////
// MockProvisioner
//////

type MockProvisioner struct {
	Config MockConfig
}

var _ packersdk.Provisioner = new(MockProvisioner)

func (b *MockProvisioner) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockProvisioner) Prepare(raws ...interface{}) error {
	return b.Config.Prepare(raws...)
}

func (b *MockProvisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, _ map[string]interface{}) error {
	return nil
}

//////
// MockDatasource
//////

type MockDatasource struct {
	Config MockConfig
}

var _ packersdk.Datasource = new(MockDatasource)

func (d *MockDatasource) ConfigSpec() hcldec.ObjectSpec {
	return d.Config.FlatMapstructure().HCL2Spec()
}

func (d *MockDatasource) OutputSpec() hcldec.ObjectSpec {
	return d.Config.FlatMapstructure().HCL2Spec()
}

func (d *MockDatasource) Configure(raws ...interface{}) error {
	return d.Config.Prepare(raws...)
}

func (d *MockDatasource) Execute() (cty.Value, error) {
	return hcl2helper.HCL2ValueFromConfig(d.Config, d.OutputSpec()), nil
}

//////
// MockPostProcessor
//////

type MockPostProcessor struct {
	Config MockConfig
}

var _ packersdk.PostProcessor = new(MockPostProcessor)

func (b *MockPostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockPostProcessor) Configure(raws ...interface{}) error {
	return b.Config.Prepare(raws...)
}

func (b *MockPostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, a packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	return nil, false, false, nil
}

//////
// MockCommunicator
//////

type MockCommunicator struct {
	Config MockConfig
	packersdk.Communicator
}

var _ packersdk.ConfigurableCommunicator = new(MockCommunicator)

func (b *MockCommunicator) ConfigSpec() hcldec.ObjectSpec {
	return b.Config.FlatMapstructure().HCL2Spec()
}

func (b *MockCommunicator) Configure(raws ...interface{}) ([]string, error) {
	return nil, b.Config.Prepare(raws...)
}

//////
// Utils
//////

type NamedMapStringString map[string]string
type NamedString string
