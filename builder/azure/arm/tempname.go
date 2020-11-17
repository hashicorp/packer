package arm

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/random"
)

type TempName struct {
	AdminPassword       string
	CertificatePassword string
	ComputeName         string
	DeploymentName      string
	KeyVaultName        string
	ResourceGroupName   string
	OSDiskName          string
	DataDiskName        string
	NicName             string
	SubnetName          string
	PublicIPAddressName string
	VirtualNetworkName  string
	NsgName             string
}

func NewTempName(p string) *TempName {
	tempName := &TempName{}

	suffix := random.AlphaNumLower(5)
	if p == "" {
		p = "pkr"
		suffix = random.AlphaNumLower(10)
	}

	tempName.ComputeName = fmt.Sprintf("%svm%s", p, suffix)
	tempName.DeploymentName = fmt.Sprintf("%sdp%s", p, suffix)
	tempName.KeyVaultName = fmt.Sprintf("%skv%s", p, suffix)
	tempName.OSDiskName = fmt.Sprintf("%sos%s", p, suffix)
	tempName.DataDiskName = fmt.Sprintf("%sdd%s", p, suffix)
	tempName.NicName = fmt.Sprintf("%sni%s", p, suffix)
	tempName.PublicIPAddressName = fmt.Sprintf("%sip%s", p, suffix)
	tempName.SubnetName = fmt.Sprintf("%ssn%s", p, suffix)
	tempName.VirtualNetworkName = fmt.Sprintf("%svn%s", p, suffix)
	tempName.NsgName = fmt.Sprintf("%ssg%s", p, suffix)
	tempName.ResourceGroupName = fmt.Sprintf("%s-Resource-Group-%s", p, suffix)

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
