// All of the methods used to communicate with the digital_ocean API
// are here. Their API is on a path to V2, so just plain JSON is used
// in place of a proper client library for now.

package digitalocean

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DIGITALOCEAN_API_URL = "https://api.digitalocean.com"

type Image struct {
	Id           uint
	Name         string
	Slug         string
	Distribution string
}

type ImagesResp struct {
	Images []Image
}

type Region struct {
	Id   uint
	Name string
	Slug string
}

type RegionsResp struct {
	Regions []Region
}

type Size struct {
	Id   uint
	Name string
	Slug string
}

type SizesResp struct {
	Sizes []Size
}

type DigitalOceanClient struct {
	// The http client for communicating
	client *http.Client

	// The base URL of the API
	BaseURL string

	// Credentials
	ClientID string
	APIKey   string
}

// Creates a new client for communicating with DO
func (d DigitalOceanClient) New(client string, key string) *DigitalOceanClient {
	c := &DigitalOceanClient{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		BaseURL:  DIGITALOCEAN_API_URL,
		ClientID: client,
		APIKey:   key,
	}
	return c
}

// Creates an SSH Key and returns it's id
func (d DigitalOceanClient) CreateKey(name string, pub string) (uint, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("ssh_pub_key", pub)

	body, err := NewRequest(d, "ssh_keys/new", params)
	if err != nil {
		return 0, err
	}

	// Read the SSH key's ID we just created
	key := body["ssh_key"].(map[string]interface{})
	keyId := key["id"].(float64)
	return uint(keyId), nil
}

// Destroys an SSH key
func (d DigitalOceanClient) DestroyKey(id uint) error {
	path := fmt.Sprintf("ssh_keys/%v/destroy", id)
	_, err := NewRequest(d, path, url.Values{})
	return err
}

// Creates a droplet and returns it's id
func (d DigitalOceanClient) CreateDroplet(name string, size string, image string, region string, keyId uint, privateNetworking bool) (uint, error) {
	params := url.Values{}
	params.Set("name", name)

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

	params.Set("size_slug", found_size.Slug)
	params.Set("image_slug", found_image.Slug)
	params.Set("region_slug", found_region.Slug)
	params.Set("ssh_key_ids", fmt.Sprintf("%v", keyId))
	params.Set("private_networking", fmt.Sprintf("%v", privateNetworking))

	body, err := NewRequest(d, "droplets/new", params)
	if err != nil {
		return 0, err
	}

	// Read the Droplets ID
	droplet := body["droplet"].(map[string]interface{})
	dropletId := droplet["id"].(float64)
	return uint(dropletId), err
}

// Destroys a droplet
func (d DigitalOceanClient) DestroyDroplet(id uint) error {
	path := fmt.Sprintf("droplets/%v/destroy", id)
	_, err := NewRequest(d, path, url.Values{})
	return err
}

// Powers off a droplet
func (d DigitalOceanClient) PowerOffDroplet(id uint) error {
	path := fmt.Sprintf("droplets/%v/power_off", id)

	_, err := NewRequest(d, path, url.Values{})

	return err
}

// Shutsdown a droplet. This is a "soft" shutdown.
func (d DigitalOceanClient) ShutdownDroplet(id uint) error {
	path := fmt.Sprintf("droplets/%v/shutdown", id)

	_, err := NewRequest(d, path, url.Values{})

	return err
}

// Creates a snaphot of a droplet by it's ID
func (d DigitalOceanClient) CreateSnapshot(id uint, name string) error {
	path := fmt.Sprintf("droplets/%v/snapshot", id)

	params := url.Values{}
	params.Set("name", name)

	_, err := NewRequest(d, path, params)

	return err
}

// Returns all available images.
func (d DigitalOceanClient) Images() ([]Image, error) {
	resp, err := NewRequest(d, "images", url.Values{})
	if err != nil {
		return nil, err
	}

	var result ImagesResp
	if err := mapstructure.Decode(resp, &result); err != nil {
		return nil, err
	}

	return result.Images, nil
}

