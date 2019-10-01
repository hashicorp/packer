package common

import (
	"bytes"
	"encoding/base64"
	"log"
	"text/template"

	"github.com/hashicorp/packer/common/random"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/ssh"
)

// TemplateCloudInitDefault is a basic go template for a cloud-init template to set a default ssh key
const TemplateCloudInitDefault = `
#cloud-config
#vim:syntax=yaml
users:
  - default
ssh_authorized_keys:
  - "{{.SSHPublicKey}}"
`

// CloudInitTemplate is the base struct for the communicator
type CloudInitTemplate struct {
	SSHPublicKey string
}

// GenerateAndAttachWinrmCredentials will create the default Windows creds in a Nutanix-way
func GenerateAndAttachWinrmCredentials(config *communicator.Config) string {
	tmpPass := random.String(random.PossibleAlphaNumUpper, 8)
	config.WinRMPassword = tmpPass
	log.Printf("WinRM temporary password: %s", tmpPass)
	return tmpPass
}

// GenerateAndAttachSSHKey will generate the random ssh key to be used for ssh
func GenerateAndAttachSSHKey(config *communicator.Config) string {
	// Create KeyPair
	sshconfig, err := ssh.NewKeyPair(
		ssh.CreateKeyPairConfig{
			Bits: 2048,
			Type: ssh.Rsa,
		},
	)
	config.SSHPrivateKey = []byte(sshconfig.PrivateKeyPemBlock)

	//sshKey64 := base64.StdEncoding.EncodeToString(sshconfig.PrivateKeyPemBlock)
	//sshVal64 := base64.StdEncoding.EncodeToString(sshconfig.PublicKeyAuthorizedKeysLine)
	tmpl, err := template.New("cloud-init").Parse(TemplateCloudInitDefault)
	if err != nil {
		panic(err)
		//return multistep.ActionHalt
	}
	var cloudInitUserData bytes.Buffer
	cloudInitVals := &CloudInitTemplate{
		SSHPublicKey: string(sshconfig.PublicKeyAuthorizedKeysLine),
	}
	tmpl.Execute(&cloudInitUserData, cloudInitVals)
	return base64.StdEncoding.EncodeToString(cloudInitUserData.Bytes())
	//reqJson.Spec.Resources.GuestCustomization.CloudInit.UserData = "I2Nsb3VkLWNvbmZpZwojdmltOnN5bnRheD15YW1sCnVzZXJzOgogIC0gZGVmYXVsdApzc2hfYXV0aG9yaXplZF9rZXlzOgogIC0gInNzaC1yc2EgQUFBQUIzTnphQzF5YzJFQUFBQURBUUFCQUFBQkFRREkyK2ltb1pRYkloTk5aYlB2NTh1RGM4bDlubElTZGgzS0t3aFFPU0pPM2V6bm15ME5Ndk05NzFxREh3a1Y2L3BvTnJZdkU3dDhIRXl5ZWZvdFREdnZZWGRnc01QTjZNZjYyd2JNakVpd1R6STRLZHpjOFVOME1iUWhRdDFzN1JGaTdZQ2dUWGk4Rit2dlhBd0JHVWhkcWFqeDg4ZXExRmFOaUpwOCtiUTRIS0IrdkdmbTBWU1JxdG95dkJBQ1dpbHdwU25XZUFpMnYzYlRKMEUyRFJnOE92UXFIUDJBWGhKL1NkTDdUOWtIcW9saTRrb1gvdTlaVytOam84Mng1aklCWFpTWTFMY2dGRXJJRkpldDVtOHpnYzJCNWhFekpIRWhrWDVXdnpjaHQyaEVaM0xRVGg1YjBwMXZzVTJ1aHdmalNHN2ZlYUpoMEpFc1pZelJSV24zIgpjaHBhc3N3ZDoKICBsaXN0OiB8CiAgICBlYzItdXNlcjpwYXNzdzByZAo="
	//s.Config.Config.Type = "ssh"
	/*s.Config.Config.SSHAgentAuth = false
	s.Config.Config.SSHPort = 22
	s.Config.Config.SSHTimeout = 5 * time.Minute*/
	//s.Config.Config.SSHPrivateKey = []byte(base64.StdEncoding.EncodeToString([]byte(sshconfig.PrivateKeyPemBlock)))

	//log.Printf("Private key: " + sshKey64)
	//	log.Printf("Public key: " + sshVal64)

}
