package profitbricks

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
	"strings"
	"time"
	"github.com/profitbricks/profitbricks-sdk-go/model"
)

const (
	waitCount = 30
)

type stepCreateServer struct{}

func (s *stepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)
	profitbricks.SetDepth("5")
	c.SSHKey = state.Get("publicKey").(string)

	ui.Say("Creating Virutal Data Center...")
	img := s.getImageId(c.Image, c)

	datacenter := model.Datacenter{
		Properties: model.DatacenterProperties{
			Name: c.SnapshotName,
			Location:c.Region,
		},
		Entities:model.DatacenterEntities{
			Servers: &model.Servers{
				Items:[]model.Server{
					model.Server{
						Properties: model.ServerProperties{
							Name : c.SnapshotName,
							Ram: c.Ram,
							Cores: c.Cores,
						},
						Entities:model.ServerEntities{
							Volumes: &model.AttachedVolumes{
								Items:[]model.Volume{
									model.Volume{
										Properties: model.VolumeProperties{
											Type_:c.DiskType,
											Size:c.DiskSize,
											Name:c.SnapshotName,
											Image:img,
											ImagePassword: "test1234",
											SshKeys: []string{c.SSHKey},
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

	datacenter = profitbricks.CompositeCreateDatacenter(datacenter)
	if datacenter.StatusCode > 299 {
		ui.Error(datacenter.Response)
		return multistep.ActionHalt
	}
	s.waitTillProvisioned(datacenter.Headers.Get("Location"), *c)

	state.Put("datacenter_id", datacenter.Id)

	lan := profitbricks.CreateLan(datacenter.Id, profitbricks.CreateLanRequest{
		LanProperties: profitbricks.LanProperties{
			Public: true,
			Name:   c.SnapshotName,
		},
	})

	err := s.checkForErrors(lan.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(lan.Resp.Headers["Location"], ""), *c)

	nic := profitbricks.CreateNic(datacenter.Id, datacenter.Entities.Servers.Items[0].Id, profitbricks.NicCreateRequest{
		NicProperties : profitbricks.NicProperties{
			Name: c.SnapshotName,
			Lan: lan.Id,
			Dhcp: true,
		},
	})

	err = s.checkForErrors(nic.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(nic.Resp.Headers["Location"], ""), *c)

	state.Put("volume_id", datacenter.Entities.Servers.Items[0].Entities.Volumes.Items[0].Id)

	server := profitbricks.GetServer(datacenter.Id, datacenter.Entities.Servers.Items[0].Id)

	state.Put("server_ip", server.Entities["nics"].Items[0].Properties["ips"].([]interface{})[0].(string))

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Removing Virtual Data Center...")

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)
	dcId := state.Get("datacenter_id").(string)

	resp := profitbricks.DeleteDatacenter(dcId)

	s.checkForErrors(resp)

	err := s.waitTillProvisioned(strings.Join(resp.Headers["Location"], ""), *c)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting Virtual Data Center. Please destroy it manually: %s", err))
	}
}

func (d *stepCreateServer) waitTillProvisioned(path string, config Config) error {
	d.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.MetaData["status"] == "DONE" {
			return nil
		}
		if request.MetaData["status"] == "FAILED" {
			return errors.New(request.MetaData["message"])
		}
		time.Sleep(10 * time.Second)
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

func (d *stepCreateServer) getImageId(imageName string, c *Config) string {
	d.setPB(c.PBUsername, c.PBPassword, c.PBUrl)

	images := profitbricks.ListImages()

	for i := 0; i < len(images.Items); i++ {
		imgName := ""
		if images.Items[i].Properties["name"] != nil {
			imgName = images.Items[i].Properties["name"].(string)
		}
		diskType := c.DiskType
		if c.DiskType == "SSD" {
			diskType = "HDD"
		}
		if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && images.Items[i].Properties["imageType"] == diskType && images.Items[i].Properties["location"] == c.Region {
			return images.Items[i].Id
		}
	}
	return ""
}
