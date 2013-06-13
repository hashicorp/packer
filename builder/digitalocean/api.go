// All of the methods used to communicate with the digital_ocean API
// are here. Their API is on a path to V2, so just plain JSON is used
// in place of a proper client library for now.

package digitalocean

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const DIGITALOCEAN_API_URL = "https://api.digitalocean.com"

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
	params := fmt.Sprintf("?name=%s&ssh_pub_key=%s", name, pub)

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
	path := fmt.Sprintf("ssh_keys/%s/destroy", id)
	_, err := NewRequest(d, path, "")
	return err
}

// Creates a droplet and returns it's id
func (d DigitalOceanClient) CreateDroplet(name string, size uint, image uint, region uint, keyId uint) (uint, error) {
	params := fmt.Sprintf(
		"name=%s&size_id=%s&image_id=%s&size_id=%s&image_id=%s&region_id=%s&ssh_key_ids=%s",
		name, size, image, size, region, keyId)

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
	path := fmt.Sprintf("droplets/%s/destroy", id)
	_, err := NewRequest(d, path, "")
	return err
}

// Powers off a droplet
func (d DigitalOceanClient) PowerOffDroplet(id uint) error {
	path := fmt.Sprintf("droplets/%s/power_off", id)

	_, err := NewRequest(d, path, "")

	return err
}

// Creates a snaphot of a droplet by it's ID
func (d DigitalOceanClient) CreateSnapshot(id uint, name string) error {
	path := fmt.Sprintf("droplets/%s/snapshot", id)
	params := fmt.Sprintf("name=%s", name)

	_, err := NewRequest(d, path, params)

	return err
}

// Returns DO's string representation of status "off" "new" "active" etc.
func (d DigitalOceanClient) DropletStatus(id uint) (string, error) {
	path := fmt.Sprintf("droplets/%s", id)

	body, err := NewRequest(d, path, "")
	if err != nil {
		return "", err
	}

	// Read the droplet's "status"
	droplet := body["droplet"].(map[string]interface{})
	status := droplet["status"].(string)

	return status, err
}

// Sends an api request and returns a generic map[string]interface of
// the response.
func NewRequest(d DigitalOceanClient, path string, params string) (map[string]interface{}, error) {
	client := d.client
	url := fmt.Sprintf("%s/%s?%s&client_id=%s&api_key=%s",
		DIGITALOCEAN_API_URL, path, params, d.ClientID, d.APIKey)

	var decodedResponse map[string]interface{}

	resp, err := client.Get(url)
	if err != nil {
		return decodedResponse, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		return decodedResponse, err
	}

	// Catch all non-200 status and return an error
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("recieved non-200 status from digitalocean: %d", resp.StatusCode))
		return decodedResponse, err
	}

	err = json.Unmarshal(body, &decodedResponse)

	if err != nil {
		return decodedResponse, err
	}

	// Catch all non-OK statuses from DO and return an error
	status := decodedResponse["status"]
	if status != "OK" {
		err = errors.New(fmt.Sprintf("recieved non-OK status from digitalocean: %d", status))
		return decodedResponse, err
	}

	return decodedResponse, nil
}
