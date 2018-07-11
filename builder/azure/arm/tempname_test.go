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

	if strings.Index(tempName.NicName, "pkrni") != 0 {
		t.Errorf("Expected NicName to begin with 'pkrni', but got '%s'!", tempName.NicName)
	}

	if strings.Index(tempName.PublicIPAddressName, "pkrip") != 0 {
		t.Errorf("Expected PublicIPAddressName to begin with 'pkrip', but got '%s'!", tempName.PublicIPAddressName)
	}

	if strings.Index(tempName.ResourceGroupName, "packer-Resource-Group-") != 0 {
		t.Errorf("Expected ResourceGroupName to begin with 'packer-Resource-Group-', but got '%s'!", tempName.ResourceGroupName)
	}

	if strings.Index(tempName.SubnetName, "pkrsn") != 0 {
		t.Errorf("Expected SubnetName to begin with 'pkrip', but got '%s'!", tempName.SubnetName)
	}

	if strings.Index(tempName.VirtualNetworkName, "pkrvn") != 0 {
		t.Errorf("Expected VirtualNetworkName to begin with 'pkrvn', but got '%s'!", tempName.VirtualNetworkName)
	}
}

func TestTempAdminPassword(t *testing.T) {
	tempName := NewTempName()

	if !strings.ContainsAny(tempName.AdminPassword, numbers) {
		t.Errorf("Expected AdminPassword to contain at least one of '%s'!", numbers)
	}
	if !strings.ContainsAny(tempName.AdminPassword, lowerCase) {
		t.Errorf("Expected AdminPassword to contain at least one of '%s'!", lowerCase)
	}
	if !strings.ContainsAny(tempName.AdminPassword, upperCase) {
		t.Errorf("Expected AdminPassword to contain at least one of '%s'!", upperCase)
	}
}

func TestTempNameShouldHaveSameSuffix(t *testing.T) {
	tempName := NewTempName()
	suffix := tempName.ComputeName[5:]

	if strings.HasSuffix(tempName.ComputeName, suffix) != true {
		t.Errorf("Expected ComputeName to end with '%s', but the value is '%s'!", suffix, tempName.ComputeName)
	}

	if strings.HasSuffix(tempName.DeploymentName, suffix) != true {
		t.Errorf("Expected DeploymentName to end with '%s', but the value is '%s'!", suffix, tempName.DeploymentName)
	}

	if strings.HasSuffix(tempName.OSDiskName, suffix) != true {
		t.Errorf("Expected OSDiskName to end with '%s', but the value is '%s'!", suffix, tempName.OSDiskName)
	}

	if strings.HasSuffix(tempName.NicName, suffix) != true {
		t.Errorf("Expected NicName to end with '%s', but the value is '%s'!", suffix, tempName.PublicIPAddressName)
	}

	if strings.HasSuffix(tempName.PublicIPAddressName, suffix) != true {
		t.Errorf("Expected PublicIPAddressName to end with '%s', but the value is '%s'!", suffix, tempName.PublicIPAddressName)
	}

	if strings.HasSuffix(tempName.ResourceGroupName, suffix) != true {
		t.Errorf("Expected ResourceGroupName to end with '%s', but the value is '%s'!", suffix, tempName.ResourceGroupName)
	}

	if strings.HasSuffix(tempName.SubnetName, suffix) != true {
		t.Errorf("Expected SubnetName to end with '%s', but the value is '%s'!", suffix, tempName.SubnetName)
	}

	if strings.HasSuffix(tempName.VirtualNetworkName, suffix) != true {
		t.Errorf("Expected VirtualNetworkName to end with '%s', but the value is '%s'!", suffix, tempName.VirtualNetworkName)
	}
}