// Destroys an image by its ID.
func (d DigitalOceanClient) DestroyImage(id uint) error {
	path := fmt.Sprintf("images/%d/destroy", id)
	_, err := NewRequest(d, path, url.Values{})
	return err
}

// Returns DO's string representation of status "off" "new" "active" etc.
func (d DigitalOceanClient) DropletStatus(id uint) (string, string, error) {
	path := fmt.Sprintf("droplets/%v", id)

	body, err := NewRequest(d, path, url.Values{})
	if err != nil {
		return "", "", err
	}

	var ip string

	// Read the droplet's "status"
	droplet := body["droplet"].(map[string]interface{})
	status := droplet["status"].(string)

	if droplet["ip_address"] != nil {
		ip = droplet["ip_address"].(string)
	}

	return ip, status, err
}

// Sends an api request and returns a generic map[string]interface of
// the response.
func NewRequest(d DigitalOceanClient, path string, params url.Values) (map[string]interface{}, error) {
	client := d.client

	// Add the authentication parameters
	params.Set("client_id", d.ClientID)
	params.Set("api_key", d.APIKey)

	url := fmt.Sprintf("%s/%s?%s", DIGITALOCEAN_API_URL, path, params.Encode())

	// Do some basic scrubbing so sensitive information doesn't appear in logs
	scrubbedUrl := strings.Replace(url, d.ClientID, "CLIENT_ID", -1)
	scrubbedUrl = strings.Replace(scrubbedUrl, d.APIKey, "API_KEY", -1)
	log.Printf("sending new request to digitalocean: %s", scrubbedUrl)

	var lastErr error
	for attempts := 1; attempts < 10; attempts++ {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		log.Printf("response from digitalocean: %s", body)

		var decodedResponse map[string]interface{}
		err = json.Unmarshal(body, &decodedResponse)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to decode JSON response (HTTP %v) from DigitalOcean: %s",
				resp.StatusCode, body))
			return decodedResponse, err
		}

		// Check for errors sent by digitalocean
		status := decodedResponse["status"].(string)
		if status == "OK" {
			return decodedResponse, nil
		}

		if status == "ERROR" {
			statusRaw, ok := decodedResponse["error_message"]
			if ok {
				status = statusRaw.(string)
			} else {
				status = fmt.Sprintf(
					"Unknown error. Full response body: %s", body)
			}
		}

		lastErr = errors.New(fmt.Sprintf("Received error from DigitalOcean (%d): %s",
			resp.StatusCode, status))
		log.Println(lastErr)
		if strings.Contains(status, "a pending event") {
			// Retry, DigitalOcean sends these dumb "pending event"
			// errors all the time.
			time.Sleep(5 * time.Second)
			continue
		}

		// Some other kind of error. Just return.
		return decodedResponse, lastErr
	}

	return nil, lastErr
}

func (d DigitalOceanClient) Image(slug_or_name_or_id string) (Image, error) {
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
func (d DigitalOceanClient) Regions() ([]Region, error) {
	resp, err := NewRequest(d, "regions", url.Values{})
	if err != nil {
		return nil, err
	}

	var result RegionsResp
	if err := mapstructure.Decode(resp, &result); err != nil {
		return nil, err
	}

	return result.Regions, nil
}

func (d DigitalOceanClient) Region(slug_or_name_or_id string) (Region, error) {
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
func (d DigitalOceanClient) Sizes() ([]Size, error) {
	resp, err := NewRequest(d, "sizes", url.Values{})
	if err != nil {
		return nil, err
	}

	var result SizesResp
	if err := mapstructure.Decode(resp, &result); err != nil {
		return nil, err
	}

	return result.Sizes, nil
}

func (d DigitalOceanClient) Size(slug_or_name_or_id string) (Size, error) {
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
