// are here. Their API is on a path to V2, so just plain JSON is used
// in place of a proper client library for now.

package digitalocean

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type DigitalOceanClientV2 struct {
	// The http client for communicating
	client *http.Client

	// Credentials
	APIToken string

	// The base URL of the API
	APIURL string
}

// Creates a new client for communicating with DO
func DigitalOceanClientNewV2(token string, url string) *DigitalOceanClientV2 {
	c := &DigitalOceanClientV2{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		APIURL:   url,
		APIToken: token,
	}
	return c
}

// Creates an SSH Key and returns it's id
func (d DigitalOceanClientV2) CreateKey(name string, pub string) (uint, error) {
	type KeyReq struct {
		Name      string `json:"name"`
		PublicKey string `json:"public_key"`
	}
	type KeyRes struct {
		SSHKey struct {
			Id          uint
			Name        string
			Fingerprint string
			PublicKey   string `json:"public_key"`
		} `json:"ssh_key"`
	}
	req := &KeyReq{Name: name, PublicKey: pub}
	res := KeyRes{}
	err := NewRequestV2(d, "v2/account/keys", "POST", req, &res)
	if err != nil {
		return 0, err
	}

	return res.SSHKey.Id, err
}

// Destroys an SSH key
func (d DigitalOceanClientV2) DestroyKey(id uint) error {
	path := fmt.Sprintf("v2/account/keys/%v", id)
	return NewRequestV2(d, path, "DELETE", nil, nil)
}

// Creates a droplet and returns it's id
func (d DigitalOceanClientV2) CreateDroplet(name string, size string, image string, region string, keyId uint, privateNetworking bool) (uint, error) {
	type DropletReq struct {
		Name              string   `json:"name"`
		Region            string   `json:"region"`
		Size              string   `json:"size"`
		Image             string   `json:"image"`
		SSHKeys           []string `json:"ssh_keys,omitempty"`
		Backups           bool     `json:"backups,omitempty"`
		IPv6              bool     `json:"ipv6,omitempty"`
		PrivateNetworking bool     `json:"private_networking,omitempty"`
	}
	type DropletRes struct {
		Droplet struct {
			Id       uint
			Name     string
			Memory   uint
			VCPUS    uint `json:"vcpus"`
			Disk     uint
			Region   Region
			Image    Image
			Size     Size
			Locked   bool
			CreateAt string `json:"created_at"`
			Status   string
			Networks struct {
				V4 []struct {
					IPAddr  string `json:"ip_address"`
					Netmask string
					Gateway string
					Type    string
				} `json:"v4,omitempty"`
				V6 []struct {
					IPAddr  string `json:"ip_address"`
					CIDR    uint   `json:"cidr"`
					Gateway string
					Type    string
				} `json:"v6,omitempty"`
			}
			Kernel struct {
				Id      uint
				Name    string
				Version string
			}
			BackupIds   []uint
			SnapshotIds []uint
			ActionIds   []uint
			Features    []string `json:"features,omitempty"`
		}
	}
	req := &DropletReq{Name: name}
	res := DropletRes{}

	found_size, err := d.Size(size)
	if err != nil {
		return 0, fmt.Errorf("Invalid size or lookup failure: '%s': %s", size, err)
	}

	found_image, err := d.Image(image)
	if err != nil {
		return 0, fmt.Errorf("Invalid image or lookup failure: '%s': %s", image, err)
	}

	found_region, err := d.Region(region)
	if err != nil {
		return 0, fmt.Errorf("Invalid region or lookup failure: '%s': %s", region, err)
	}

	req.Size = found_size.Slug
	req.Image = found_image.Slug
	req.Region = found_region.Slug
	req.SSHKeys = []string{fmt.Sprintf("%v", keyId)}
	req.PrivateNetworking = privateNetworking

	err = NewRequestV2(d, "v2/droplets", "POST", req, &res)
	if err != nil {
		return 0, err
	}

	return res.Droplet.Id, err
}

// Destroys a droplet
func (d DigitalOceanClientV2) DestroyDroplet(id uint) error {
	path := fmt.Sprintf("v2/droplets/%v", id)
	return NewRequestV2(d, path, "DELETE", nil, nil)
}

// Powers off a droplet
func (d DigitalOceanClientV2) PowerOffDroplet(id uint) error {
	type ActionReq struct {
		Type string `json:"type"`
	}
	type ActionRes struct {
	}
	req := &ActionReq{Type: "power_off"}
	path := fmt.Sprintf("v2/droplets/%v/actions", id)
	return NewRequestV2(d, path, "POST", req, nil)
}

// Shutsdown a droplet. This is a "soft" shutdown.
func (d DigitalOceanClientV2) ShutdownDroplet(id uint) error {
	type ActionReq struct {
		Type string `json:"type"`
	}
	type ActionRes struct {
	}
	req := &ActionReq{Type: "shutdown"}

	path := fmt.Sprintf("v2/droplets/%v/actions", id)
	return NewRequestV2(d, path, "POST", req, nil)
}

// Creates a snaphot of a droplet by it's ID
func (d DigitalOceanClientV2) CreateSnapshot(id uint, name string) error {
	type ActionReq struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	type ActionRes struct {
	}
	req := &ActionReq{Type: "snapshot", Name: name}
	path := fmt.Sprintf("v2/droplets/%v/actions", id)
	return NewRequestV2(d, path, "POST", req, nil)
}

