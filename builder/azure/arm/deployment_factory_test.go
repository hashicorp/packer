// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"testing"
)

func TestDeploymentFactoryShouldBeIncremental(t *testing.T) {
	var testSubject = newDeploymentFactory(Linux)

	deployment, err := testSubject.create(getTemplateParameters())
	if err != nil {
		t.Fatalf("ERROR: %s\n", err)
	}

	if deployment.Properties.Mode != "Incremental" {
		t.Fatalf("Expected the mode to be 'Incremental', but got '%s'.", deployment.Properties.Mode)
	}
}

// Either {Template,Parameter} are set or {Template,Parameter}Link values are
// set, but never both.
func TestDeploymentFactoryShouldNotSetLinks(t *testing.T) {
	testSubject := newDeploymentFactory(Linux)

	deployment, err := testSubject.create(getTemplateParameters())
	if err != nil {
		t.Fatalf("ERROR: %s\n", err)
	}

	if deployment.Properties.ParametersLink != nil {
		t.Fatalf("Expected the ParametersLink to be nil!")
	}

	if deployment.Properties.TemplateLink != nil {
		t.Fatalf("Expected the TemplateLink to be nil!")
	}

	if deployment.Properties.Parameters == nil {
		t.Fatalf("Expected the Parameters to not be nil!")
	}

	if deployment.Properties.Template == nil {
		t.Fatalf("Expected the Template to not be nil!")
	}
}

func TestFactoryShouldCreateDeploymentInstance(t *testing.T) {
	testSubject := newDeploymentFactory(Linux)

	deployment, err := testSubject.create(getTemplateParameters())
	if err != nil {
		t.Fatalf("ERROR: %s\n", err)
	}

	// spot check well known values to ensure correct serialization.

	parametersMap := *deployment.Properties.Parameters
	if _, ok := parametersMap["adminUsername"]; ok == false {
		t.Fatalf("Expected the parameter value 'adminUsername' to be set, but it was not")
	}

	templateMap := *deployment.Properties.Template
	if _, ok := templateMap["contentVersion"]; ok == false {
		t.Fatalf("Expected the parameter value 'contentVersion' to be set, but it was not")
	}
}

func TestMalformedTemplatesShouldReturnError(t *testing.T) {
	testSubject := newDeploymentFactory("")

	_, err := testSubject.create(getTemplateParameters())
	if err == nil {
		t.Fatalf("Expected an error, but did not receive one!\n")
	}
}

func getTemplateParameters() TemplateParameters {
	templateParameters := TemplateParameters{
		AdminUsername:              &TemplateParameter{"adminusername00"},
		DnsNameForPublicIP:         &TemplateParameter{"dnsnameforpublicip00"},
		OSDiskName:                 &TemplateParameter{"osdiskname00"},
		SshAuthorizedKey:           &TemplateParameter{"sshkeydata00"},
		StorageAccountBlobEndpoint: &TemplateParameter{"storageaccountblobendpoint00"},
		VMName: &TemplateParameter{"vmname00"},
		VMSize: &TemplateParameter{"vmsize00"},
	}

	return templateParameters
}
