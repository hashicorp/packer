package profitbricks

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func mkVolume(dcID string) string {

	var request = Volume{
		Properties: VolumeProperties{
			Size:          2,
			Name:          "Volume Test",
			Type:          "HDD",
			ImagePassword: "test1234",
			ImageAlias:    "ubuntu:latest",
		},
	}

	resp := CreateVolume(dcID, request)
	waitTillProvisioned(resp.Headers.Get("Location"))
	return resp.Id
}

func mkipid(name string) string {
	var obj = IpBlock{
		Properties: IpBlockProperties{
			Name:     "GO SDK Test",
			Size:     1,
			Location: "us/las",
		},
	}

	resp := ReserveIpBlock(obj)
	return resp.Id
}

func mksnapshotId(name string, dcId string) string {
	svolumeId := mkVolume(dcId)
	resp := CreateSnapshot(dcId, svolumeId, name, "description")
	waitTillProvisioned(resp.Headers.Get("Location"))
	return resp.Id
}

func mkdcid(name string) string {
	request := Datacenter{
		Properties: DatacenterProperties{
			Name:        name,
			Description: "description",
			Location:    "us/las",
		},
	}
	dc := CreateDatacenter(request)
	return dc.Id
}

func mksrvid(srv_dcid string) string {
	var req = Server{
		Properties: ServerProperties{
			Name:  "GO SDK test",
			Ram:   1024,
			Cores: 2,
		},
	}
	srv := CreateServer(srv_dcid, req)
	waitTillProvisioned(srv.Headers.Get("Location"))
	return srv.Id
}

func mknic(lbal_dcid, serverid string) string {
	var request = Nic{
		Properties: &NicProperties{
			Lan:            1,
			Name:           "GO SDK Test",
			Nat:            false,
			Dhcp:           true,
			FirewallActive: true,
			Ips:            []string{"10.0.0.1"},
		},
	}

	resp := CreateNic(lbal_dcid, serverid, request)
	waitTillProvisioned(resp.Headers.Get("Location"))
	return resp.Id
}

func mknic_custom(lbal_dcid, serverid string, lanid int, ips []string) string {
	var request = Nic{
		Properties: &NicProperties{
			Lan:            lanid,
			Name:           "GO SDK Test",
			Nat:            false,
			Dhcp:           true,
			FirewallActive: true,
			Ips:            ips,
		},
	}

	resp := CreateNic(lbal_dcid, serverid, request)
	waitTillProvisioned(resp.Headers.Get("Location"))
	return resp.Id
}

func waitTillProvisioned(path string) {
	waitCount := 120
	for i := 0; i < waitCount; i++ {
		request := GetRequestStatus(path)
		if request.Metadata.Status == "DONE" {
			break
		}
		time.Sleep(1 * time.Second)
		i++
	}
}

func getImageId(location string, imageName string, imageType string) string {
	if imageName == "" {
		return ""
	}

	SetAuth(os.Getenv("PROFITBRICKS_USERNAME"), os.Getenv("PROFITBRICKS_PASSWORD"))

	images := ListImages()
	if images.StatusCode > 299 {
		fmt.Printf("Error while fetching the list of images %s", images.Response)
	}

	if len(images.Items) > 0 {
		for _, i := range images.Items {
			imgName := ""
			if i.Properties.Name != "" {
				imgName = i.Properties.Name
			}

			if imageType == "SSD" {
				imageType = "HDD"
			}
			if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && i.Properties.ImageType == imageType && i.Properties.Location == location && i.Properties.Public == true {
				return i.Id
			}
		}
	}
	return ""
}
