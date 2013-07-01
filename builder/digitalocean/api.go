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
	"strings"
)

const DIGITALOCEAN_API_URL = "https://api.digitalocean.com"

type Image struct {
	Id           uint
	Name         string
	Distribution string
}

type ImagesResp struct {
	Images []Image
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
		client:   http.DefaultClient,
		BaseURL:  DIGITALOCEAN_API_URL,
		ClientID: client,
		APIKey:   key,
	}
	return c
}

// Creates an SSH Key and returns it's id
func (d DigitalOceanClient) CreateKey(name string, pub string) (uint, error) {
	// Escape the public key
	pub = url.QueryEscape(pub)

	params := fmt.Sprintf("name=%v&ssh_pub_key=%v", name, pub)

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
	_, err := NewRequest(d, path, "")
	return err
}

// Creates a droplet and returns it's id
func (d DigitalOceanClient) CreateDroplet(name string, size uint, image uint, region uint, keyId uint) (uint, error) {
	params := fmt.Sprintf(
		"name=%v&image_id=%v&size_id=%v&region_id=%v&ssh_key_ids=%v",
		name, image, size, region, keyId)

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
	_, err := NewRequest(d, path, "")
	return err
}

// Powers off a droplet
func (d DigitalOceanClient) PowerOffDroplet(id uint) error {
	path := fmt.Sprintf("droplets/%v/power_off", id)

	_, err := NewRequest(d, path, "")

	return err
}

// Creates a snaphot of a droplet by it's ID
func (d DigitalOceanClient) CreateSnapshot(id uint, name string) error {
	path := fmt.Sprintf("droplets/%v/snapshot", id)
	params := fmt.Sprintf("name=%v", name)

	_, err := NewRequest(d, path, params)

	return err
}

// Returns all available images.
func (d DigitalOceanClient) Images() ([]Image, error) {
	resp, err := NewRequest(d, "images", "")
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
	_, err := NewRequest(d, path, "")
	return err
}

// Returns DO's string representation of status "off" "new" "active" etc.
func (d DigitalOceanClient) DropletStatus(id uint) (string, string, error) {
	path := fmt.Sprintf("droplets/%v", id)

	body, err := NewRequest(d, path, "")
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
func NewRequest(d DigitalOceanClient, path string, params string) (map[string]interface{}, error) {
	client := d.client
	url := fmt.Sprintf("%s/%s?%s&client_id=%s&api_key=%s",
		DIGITALOCEAN_API_URL, path, params, d.ClientID, d.APIKey)

	var decodedResponse map[string]interface{}

	// Do some basic scrubbing so sensitive information doesn't appear in logs
	scrubbedUrl := strings.Replace(url, d.ClientID, "CLIENT_ID", -1)
	scrubbedUrl = strings.Replace(scrubbedUrl, d.APIKey, "API_KEY", -1)
	log.Printf("sending new request to digitalocean: %s", scrubbedUrl)

	resp, err := client.Get(url)
	if err != nil {
		return decodedResponse, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		return decodedResponse, err
	}

	log.Printf("response from digitalocean: %s", body)

	err = json.Unmarshal(body, &decodedResponse)

	// Check for bad JSON
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to decode JSON response (HTTP %v) from DigitalOcean: %s",
			resp.StatusCode, body))
		return decodedResponse, err
	}

	// Check for errors sent by digitalocean
	status := decodedResponse["status"]
	if status != "OK" {
		// Get the actual error message if there is one
		if status == "ERROR" {
			status = decodedResponse["error_message"]
		}
		err = errors.New(fmt.Sprintf("Received bad response (HTTP %v) from DigitalOcean: %s", resp.StatusCode, status))
		return decodedResponse, err
	}

	return decodedResponse, nil
}
