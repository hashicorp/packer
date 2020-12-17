package dtl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/random"
)

type TempName struct {
	AdminPassword       string
	CertificatePassword string
	ComputeName         string
	DeploymentName      string
	KeyVaultName        string
	ResourceGroupName   string
	OSDiskName          string
	NicName             string
	SubnetName          string
	PublicIPAddressName string
	VirtualNetworkName  string
}

func NewTempName(c *Config) *TempName {
	tempName := &TempName{}
	suffix := random.AlphaNumLower(10)

	if c.VMName != "" {
		suffix = c.VMName
	}

	tempName.ComputeName = suffix
	tempName.DeploymentName = fmt.Sprintf("pkrdp%s", suffix)
	tempName.KeyVaultName = fmt.Sprintf("pkrkv%s", suffix)
	tempName.OSDiskName = fmt.Sprintf("pkros%s", suffix)
	tempName.NicName = tempName.ComputeName
	tempName.PublicIPAddressName = tempName.ComputeName
	tempName.SubnetName = fmt.Sprintf("pkrsn%s", suffix)
	tempName.VirtualNetworkName = fmt.Sprintf("pkrvn%s", suffix)
	tempName.ResourceGroupName = fmt.Sprintf("packer-Resource-Group-%s", suffix)

	tempName.AdminPassword = generatePassword()
	tempName.CertificatePassword = random.AlphaNum(32)
	return tempName
}

// generate a password that is acceptable to Azure
// Three of the four items must be met.
//  1. Contains an uppercase character
//  2. Contains a lowercase character
//  3. Contains a numeric digit
//  4. Contains a special character
func generatePassword() string {
	var s string
	for i := 0; i < 100; i++ {
		s := random.AlphaNum(32)
		if !strings.ContainsAny(s, random.PossibleNumbers) {
			continue
		}

		if !strings.ContainsAny(s, random.PossibleLowerCase) {
			continue
		}

		if !strings.ContainsAny(s, random.PossibleUpperCase) {
			continue
		}

		return s
	}

	// if an acceptable password cannot be generated in 100 tries, give up
	return s
}
