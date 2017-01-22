package profitbricks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
	"strconv"
	"strings"
	"time"
)

type stepCreateServer struct{}

func (s *stepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)
	profitbricks.SetDepth("5")
	if sshkey, ok := state.GetOk("publicKey"); ok {
		c.SSHKey = sshkey.(string)
	}
	ui.Say("Creating Virtual Data Center...")
	img := s.getImageId(c.Image, c)

	datacenter := profitbricks.Datacenter{
		Properties: profitbricks.DatacenterProperties{
			Name:     c.SnapshotName,
			Location: c.Region,
		},
		Entities: profitbricks.DatacenterEntities{
			Servers: &profitbricks.Servers{
				Items: []profitbricks.Server{
					{
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
											Type:  c.DiskType,
											Size:  c.DiskSize,
											Name:  c.SnapshotName,
											Image: img,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if c.SSHKey != "" {
		datacenter.Entities.Servers.Items[0].Entities.Volumes.Items[0].Properties.SshKeys = []string{c.SSHKey}
	}

	if c.Comm.SSHPassword != "" {
		datacenter.Entities.Servers.Items[0].Entities.Volumes.Items[0].Properties.ImagePassword = c.Comm.SSHPassword
	}

	datacenter = profitbricks.CompositeCreateDatacenter(datacenter)
	if datacenter.StatusCode > 299 {
		if datacenter.StatusCode > 299 {
			var restError RestError
			json.Unmarshal([]byte(datacenter.Response), &restError)
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
		ui.Error(fmt.Sprintf("Error occured while creating a datacenter %s", err.Error()))
		return multistep.ActionHalt
	}

	state.Put("datacenter_id", datacenter.Id)

	lan := profitbricks.CreateLan(datacenter.Id, profitbricks.Lan{
		Properties: profitbricks.LanProperties{
			Public: true,
			Name:   c.SnapshotName,
		},
	})

	if lan.StatusCode > 299 {
		ui.Error(fmt.Sprintf("Error occured %s", parseErrorMessage(lan.Response)))
		return multistep.ActionHalt
	}

	err = s.waitTillProvisioned(lan.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occured while creating a LAN %s", err.Error()))
		return multistep.ActionHalt
	}

	lanId, _ := strconv.Atoi(lan.Id)
	nic := profitbricks.CreateNic(datacenter.Id, datacenter.Entities.Servers.Items[0].Id, profitbricks.Nic{
		Properties: profitbricks.NicProperties{
			Name: c.SnapshotName,
			Lan:  lanId,
			Dhcp: true,
		},
	})

	if lan.StatusCode > 299 {
		ui.Error(fmt.Sprintf("Error occured %s", parseErrorMessage(nic.Response)))
		return multistep.ActionHalt
	}

	err = s.waitTillProvisioned(nic.Headers.Get("Location"), *c)
	if err != nil {
		ui.Error(fmt.Sprintf("Error occured while creating a NIC %s", err.Error()))
		return multistep.ActionHalt
	}

	state.Put("volume_id", datacenter.Entities.Servers.Items[0].Entities.Volumes.Items[0].Id)

	server := profitbricks.GetServer(datacenter.Id, datacenter.Entities.Servers.Items[0].Id)

	state.Put("server_ip", server.Entities.Nics.Items[0].Properties.Ips[0])

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Removing Virtual Data Center...")

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)

	if dcId, ok := state.GetOk("datacenter_id"); ok {
		resp := profitbricks.DeleteDatacenter(dcId.(string))
		s.checkForErrors(resp)
		err := s.waitTillProvisioned(resp.Headers.Get("Location"), *c)
		if err != nil {
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
		return errors.New(fmt.Sprintf("Error occured %s", string(instance.Body)))
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