// Returns all available images.
func (d DigitalOceanClientV2) Images() ([]Image, error) {
	res := ImagesResp{}

	err := NewRequestV2(d, "v2/images?per_page=200", "GET", nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Images, nil
}

// Destroys an image by its ID.
func (d DigitalOceanClientV2) DestroyImage(id uint) error {
	path := fmt.Sprintf("v2/images/%d", id)
	return NewRequestV2(d, path, "DELETE", nil, nil)
}

// Returns DO's string representation of status "off" "new" "active" etc.
func (d DigitalOceanClientV2) DropletStatus(id uint) (string, string, error) {
	path := fmt.Sprintf("v2/droplets/%v", id)
	type DropletRes struct {
		Droplet struct {
			Id       uint
			Name     string
			Memory   uint
			VCPUS    uint `json:"vcpus"`
			Disk     uint
			Region   Region
			Image    Image
			Size     Size
			Locked   bool
			CreateAt string `json:"created_at"`
			Status   string
			Networks struct {
				V4 []struct {
					IPAddr  string `json:"ip_address"`
					Netmask string
					Gateway string
					Type    string
				} `json:"v4,omitempty"`
				V6 []struct {
					IPAddr  string `json:"ip_address"`
					CIDR    uint   `json:"cidr"`
					Gateway string
					Type    string
				} `json:"v6,omitempty"`
			}
			Kernel struct {
				Id      uint
				Name    string
				Version string
			}
			BackupIds   []uint
			SnapshotIds []uint
			ActionIds   []uint
			Features    []string `json:"features,omitempty"`
		}
	}
	res := DropletRes{}
	err := NewRequestV2(d, path, "GET", nil, &res)
	if err != nil {
		return "", "", err
	}
	var ip string

	for _, n := range res.Droplet.Networks.V4 {
		if n.Type == "public" {
			ip = n.IPAddr
		}
	}

	return ip, res.Droplet.Status, err
}

// Sends an api request and returns a generic map[string]interface of
// the response.
func NewRequestV2(d DigitalOceanClientV2, path string, method string, req interface{}, res interface{}) error {
	var err error
	var request *http.Request

	client := d.client

	buf := new(bytes.Buffer)
	// Add the authentication parameters
	url := fmt.Sprintf("%s/%s", d.APIURL, path)
	if req != nil {
		enc := json.NewEncoder(buf)
		enc.Encode(req)
		defer buf.Reset()
		request, err = http.NewRequest(method, url, buf)
		request.Header.Add("Content-Type", "application/json")
	} else {
		request, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return err
	}

	// Add the authentication parameters
	request.Header.Add("Authorization", "Bearer "+d.APIToken)
	if buf != nil {
		log.Printf("sending new request to digitalocean: %s buffer: %s", url, buf)
	} else {
		log.Printf("sending new request to digitalocean: %s", url)
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	if method == "DELETE" && resp.StatusCode == 204 {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return nil
	}

	if resp.Body == nil {
		return errors.New("Request returned empty body")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	log.Printf("response from digitalocean: %s", body)

	err = json.Unmarshal(body, &res)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode JSON response %s (HTTP %v) from DigitalOcean: %s", err.Error(),
			resp.StatusCode, body))
	}
	switch resp.StatusCode {
	case 403, 401, 429, 422, 404, 503, 500:
		return errors.New(fmt.Sprintf("digitalocean request error: %+v", res))
	}
	return nil
}

func (d DigitalOceanClientV2) Image(slug_or_name_or_id string) (Image, error) {
	images, err := d.Images()
	if err != nil {
		return Image{}, err
	}

	for _, image := range images {
		if strings.EqualFold(image.Slug, slug_or_name_or_id) {
			return image, nil
		}
	}

	for _, image := range images {
		if strings.EqualFold(image.Name, slug_or_name_or_id) {
			return image, nil
		}
	}

	for _, image := range images {
		id, err := strconv.Atoi(slug_or_name_or_id)
		if err == nil {
			if image.Id == uint(id) {
				return image, nil
			}
		}
	}

	err = errors.New(fmt.Sprintf("Unknown image '%v'", slug_or_name_or_id))

	return Image{}, err
}

// Returns all available regions.
func (d DigitalOceanClientV2) Regions() ([]Region, error) {
	res := RegionsResp{}
	err := NewRequestV2(d, "v2/regions?per_page=200", "GET", nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Regions, nil
}

func (d DigitalOceanClientV2) Region(slug_or_name_or_id string) (Region, error) {
	regions, err := d.Regions()
	if err != nil {
		return Region{}, err
	}

	for _, region := range regions {
		if strings.EqualFold(region.Slug, slug_or_name_or_id) {
			return region, nil
		}
	}

	for _, region := range regions {
		if strings.EqualFold(region.Name, slug_or_name_or_id) {
			return region, nil
		}
	}

	for _, region := range regions {
		id, err := strconv.Atoi(slug_or_name_or_id)
		if err == nil {
			if region.Id == uint(id) {
				return region, nil
			}
		}
	}

	err = errors.New(fmt.Sprintf("Unknown region '%v'", slug_or_name_or_id))

	return Region{}, err
}

// Returns all available sizes.
func (d DigitalOceanClientV2) Sizes() ([]Size, error) {
	res := SizesResp{}
	err := NewRequestV2(d, "v2/sizes?per_page=200", "GET", nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Sizes, nil
}

func (d DigitalOceanClientV2) Size(slug_or_name_or_id string) (Size, error) {
	sizes, err := d.Sizes()
	if err != nil {
		return Size{}, err
	}

	for _, size := range sizes {
		if strings.EqualFold(size.Slug, slug_or_name_or_id) {
			return size, nil
		}
	}

	for _, size := range sizes {
		if strings.EqualFold(size.Name, slug_or_name_or_id) {
			return size, nil
		}
	}

	for _, size := range sizes {
		id, err := strconv.Atoi(slug_or_name_or_id)
		if err == nil {
			if size.Id == uint(id) {
				return size, nil
			}
		}
	}

	err = errors.New(fmt.Sprintf("Unknown size '%v'", slug_or_name_or_id))

	return Size{}, err
}
