// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"fmt"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type MapOfProvisioner map[string]func() (packersdk.Provisioner, error)

func (mop MapOfProvisioner) Has(provisioner string) bool {
	_, res := mop[provisioner]
	return res
}

func (mop MapOfProvisioner) Set(provisioner string, starter func() (packersdk.Provisioner, error)) {
	mop[provisioner] = starter
}

func (mop MapOfProvisioner) Start(provisioner string) (packersdk.Provisioner, error) {
	p, found := mop[provisioner]
	if !found {
		return nil, fmt.Errorf("Unknown provisioner %s", provisioner)
	}
	return p()
}

func (mop MapOfProvisioner) List() []string {
	res := []string{}
	for k := range mop {
		res = append(res, k)
	}
	return res
}

type MapOfPostProcessor map[string]func() (packersdk.PostProcessor, error)

func (mopp MapOfPostProcessor) Has(postProcessor string) bool {
	_, res := mopp[postProcessor]
	return res
}

func (mopp MapOfPostProcessor) Set(postProcessor string, starter func() (packersdk.PostProcessor, error)) {
	mopp[postProcessor] = starter
}

func (mopp MapOfPostProcessor) Start(postProcessor string) (packersdk.PostProcessor, error) {
	p, found := mopp[postProcessor]
	if !found {
		return nil, fmt.Errorf("Unknown post-processor %s", postProcessor)
	}
	return p()
}

func (mopp MapOfPostProcessor) List() []string {
	res := []string{}
	for k := range mopp {
		res = append(res, k)
	}
	return res
}

type MapOfBuilder map[string]func() (packersdk.Builder, error)

func (mob MapOfBuilder) Has(builder string) bool {
	_, res := mob[builder]
	return res
}

func (mob MapOfBuilder) Set(builder string, starter func() (packersdk.Builder, error)) {
	mob[builder] = starter
}

func (mob MapOfBuilder) Start(builder string) (packersdk.Builder, error) {
	d, found := mob[builder]
	if !found {
		return nil, fmt.Errorf("Unknown builder %s", builder)
	}
	return d()
}

func (mob MapOfBuilder) List() []string {
	res := []string{}
	for k := range mob {
		res = append(res, k)
	}
	return res
}

type MapOfDatasource map[string]func() (packersdk.Datasource, error)

func (mod MapOfDatasource) Has(dataSource string) bool {
	_, res := mod[dataSource]
	return res
}

func (mod MapOfDatasource) Set(dataSource string, starter func() (packersdk.Datasource, error)) {
	mod[dataSource] = starter
}

func (mod MapOfDatasource) Start(dataSource string) (packersdk.Datasource, error) {
	d, found := mod[dataSource]
	if !found {
		return nil, fmt.Errorf("Unknown data source %s", dataSource)
	}
	return d()
}

func (mod MapOfDatasource) List() []string {
	res := []string{}
	for k := range mod {
		res = append(res, k)
	}
	return res
}
