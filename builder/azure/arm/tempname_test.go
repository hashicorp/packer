// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"strings"
	"testing"
)

func TestTempNameShouldCreatePrefixedRandomNames(t *testing.T) {
	tempName := NewTempName()

	if strings.Index(tempName.ComputeName, "pkrvm") != 0 {
		t.Errorf("Expected ComputeName to begin with 'pkrvm', but got '%s'!", tempName.ComputeName)
	}

	if strings.Index(tempName.DeploymentName, "pkrdp") != 0 {
		t.Errorf("Expected ComputeName to begin with 'pkrdp', but got '%s'!", tempName.ComputeName)
	}

	if strings.Index(tempName.OSDiskName, "pkros") != 0 {
		t.Errorf("Expected OSDiskName to begin with 'pkros', but got '%s'!", tempName.OSDiskName)
	}

	if strings.Index(tempName.ResourceGroupName, "packer-Resource-Group-") != 0 {
		t.Errorf("Expected ResourceGroupName to begin with 'packer-Resource-Group-', but got '%s'!", tempName.ResourceGroupName)
	}
}

func TestTempNameShouldHaveSameSuffix(t *testing.T) {
	tempName := NewTempName()
	suffix := tempName.ComputeName[5:]

	if strings.HasSuffix(tempName.DeploymentName, suffix) != true {
		t.Errorf("Expected DeploymentName to end with '%s', but the value is '%s'!", suffix, tempName.DeploymentName)
	}

	if strings.HasSuffix(tempName.OSDiskName, suffix) != true {
		t.Errorf("Expected OSDiskName to end with '%s', but the value is '%s'!", suffix, tempName.OSDiskName)
	}

	if strings.HasSuffix(tempName.ResourceGroupName, suffix) != true {
		t.Errorf("Expected ResourceGroupName to end with '%s', but the value is '%s'!", suffix, tempName.ResourceGroupName)
	}

}
