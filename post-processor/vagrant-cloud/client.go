package vagrantcloud

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	commonhelper "github.com/hashicorp/packer/helper/common"
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

func (v VagrantCloudClient) New(baseUrl string, token string, InsecureSkipTLSVerify bool) (*VagrantCloudClient, error) {
	c := &VagrantCloudClient{
		client:      commonhelper.HttpClientWithEnvironmentProxy(),
		BaseURL:     baseUrl,
		AccessToken: token,
	}

	if InsecureSkipTLSVerify {
		transport := c.client.Transport.(*http.Transport)
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return c, c.ValidateAuthentication()
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

func (v *VagrantCloudClient) ValidateAuthentication() error {
	resp, err := v.Get("authenticate")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (v *VagrantCloudClient) Get(path string) (*http.Response, error) {
	reqUrl := fmt.Sprintf("%s/%s", v.BaseURL, path)

	log.Printf("Post-Processor Vagrant Cloud API GET: %s", reqUrl)

	req, err := v.newRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%+v", resp)

	return resp, err
}

func (v *VagrantCloudClient) Delete(path string) (*http.Response, error) {
	reqUrl := fmt.Sprintf("%s/%s", v.BaseURL, path)

	// Scrub API key for logs
	scrubbedUrl := strings.Replace(reqUrl, v.AccessToken, "ACCESS_TOKEN", -1)
	log.Printf("Post-Processor Vagrant Cloud API DELETE: %s", scrubbedUrl)

	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.AccessToken))
	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%+v", resp)

	return resp, err
}

func (v *VagrantCloudClient) Upload(path string, url string) (*http.Response, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("Error opening file for upload: %s", err)
	}

	fi, err := file.Stat()

	if err != nil {
		return nil, fmt.Errorf("Error stating file for upload: %s", err)
	}

	defer file.Close()

	request, err := v.newRequest("PUT", url, file)

	if err != nil {
		return nil, fmt.Errorf("Error preparing upload request: %s", err)
	}

	log.Printf("Post-Processor Vagrant Cloud API Upload: %s %s", path, url)

	request.ContentLength = fi.Size()
	resp, err := v.client.Do(request)

	log.Printf("Post-Processor Vagrant Cloud Upload Response: \n\n%+v", resp)

	return resp, err
}

func (v *VagrantCloudClient) Post(path string, body interface{}) (*http.Response, error) {
	reqUrl := fmt.Sprintf("%s/%s", v.BaseURL, path)

	encBody, err := encodeBody(body)

	if err != nil {
		return nil, fmt.Errorf("Error encoding body for request: %s", err)
	}

	log.Printf("Post-Processor Vagrant Cloud API POST: %s. \n\n Body: %s", reqUrl, encBody)

	req, err := v.newRequest("POST", reqUrl, encBody)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%+v", resp)

	return resp, err
}

func (v *VagrantCloudClient) Put(path string) (*http.Response, error) {
	reqUrl := fmt.Sprintf("%s/%s", v.BaseURL, path)

	log.Printf("Post-Processor Vagrant Cloud API PUT: %s", reqUrl)

	req, err := v.newRequest("PUT", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)

	log.Printf("Post-Processor Vagrant Cloud API Response: \n\n%+v", resp)

	return resp, err
}

func (v *VagrantCloudClient) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.AccessToken))
	return req, err
}
