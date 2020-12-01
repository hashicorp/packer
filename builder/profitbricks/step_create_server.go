package profitbricks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
)

type stepCreateServer struct{}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)
	profitbricks.SetDepth("5")
	if sshkey, ok := state.GetOk("publicKey"); ok {
		c.SSHKey = sshkey.(string)
	}
	ui.Say("Creating Virtual Data Center...")
	img := s.getImageId(c.Image, c)
	alias := ""
	if img == "" {
		alias = s.getImageAlias(c.Image, c.Region, ui)
	}

	datacenter := profitbricks.Datacenter{
		Properties: profitbricks.DatacenterProperties{
			Name:     c.SnapshotName,
			Location: c.Region,
		},
	}
	server := profitbricks.Server{
		Properties: profitbricks.ServerProperties{
			Name:  c.SnapshotName,
			Ram:   c.Ram,
			Cores: c.Cores,
		},
		Entities: &profitbricks.ServerEntities{
			Volumes: &profitbricks.Volumes{
				Items: []profitbricks.Volume{
					{
						Properties: profitbricks.VolumeProperties{
							Type:       c.DiskType,
							Size:       c.DiskSize,
							Name:       c.SnapshotName,
							ImageAlias: alias,
							Image:      img,
						},
					},
				},
			},
		},
	}
	if c.SSHKey != "" {
		server.Entities.Volumes.Items[0].Properties.SshKeys = []string{c.SSHKey}
	}

	if c.Comm.SSHPassword != "" {
		server.Entities.Volumes.Items[0].Properties.ImagePassword = c.Comm.SSHPassword
	}

	datacenter = profitbricks.CompositeCreateDatacenter(datacenter)
	if datacenter.StatusCode > 299 {
		if datacenter.StatusCode > 299 {
			var restError RestError
			err := json.Unmarshal([]byte(datacenter.Response), &restError)
			if err != nil {
				ui.Error(fmt.Sprintf("Error decoding json response: %s", err.Error()))
				return multistep.ActionHalt
			}
			if len(restError.Messages) > 0 {
				ui.Error(restError.Messages[0].Message)
			} else {
				ui.Error(datacenter.Response)
			}
			return multistep.ActionHalt
		}
	}

	err := s.waitTillProvisioned(datacenter.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occurred while creating a datacenter %s", err.Error()))
		return multistep.ActionHalt
	}

	state.Put("datacenter_id", datacenter.Id)

	server = profitbricks.CreateServer(datacenter.Id, server)
	if server.StatusCode > 299 {
		ui.Error(fmt.Sprintf("Error occurred %s", parseErrorMessage(server.Response)))
		return multistep.ActionHalt
	}

	err = s.waitTillProvisioned(server.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occurred while creating a server %s", err.Error()))
		return multistep.ActionHalt
	}

	lan := profitbricks.CreateLan(datacenter.Id, profitbricks.CreateLanRequest{
		Properties: profitbricks.CreateLanProperties{
			Public: true,
			Name:   c.SnapshotName,
		},
	})

	if lan.StatusCode > 299 {
		ui.Error(fmt.Sprintf("Error occurred %s", parseErrorMessage(lan.Response)))
		return multistep.ActionHalt
	}

	err = s.waitTillProvisioned(lan.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occurred while creating a LAN %s", err.Error()))
		return multistep.ActionHalt
	}

	lanId, _ := strconv.Atoi(lan.Id)
	nic := profitbricks.CreateNic(datacenter.Id, server.Id, profitbricks.Nic{
		Properties: &profitbricks.NicProperties{
			Name: c.SnapshotName,
			Lan:  lanId,
			Dhcp: true,
		},
	})

	if lan.StatusCode > 299 {
		ui.Error(fmt.Sprintf("Error occurred %s", parseErrorMessage(nic.Response)))
		return multistep.ActionHalt
	}

	err = s.waitTillProvisioned(nic.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occurred while creating a NIC %s", err.Error()))
		return multistep.ActionHalt
	}

	state.Put("volume_id", server.Entities.Volumes.Items[0].Id)

	server = profitbricks.GetServer(datacenter.Id, server.Id)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", server.Id)

	state.Put("server_ip", server.Entities.Nics.Items[0].Properties.Ips[0])

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Removing Virtual Data Center...")

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)

	if dcId, ok := state.GetOk("datacenter_id"); ok {
		resp := profitbricks.DeleteDatacenter(dcId.(string))
		if err := s.checkForErrors(resp); err != nil {
			ui.Error(fmt.Sprintf(
				"Error deleting Virtual Data Center. Please destroy it manually: %s", err))
		}
		if err := s.waitTillProvisioned(resp.Headers.Get("Location"), *c); err != nil {
			ui.Error(fmt.Sprintf(
				"Error deleting Virtual Data Center. Please destroy it manually: %s", err))
		}
	}
}

func (d *stepCreateServer) waitTillProvisioned(path string, config Config) error {
	d.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	waitCount := 120
	if config.Retries > 0 {
		waitCount = config.Retries
	}
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.Metadata.Status == "DONE" {
			return nil
		}
		if request.Metadata.Status == "FAILED" {
			return errors.New(request.Metadata.Message)
		}
		time.Sleep(1 * time.Second)
		i++
	}
	return nil
}

func (d *stepCreateServer) setPB(username string, password string, url string) {
	profitbricks.SetAuth(username, password)
	profitbricks.SetEndpoint(url)
}

func (d *stepCreateServer) checkForErrors(instance profitbricks.Resp) error {
	if instance.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Error occurred %s", string(instance.Body)))
	}
	return nil
}

type RestError struct {
	HttpStatus int       `json:"httpStatus,omitempty"`
	Messages   []Message `json:"messages,omitempty"`
}

type Message struct {
	ErrorCode string `json:"errorCode,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (d *stepCreateServer) getImageId(imageName string, c *Config) string {
	d.setPB(c.PBUsername, c.PBPassword, c.PBUrl)

	images := profitbricks.ListImages()

	for i := 0; i < len(images.Items); i++ {
		imgName := ""
		if images.Items[i].Properties.Name != "" {
			imgName = images.Items[i].Properties.Name
		}
		diskType := c.DiskType
		if c.DiskType == "SSD" {
			diskType = "HDD"
		}
		if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && images.Items[i].Properties.ImageType == diskType && images.Items[i].Properties.Location == c.Region && images.Items[i].Properties.Public == true {
			return images.Items[i].Id
		}
	}
	return ""
}

func (d *stepCreateServer) getImageAlias(imageAlias string, location string, ui packersdk.Ui) string {
	if imageAlias == "" {
		return ""
	}
	locations := profitbricks.GetLocation(location)
	if len(locations.Properties.ImageAliases) > 0 {
		for _, i := range locations.Properties.ImageAliases {
			alias := ""
			if i != "" {
				alias = i
			}
			if alias != "" && strings.EqualFold(alias, imageAlias) {
				return alias
			}
		}
	}
	return ""
}

func parseErrorMessage(raw string) (toreturn string) {
	var tmp map[string]interface{}
	json.Unmarshal([]byte(raw), &tmp)

	for _, v := range tmp["messages"].([]interface{}) {
		for index, i := range v.(map[string]interface{}) {
			if index == "message" {
				toreturn = toreturn + i.(string) + "\n"
			}
		}
	}
	return toreturn
}
