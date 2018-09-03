package arm

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/common/random"
)

const (
	TempNameAlphabet = "0123456789bcdfghjklmnpqrstvwxyz"

	numbers   = "0123456789"
	lowerCase = "abcdefghijklmnopqrstuvwxyz"
	upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
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

func NewTempName() *TempName {
	tempName := &TempName{}

	suffix := random.String(TempNameAlphabet, 10)
	tempName.ComputeName = fmt.Sprintf("pkrvm%s", suffix)
	tempName.DeploymentName = fmt.Sprintf("pkrdp%s", suffix)
	tempName.KeyVaultName = fmt.Sprintf("pkrkv%s", suffix)
	tempName.OSDiskName = fmt.Sprintf("pkros%s", suffix)
	tempName.NicName = fmt.Sprintf("pkrni%s", suffix)
	tempName.PublicIPAddressName = fmt.Sprintf("pkrip%s", suffix)
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
		if !strings.ContainsAny(s, numbers) {
			continue
		}

		if !strings.ContainsAny(s, lowerCase) {
			continue
		}

		if !strings.ContainsAny(s, upperCase) {
			continue
		}

		return s
	}

	// if an acceptable password cannot be generated in 100 tries, give up
	return s
}
