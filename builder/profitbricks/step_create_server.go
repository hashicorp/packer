package profitbricks

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
	"strings"
	"time"
)

const (
	waitCount = 30
)

type stepCreateServer struct{}

func (s *stepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)

	ui.Say("Creating Virutal datacenter...")

	datacenter := profitbricks.CreateDatacenter(profitbricks.CreateDatacenterRequest{
		DCProperties: profitbricks.DCProperties{
			Name:     c.ServerName,
			Location: c.Region,
		},
	})

	err := s.checkForErrors(datacenter.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(datacenter.Resp.Headers["Location"], ""), *c)

	state.Put("datacenter_id", datacenter.Id)

	ui.Say("Creating ProfitBricks server...")

	server := profitbricks.CreateServer(datacenter.Id, profitbricks.CreateServerRequest{
		ServerProperties: profitbricks.ServerProperties{
			Name:  c.ServerName,
			Ram:   c.Ram,
			Cores: c.Cores,
		},
	})

	err = s.checkForErrors(server.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(server.Resp.Headers["Location"], ""), *c)

	ui.Say("Creating a volume...")

	c.SSHKey = state.Get("publicKey").(string)

	img := s.getImageId(c.Image, c)

	volume := profitbricks.CreateVolume(datacenter.Id, profitbricks.CreateVolumeRequest{
		VolumeProperties: profitbricks.VolumeProperties{
			Size:   c.DiskSize,
			Name:   c.ServerName,
			Image:  img,
			Type:   c.DiskType,
			SshKey: []string{c.SSHKey},
		},
	})

	err = s.checkForErrors(volume.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(volume.Resp.Headers["Location"], ""), *c)

	attachresponse := profitbricks.AttachVolume(datacenter.Id, server.Id, volume.Id)

	s.waitTillProvisioned(strings.Join(attachresponse.Resp.Headers["Location"], ""), *c)
	ui.Say("Creating a LAN...")

	lan := profitbricks.CreateLan(datacenter.Id, profitbricks.CreateLanRequest{
		LanProperties: profitbricks.LanProperties{
			Public: true,
			Name:   c.ServerName,
		},
	})

	err = s.checkForErrors(lan.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(lan.Resp.Headers["Location"], ""), *c)

	ui.Say("Creating a NIC...")

	nic := profitbricks.CreateNic(datacenter.Id, server.Id, profitbricks.NicCreateRequest{
		NicProperties: profitbricks.NicProperties{
			Name: c.ServerName,
			Lan:  lan.Id,
			Dhcp: true,
		},
	})

	err = s.checkForErrors(nic.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(nic.Resp.Headers["Location"], ""), *c)

	bootVolume := profitbricks.Instance{
		Properties: nil,
		Entities:   nil,
		MetaData:   nil,
	}

	bootVolume.Id = volume.Id

	serverpatchresponse := profitbricks.PatchServer(datacenter.Id, server.Id, profitbricks.ServerProperties{
		BootVolume: &bootVolume,
	})

	state.Put("volume_id", volume.Id)

	err = s.checkForErrors(serverpatchresponse.Resp)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.waitTillProvisioned(strings.Join(serverpatchresponse.Resp.Headers["Location"], ""), *c)

	server = profitbricks.GetServer(datacenter.Id, server.Id)

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
