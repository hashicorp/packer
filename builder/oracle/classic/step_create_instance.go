package classic

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type instanceOptions struct {
	Username       string
	IdentityDomain string
	SshKey         string
	Shape          string
	ImageList      string
	InstanceIP     string
}

var instanceTemplate = template.Must(template.New("instanceRequestBody").Parse(`
{
  "instances": [{
      "shape": "{{.Shape}}",
      "sshkeys": ["/Compute-{{.IdentityDomain}}/{{Username}}/{{.SshKey}}"],
      "name": "Compute-{{.IdentityDomain}}/{{Username}}/packer-instance",
      "label": "packer-instance",
      "imagelist": "/Compute-{{.IdentityDomain}}/{{Username}}/{{.ImageList}}",
      "networking": {
        "eth0": {
          "nat": "ipreservation:/Compute-{{.IdentityDomain}}/{{Username}}/{{.InstanceIP}}"
        }
      }
  }]
}
`))

type stepCreateInstance struct{}

func (s *stepCreateIPReservation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	const endpoint_path = "/launchplan/" // POST

	ui.Say("Creating Instance...")

	// generate launch plan definition for this instance
	var buffer bytes.Buffer
	err = instanceTemplate.Execute(&buffer, instanceOptions{
		Username:       config.Username,
		IdentityDomain: config.IdentityDomain,
		SshKey:         config.SshKey,
		Shape:          config.Shape,
		ImageList:      config.ImageList,
	})
	if err != nil {
		fmt.Printf("Error creating launch plan definition: %s", err)
		return "", err
	}
	// for API docs see
	// https://docs.oracle.com/en/cloud/iaas/compute-iaas-cloud/stcsa/op-launchplan--post.html
	instanceID, err := client.CreateInstance(publicKey)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_id", instanceID)

	ui.Say(fmt.Sprintf("Created instance (%s).", instanceID))
}
