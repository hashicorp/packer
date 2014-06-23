package vagrantcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type VagrantCloudClient struct {
	// The http client for communicating
	client *http.Client

	// The base URL of the API
	BaseURL string

	// Access token
	AccessToken string
}

type VagrantCloudErrors struct {
	Errors map[string][]string `json:"errors"`
}

func (v VagrantCloudErrors) FormatErrors() string {
	errs := make([]string, 0)
	for e := range v.Errors {
		msg := fmt.Sprintf("%s %s", e, strings.Join(v.Errors[e], ","))
		errs = append(errs, msg)
	}
	return strings.Join(errs, ". ")
}

func (v VagrantCloudClient) New(baseUrl string, token string) *VagrantCloudClient {
	c := &VagrantCloudClient{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		BaseURL:     baseUrl,
		AccessToken: token,
	}
	return c
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

// encodeBody is used to encode a request body
func encodeBody(obj interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return buf, nil
}

func (v VagrantCloudClient) Get(path string) (*http.Response, error) {
	params := url.Values{}
	params.Set("access_token", v.AccessToken)
	reqUrl := fmt.Sprintf("%s/%s?%s", v.BaseURL, path, params.Encode())

	// Scrub API key for logs
	scrubbedUrl := strings.Replace(reqUrl, v.AccessToken, "ACCESS_TOKEN", -1)
	log.Printf("Post-Processor Vagrant Cloud API GET: %s", scrubbedUrl)

	req, err := http.NewRequest("GET", reqUrl, nil)
	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%s", resp)

	return resp, err
}

func (v VagrantCloudClient) Delete(path string) (*http.Response, error) {
	params := url.Values{}
	params.Set("access_token", v.AccessToken)
	reqUrl := fmt.Sprintf("%s/%s?%s", v.BaseURL, path, params.Encode())

	// Scrub API key for logs
	scrubbedUrl := strings.Replace(reqUrl, v.AccessToken, "ACCESS_TOKEN", -1)
	log.Printf("Post-Processor Vagrant Cloud API DELETE: %s", scrubbedUrl)

	req, err := http.NewRequest("DELETE", reqUrl, nil)
	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%s", resp)

	return resp, err
}

func (v VagrantCloudClient) Post(path string, body interface{}) (*http.Response, error) {
	params := url.Values{}
	params.Set("access_token", v.AccessToken)
	reqUrl := fmt.Sprintf("%s/%s?%s", v.BaseURL, path, params.Encode())

	log.Println(reqUrl)

	encBody, err := encodeBody(body)

	log.Println(encBody)

	if err != nil {
		return nil, fmt.Errorf("Error encoding body for request: %s", err)
	}

	// Scrub API key for logs
	scrubbedUrl := strings.Replace(reqUrl, v.AccessToken, "ACCESS_TOKEN", -1)
	log.Printf("Post-Processor Vagrant Cloud API POST: %s. \n\n Body: %s", scrubbedUrl, encBody)

	req, err := http.NewRequest("POST", reqUrl, encBody)
	req.Header.Add("Content-Type", "application/json")

	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%s", resp)

	return resp, err
}
