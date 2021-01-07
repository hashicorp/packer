package packer

import (
	"fmt"
)

type MapOfProvisioner map[string]func() (Provisioner, error)

func (mop MapOfProvisioner) Has(provisioner string) bool {
	_, res := mop[provisioner]
	return res
}

func (mop MapOfProvisioner) Start(provisioner string) (Provisioner, error) {
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

type MapOfPostProcessor map[string]func() (PostProcessor, error)

func (mopp MapOfPostProcessor) Has(postProcessor string) bool {
	_, res := mopp[postProcessor]
	return res
}

func (mopp MapOfPostProcessor) Start(postProcessor string) (PostProcessor, error) {
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

type MapOfBuilder map[string]func() (Builder, error)

func (mob MapOfBuilder) Has(builder string) bool {
	_, res := mob[builder]
	return res
}

func (mob MapOfBuilder) Start(builder string) (Builder, error) {
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

type MapOfDatasource map[string]func() (Datasource, error)

func (mob MapOfDatasource) Has(dataSource string) bool {
	_, res := mob[dataSource]
	return res
}

func (mob MapOfDatasource) Start(dataSource string) (Datasource, error) {
	d, found := mob[dataSource]
	if !found {
		return nil, fmt.Errorf("Unknown data source %s", dataSource)
	}
	return d()
}

func (mob MapOfDatasource) List() []string {
	res := []string{}
	for k := range mob {
		res = append(res, k)
	}
	return res
}
