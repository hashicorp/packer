// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
)

type DeploymentFactory struct {
	template string
}

func newDeploymentFactory(template string) DeploymentFactory {
	return DeploymentFactory{
		template: template,
	}
}

func (f *DeploymentFactory) create(templateParameters TemplateParameters) (*resources.Deployment, error) {
	template, err := f.getTemplate(templateParameters)
	if err != nil {
		return nil, err
	}

	parameters, err := f.getTemplateParameters(templateParameters)
	if err != nil {
		return nil, err
	}

	return &resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Mode:       resources.Incremental,
			Template:   template,
			Parameters: parameters,
		},
	}, nil
}

func (f *DeploymentFactory) getTemplate(templateParameters TemplateParameters) (*map[string]interface{}, error) {
	var t map[string]interface{}
	err := json.Unmarshal([]byte(f.template), &t)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (f *DeploymentFactory) getTemplateParameters(templateParameters TemplateParameters) (*map[string]interface{}, error) {
	b, err := json.Marshal(templateParameters)
	if err != nil {
		return nil, err
	}

	var t map[string]interface{}
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
