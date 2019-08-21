// GENERATED FILE: DO NOT EDIT!

package oapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/outscale/osc-go/utils"
)

type Client struct {
	service string

	signer *v4.Signer

	client *http.Client

	config *Config
}

type Config struct {
	AccessKey string
	SecretKey string
	Region    string
	URL       string

	//Only Used for OAPI
	Service string

	// User agent for client
	UserAgent string
}

func (c Config) ServiceURL() string {
	s := fmt.Sprintf("https://%s.%s.%s", c.Service, c.Region, c.URL)

	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u.String()
}

// NewClient creates an API client.
func NewClient(config *Config, c *http.Client) *Client {
	client := &Client{}
	client.service = config.ServiceURL()
	if c != nil {
		client.client = c
	} else {
		client.client = http.DefaultClient
	}

	s := &v4.Signer{
		Credentials: credentials.NewStaticCredentials(config.AccessKey,
			config.SecretKey, ""),
	}

	client.signer = s
	client.config = config

	return client
}

func (c *Client) GetConfig() *Config {
	return c.config
}

// Sign ...
func (c *Client) Sign(req *http.Request, body []byte) error {
	reader := strings.NewReader(string(body))
	timestamp := time.Now()
	_, err := c.signer.Sign(req, reader, "oapi", c.config.Region, timestamp)
	utils.DebugRequest(req)
	return err

}

// Do ...
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[DEBUG] Error in Do Request %s", err)
	}

	if resp != nil {
		utils.DebugResponse(resp)
	} else {
		log.Println("[DEBUG] No response to show.")
	}

	return resp, err
}

//
func (client *Client) POST_AcceptNetPeering(
	acceptnetpeeringrequest AcceptNetPeeringRequest,
) (
	response *POST_AcceptNetPeeringResponses,
	err error,
) {
	path := client.service + "/AcceptNetPeering"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(acceptnetpeeringrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_AcceptNetPeeringResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &AcceptNetPeeringResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_AuthenticateAccount(
	authenticateaccountrequest AuthenticateAccountRequest,
) (
	response *POST_AuthenticateAccountResponses,
	err error,
) {
	path := client.service + "/AuthenticateAccount"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(authenticateaccountrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_AuthenticateAccountResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &AuthenticateAccountResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CheckSignature(
	checksignaturerequest CheckSignatureRequest,
) (
	response *POST_CheckSignatureResponses,
	err error,
) {
	path := client.service + "/CheckSignature"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(checksignaturerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CheckSignatureResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CheckSignatureResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CopyAccount(
	copyaccountrequest CopyAccountRequest,
) (
	response *POST_CopyAccountResponses,
	err error,
) {
	path := client.service + "/CopyAccount"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(copyaccountrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CopyAccountResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CopyAccountResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateAccount(
	createaccountrequest CreateAccountRequest,
) (
	response *POST_CreateAccountResponses,
	err error,
) {
	path := client.service + "/CreateAccount"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createaccountrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateAccountResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateAccountResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateApiKey(
	createapikeyrequest CreateApiKeyRequest,
) (
	response *POST_CreateApiKeyResponses,
	err error,
) {
	path := client.service + "/CreateApiKey"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createapikeyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateApiKeyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateApiKeyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateClientGateway(
	createclientgatewayrequest CreateClientGatewayRequest,
) (
	response *POST_CreateClientGatewayResponses,
	err error,
) {
	path := client.service + "/CreateClientGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createclientgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateClientGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateClientGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateDhcpOptions(
	createdhcpoptionsrequest CreateDhcpOptionsRequest,
) (
	response *POST_CreateDhcpOptionsResponses,
	err error,
) {
	path := client.service + "/CreateDhcpOptions"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createdhcpoptionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateDhcpOptionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateDhcpOptionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateDirectLink(
	createdirectlinkrequest CreateDirectLinkRequest,
) (
	response *POST_CreateDirectLinkResponses,
	err error,
) {
	path := client.service + "/CreateDirectLink"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createdirectlinkrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateDirectLinkResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateDirectLinkResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateDirectLinkInterface(
	createdirectlinkinterfacerequest CreateDirectLinkInterfaceRequest,
) (
	response *POST_CreateDirectLinkInterfaceResponses,
	err error,
) {
	path := client.service + "/CreateDirectLinkInterface"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createdirectlinkinterfacerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateDirectLinkInterfaceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateDirectLinkInterfaceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateImage(
	createimagerequest CreateImageRequest,
) (
	response *POST_CreateImageResponses,
	err error,
) {
	path := client.service + "/CreateImage"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createimagerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateImageResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateImageResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateImageExportTask(
	createimageexporttaskrequest CreateImageExportTaskRequest,
) (
	response *POST_CreateImageExportTaskResponses,
	err error,
) {
	path := client.service + "/CreateImageExportTask"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createimageexporttaskrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateImageExportTaskResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateImageExportTaskResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateInternetService(
	createinternetservicerequest CreateInternetServiceRequest,
) (
	response *POST_CreateInternetServiceResponses,
	err error,
) {
	path := client.service + "/CreateInternetService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createinternetservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateInternetServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateInternetServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateKeypair(
	createkeypairrequest CreateKeypairRequest,
) (
	response *POST_CreateKeypairResponses,
	err error,
) {
	path := client.service + "/CreateKeypair"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createkeypairrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateKeypairResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateKeypairResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateListenerRule(
	createlistenerrulerequest CreateListenerRuleRequest,
) (
	response *POST_CreateListenerRuleResponses,
	err error,
) {
	path := client.service + "/CreateListenerRule"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createlistenerrulerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateListenerRuleResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateListenerRuleResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateLoadBalancer(
	createloadbalancerrequest CreateLoadBalancerRequest,
) (
	response *POST_CreateLoadBalancerResponses,
	err error,
) {
	path := client.service + "/CreateLoadBalancer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createloadbalancerrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateLoadBalancerResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateLoadBalancerResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateLoadBalancerListeners(
	createloadbalancerlistenersrequest CreateLoadBalancerListenersRequest,
) (
	response *POST_CreateLoadBalancerListenersResponses,
	err error,
) {
	path := client.service + "/CreateLoadBalancerListeners"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createloadbalancerlistenersrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateLoadBalancerListenersResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateLoadBalancerListenersResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateLoadBalancerPolicy(
	createloadbalancerpolicyrequest CreateLoadBalancerPolicyRequest,
) (
	response *POST_CreateLoadBalancerPolicyResponses,
	err error,
) {
	path := client.service + "/CreateLoadBalancerPolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createloadbalancerpolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateLoadBalancerPolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateLoadBalancerPolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateNatService(
	createnatservicerequest CreateNatServiceRequest,
) (
	response *POST_CreateNatServiceResponses,
	err error,
) {
	path := client.service + "/CreateNatService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createnatservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateNatServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateNatServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateNet(
	createnetrequest CreateNetRequest,
) (
	response *POST_CreateNetResponses,
	err error,
) {
	path := client.service + "/CreateNet"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createnetrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateNetResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateNetResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateNetAccessPoint(
	createnetaccesspointrequest CreateNetAccessPointRequest,
) (
	response *POST_CreateNetAccessPointResponses,
	err error,
) {
	path := client.service + "/CreateNetAccessPoint"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createnetaccesspointrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateNetAccessPointResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateNetAccessPointResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateNetPeering(
	createnetpeeringrequest CreateNetPeeringRequest,
) (
	response *POST_CreateNetPeeringResponses,
	err error,
) {
	path := client.service + "/CreateNetPeering"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createnetpeeringrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateNetPeeringResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateNetPeeringResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateNic(
	createnicrequest CreateNicRequest,
) (
	response *POST_CreateNicResponses,
	err error,
) {
	path := client.service + "/CreateNic"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createnicrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateNicResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateNicResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreatePolicy(
	createpolicyrequest CreatePolicyRequest,
) (
	response *POST_CreatePolicyResponses,
	err error,
) {
	path := client.service + "/CreatePolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createpolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreatePolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreatePolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreatePublicIp(
	createpubliciprequest CreatePublicIpRequest,
) (
	response *POST_CreatePublicIpResponses,
	err error,
) {
	path := client.service + "/CreatePublicIp"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createpubliciprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreatePublicIpResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreatePublicIpResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateRoute(
	createrouterequest CreateRouteRequest,
) (
	response *POST_CreateRouteResponses,
	err error,
) {
	path := client.service + "/CreateRoute"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createrouterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateRouteResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateRouteResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateRouteTable(
	createroutetablerequest CreateRouteTableRequest,
) (
	response *POST_CreateRouteTableResponses,
	err error,
) {
	path := client.service + "/CreateRouteTable"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createroutetablerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateRouteTableResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateRouteTableResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateSecurityGroup(
	createsecuritygrouprequest CreateSecurityGroupRequest,
) (
	response *POST_CreateSecurityGroupResponses,
	err error,
) {
	path := client.service + "/CreateSecurityGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createsecuritygrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateSecurityGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateSecurityGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateSecurityGroupRule(
	createsecuritygrouprulerequest CreateSecurityGroupRuleRequest,
) (
	response *POST_CreateSecurityGroupRuleResponses,
	err error,
) {
	path := client.service + "/CreateSecurityGroupRule"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createsecuritygrouprulerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateSecurityGroupRuleResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateSecurityGroupRuleResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateServerCertificate(
	createservercertificaterequest CreateServerCertificateRequest,
) (
	response *POST_CreateServerCertificateResponses,
	err error,
) {
	path := client.service + "/CreateServerCertificate"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createservercertificaterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateServerCertificateResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateServerCertificateResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateSnapshot(
	createsnapshotrequest CreateSnapshotRequest,
) (
	response *POST_CreateSnapshotResponses,
	err error,
) {
	path := client.service + "/CreateSnapshot"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createsnapshotrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateSnapshotResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateSnapshotResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateSnapshotExportTask(
	createsnapshotexporttaskrequest CreateSnapshotExportTaskRequest,
) (
	response *POST_CreateSnapshotExportTaskResponses,
	err error,
) {
	path := client.service + "/CreateSnapshotExportTask"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createsnapshotexporttaskrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateSnapshotExportTaskResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateSnapshotExportTaskResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateSubnet(
	createsubnetrequest CreateSubnetRequest,
) (
	response *POST_CreateSubnetResponses,
	err error,
) {
	path := client.service + "/CreateSubnet"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createsubnetrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateSubnetResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateSubnetResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateTags(
	createtagsrequest CreateTagsRequest,
) (
	response *POST_CreateTagsResponses,
	err error,
) {
	path := client.service + "/CreateTags"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createtagsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateTagsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateTagsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateUser(
	createuserrequest CreateUserRequest,
) (
	response *POST_CreateUserResponses,
	err error,
) {
	path := client.service + "/CreateUser"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createuserrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateUserResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateUserResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateUserGroup(
	createusergrouprequest CreateUserGroupRequest,
) (
	response *POST_CreateUserGroupResponses,
	err error,
) {
	path := client.service + "/CreateUserGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createusergrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateUserGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateUserGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateVirtualGateway(
	createvirtualgatewayrequest CreateVirtualGatewayRequest,
) (
	response *POST_CreateVirtualGatewayResponses,
	err error,
) {
	path := client.service + "/CreateVirtualGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createvirtualgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateVirtualGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateVirtualGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateVms(
	createvmsrequest CreateVmsRequest,
) (
	response *POST_CreateVmsResponses,
	err error,
) {
	path := client.service + "/CreateVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateVolume(
	createvolumerequest CreateVolumeRequest,
) (
	response *POST_CreateVolumeResponses,
	err error,
) {
	path := client.service + "/CreateVolume"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createvolumerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateVolumeResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateVolumeResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		return nil, checkErrorResponse(resp)
	}
	return
}

//
func (client *Client) POST_CreateVpnConnection(
	createvpnconnectionrequest CreateVpnConnectionRequest,
) (
	response *POST_CreateVpnConnectionResponses,
	err error,
) {
	path := client.service + "/CreateVpnConnection"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createvpnconnectionrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateVpnConnectionResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateVpnConnectionResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_CreateVpnConnectionRoute(
	createvpnconnectionrouterequest CreateVpnConnectionRouteRequest,
) (
	response *POST_CreateVpnConnectionRouteResponses,
	err error,
) {
	path := client.service + "/CreateVpnConnectionRoute"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(createvpnconnectionrouterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_CreateVpnConnectionRouteResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &CreateVpnConnectionRouteResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteApiKey(
	deleteapikeyrequest DeleteApiKeyRequest,
) (
	response *POST_DeleteApiKeyResponses,
	err error,
) {
	path := client.service + "/DeleteApiKey"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteapikeyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteApiKeyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteApiKeyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteClientGateway(
	deleteclientgatewayrequest DeleteClientGatewayRequest,
) (
	response *POST_DeleteClientGatewayResponses,
	err error,
) {
	path := client.service + "/DeleteClientGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteclientgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteClientGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteClientGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteDhcpOptions(
	deletedhcpoptionsrequest DeleteDhcpOptionsRequest,
) (
	response *POST_DeleteDhcpOptionsResponses,
	err error,
) {
	path := client.service + "/DeleteDhcpOptions"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletedhcpoptionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteDhcpOptionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteDhcpOptionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteDirectLink(
	deletedirectlinkrequest DeleteDirectLinkRequest,
) (
	response *POST_DeleteDirectLinkResponses,
	err error,
) {
	path := client.service + "/DeleteDirectLink"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletedirectlinkrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteDirectLinkResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteDirectLinkResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteDirectLinkInterface(
	deletedirectlinkinterfacerequest DeleteDirectLinkInterfaceRequest,
) (
	response *POST_DeleteDirectLinkInterfaceResponses,
	err error,
) {
	path := client.service + "/DeleteDirectLinkInterface"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletedirectlinkinterfacerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteDirectLinkInterfaceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteDirectLinkInterfaceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteExportTask(
	deleteexporttaskrequest DeleteExportTaskRequest,
) (
	response *POST_DeleteExportTaskResponses,
	err error,
) {
	path := client.service + "/DeleteExportTask"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteexporttaskrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteExportTaskResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteExportTaskResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteImage(
	deleteimagerequest DeleteImageRequest,
) (
	response *POST_DeleteImageResponses,
	err error,
) {
	path := client.service + "/DeleteImage"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteimagerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteImageResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteImageResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteInternetService(
	deleteinternetservicerequest DeleteInternetServiceRequest,
) (
	response *POST_DeleteInternetServiceResponses,
	err error,
) {
	path := client.service + "/DeleteInternetService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteinternetservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteInternetServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteInternetServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteKeypair(
	deletekeypairrequest DeleteKeypairRequest,
) (
	response *POST_DeleteKeypairResponses,
	err error,
) {
	path := client.service + "/DeleteKeypair"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletekeypairrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteKeypairResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteKeypairResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteListenerRule(
	deletelistenerrulerequest DeleteListenerRuleRequest,
) (
	response *POST_DeleteListenerRuleResponses,
	err error,
) {
	path := client.service + "/DeleteListenerRule"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletelistenerrulerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteListenerRuleResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteListenerRuleResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteLoadBalancer(
	deleteloadbalancerrequest DeleteLoadBalancerRequest,
) (
	response *POST_DeleteLoadBalancerResponses,
	err error,
) {
	path := client.service + "/DeleteLoadBalancer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteloadbalancerrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteLoadBalancerResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteLoadBalancerResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteLoadBalancerListeners(
	deleteloadbalancerlistenersrequest DeleteLoadBalancerListenersRequest,
) (
	response *POST_DeleteLoadBalancerListenersResponses,
	err error,
) {
	path := client.service + "/DeleteLoadBalancerListeners"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteloadbalancerlistenersrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteLoadBalancerListenersResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteLoadBalancerListenersResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteLoadBalancerPolicy(
	deleteloadbalancerpolicyrequest DeleteLoadBalancerPolicyRequest,
) (
	response *POST_DeleteLoadBalancerPolicyResponses,
	err error,
) {
	path := client.service + "/DeleteLoadBalancerPolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteloadbalancerpolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteLoadBalancerPolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteLoadBalancerPolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteNatService(
	deletenatservicerequest DeleteNatServiceRequest,
) (
	response *POST_DeleteNatServiceResponses,
	err error,
) {
	path := client.service + "/DeleteNatService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletenatservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteNatServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteNatServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteNet(
	deletenetrequest DeleteNetRequest,
) (
	response *POST_DeleteNetResponses,
	err error,
) {
	path := client.service + "/DeleteNet"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletenetrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteNetResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteNetResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteNetAccessPoints(
	deletenetaccesspointsrequest DeleteNetAccessPointsRequest,
) (
	response *POST_DeleteNetAccessPointsResponses,
	err error,
) {
	path := client.service + "/DeleteNetAccessPoints"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletenetaccesspointsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteNetAccessPointsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteNetAccessPointsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteNetPeering(
	deletenetpeeringrequest DeleteNetPeeringRequest,
) (
	response *POST_DeleteNetPeeringResponses,
	err error,
) {
	path := client.service + "/DeleteNetPeering"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletenetpeeringrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteNetPeeringResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteNetPeeringResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteNic(
	deletenicrequest DeleteNicRequest,
) (
	response *POST_DeleteNicResponses,
	err error,
) {
	path := client.service + "/DeleteNic"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletenicrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteNicResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteNicResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeletePolicy(
	deletepolicyrequest DeletePolicyRequest,
) (
	response *POST_DeletePolicyResponses,
	err error,
) {
	path := client.service + "/DeletePolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletepolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeletePolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeletePolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeletePublicIp(
	deletepubliciprequest DeletePublicIpRequest,
) (
	response *POST_DeletePublicIpResponses,
	err error,
) {
	path := client.service + "/DeletePublicIp"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletepubliciprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeletePublicIpResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeletePublicIpResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteRoute(
	deleterouterequest DeleteRouteRequest,
) (
	response *POST_DeleteRouteResponses,
	err error,
) {
	path := client.service + "/DeleteRoute"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleterouterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteRouteResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteRouteResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteRouteTable(
	deleteroutetablerequest DeleteRouteTableRequest,
) (
	response *POST_DeleteRouteTableResponses,
	err error,
) {
	path := client.service + "/DeleteRouteTable"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteroutetablerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteRouteTableResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteRouteTableResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteSecurityGroup(
	deletesecuritygrouprequest DeleteSecurityGroupRequest,
) (
	response *POST_DeleteSecurityGroupResponses,
	err error,
) {
	path := client.service + "/DeleteSecurityGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletesecuritygrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteSecurityGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteSecurityGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteSecurityGroupRule(
	deletesecuritygrouprulerequest DeleteSecurityGroupRuleRequest,
) (
	response *POST_DeleteSecurityGroupRuleResponses,
	err error,
) {
	path := client.service + "/DeleteSecurityGroupRule"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletesecuritygrouprulerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteSecurityGroupRuleResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteSecurityGroupRuleResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteServerCertificate(
	deleteservercertificaterequest DeleteServerCertificateRequest,
) (
	response *POST_DeleteServerCertificateResponses,
	err error,
) {
	path := client.service + "/DeleteServerCertificate"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteservercertificaterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteServerCertificateResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteServerCertificateResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteSnapshot(
	deletesnapshotrequest DeleteSnapshotRequest,
) (
	response *POST_DeleteSnapshotResponses,
	err error,
) {
	path := client.service + "/DeleteSnapshot"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletesnapshotrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteSnapshotResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteSnapshotResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteSubnet(
	deletesubnetrequest DeleteSubnetRequest,
) (
	response *POST_DeleteSubnetResponses,
	err error,
) {
	path := client.service + "/DeleteSubnet"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletesubnetrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteSubnetResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteSubnetResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteTags(
	deletetagsrequest DeleteTagsRequest,
) (
	response *POST_DeleteTagsResponses,
	err error,
) {
	path := client.service + "/DeleteTags"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletetagsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteTagsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteTagsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteUser(
	deleteuserrequest DeleteUserRequest,
) (
	response *POST_DeleteUserResponses,
	err error,
) {
	path := client.service + "/DeleteUser"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteuserrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteUserResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteUserResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteUserGroup(
	deleteusergrouprequest DeleteUserGroupRequest,
) (
	response *POST_DeleteUserGroupResponses,
	err error,
) {
	path := client.service + "/DeleteUserGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deleteusergrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteUserGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteUserGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteVirtualGateway(
	deletevirtualgatewayrequest DeleteVirtualGatewayRequest,
) (
	response *POST_DeleteVirtualGatewayResponses,
	err error,
) {
	path := client.service + "/DeleteVirtualGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletevirtualgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteVirtualGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteVirtualGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteVms(
	deletevmsrequest DeleteVmsRequest,
) (
	response *POST_DeleteVmsResponses,
	err error,
) {
	path := client.service + "/DeleteVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletevmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteVolume(
	deletevolumerequest DeleteVolumeRequest,
) (
	response *POST_DeleteVolumeResponses,
	err error,
) {
	path := client.service + "/DeleteVolume"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletevolumerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteVolumeResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteVolumeResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteVpnConnection(
	deletevpnconnectionrequest DeleteVpnConnectionRequest,
) (
	response *POST_DeleteVpnConnectionResponses,
	err error,
) {
	path := client.service + "/DeleteVpnConnection"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletevpnconnectionrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteVpnConnectionResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteVpnConnectionResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeleteVpnConnectionRoute(
	deletevpnconnectionrouterequest DeleteVpnConnectionRouteRequest,
) (
	response *POST_DeleteVpnConnectionRouteResponses,
	err error,
) {
	path := client.service + "/DeleteVpnConnectionRoute"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deletevpnconnectionrouterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeleteVpnConnectionRouteResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeleteVpnConnectionRouteResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeregisterUserInUserGroup(
	deregisteruserinusergrouprequest DeregisterUserInUserGroupRequest,
) (
	response *POST_DeregisterUserInUserGroupResponses,
	err error,
) {
	path := client.service + "/DeregisterUserInUserGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deregisteruserinusergrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeregisterUserInUserGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeregisterUserInUserGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_DeregisterVmsInLoadBalancer(
	deregistervmsinloadbalancerrequest DeregisterVmsInLoadBalancerRequest,
) (
	response *POST_DeregisterVmsInLoadBalancerResponses,
	err error,
) {
	path := client.service + "/DeregisterVmsInLoadBalancer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(deregistervmsinloadbalancerrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_DeregisterVmsInLoadBalancerResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &DeregisterVmsInLoadBalancerResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkInternetService(
	linkinternetservicerequest LinkInternetServiceRequest,
) (
	response *POST_LinkInternetServiceResponses,
	err error,
) {
	path := client.service + "/LinkInternetService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkinternetservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkInternetServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkInternetServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkNic(
	linknicrequest LinkNicRequest,
) (
	response *POST_LinkNicResponses,
	err error,
) {
	path := client.service + "/LinkNic"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linknicrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkNicResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkNicResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkPolicy(
	linkpolicyrequest LinkPolicyRequest,
) (
	response *POST_LinkPolicyResponses,
	err error,
) {
	path := client.service + "/LinkPolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkpolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkPolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkPolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkPrivateIps(
	linkprivateipsrequest LinkPrivateIpsRequest,
) (
	response *POST_LinkPrivateIpsResponses,
	err error,
) {
	path := client.service + "/LinkPrivateIps"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkprivateipsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkPrivateIpsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkPrivateIpsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkPublicIp(
	linkpubliciprequest LinkPublicIpRequest,
) (
	response *POST_LinkPublicIpResponses,
	err error,
) {
	path := client.service + "/LinkPublicIp"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkpubliciprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkPublicIpResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkPublicIpResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkRouteTable(
	linkroutetablerequest LinkRouteTableRequest,
) (
	response *POST_LinkRouteTableResponses,
	err error,
) {
	path := client.service + "/LinkRouteTable"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkroutetablerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkRouteTableResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkRouteTableResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkVirtualGateway(
	linkvirtualgatewayrequest LinkVirtualGatewayRequest,
) (
	response *POST_LinkVirtualGatewayResponses,
	err error,
) {
	path := client.service + "/LinkVirtualGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkvirtualgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkVirtualGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkVirtualGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_LinkVolume(
	linkvolumerequest LinkVolumeRequest,
) (
	response *POST_LinkVolumeResponses,
	err error,
) {
	path := client.service + "/LinkVolume"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(linkvolumerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_LinkVolumeResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &LinkVolumeResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_PurchaseReservedVmsOffer(
	purchasereservedvmsofferrequest PurchaseReservedVmsOfferRequest,
) (
	response *POST_PurchaseReservedVmsOfferResponses,
	err error,
) {
	path := client.service + "/PurchaseReservedVmsOffer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(purchasereservedvmsofferrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_PurchaseReservedVmsOfferResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &PurchaseReservedVmsOfferResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadAccount(
	readaccountrequest ReadAccountRequest,
) (
	response *POST_ReadAccountResponses,
	err error,
) {
	path := client.service + "/ReadAccount"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readaccountrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadAccountResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadAccountResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadAccountConsumption(
	readaccountconsumptionrequest ReadAccountConsumptionRequest,
) (
	response *POST_ReadAccountConsumptionResponses,
	err error,
) {
	path := client.service + "/ReadAccountConsumption"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readaccountconsumptionrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadAccountConsumptionResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadAccountConsumptionResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadAdminPassword(
	readadminpasswordrequest ReadAdminPasswordRequest,
) (
	response *POST_ReadAdminPasswordResponses,
	err error,
) {
	path := client.service + "/ReadAdminPassword"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readadminpasswordrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadAdminPasswordResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadAdminPasswordResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadApiKeys(
	readapikeysrequest ReadApiKeysRequest,
) (
	response *POST_ReadApiKeysResponses,
	err error,
) {
	path := client.service + "/ReadApiKeys"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readapikeysrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadApiKeysResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadApiKeysResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadApiLogs(
	readapilogsrequest ReadApiLogsRequest,
) (
	response *POST_ReadApiLogsResponses,
	err error,
) {
	path := client.service + "/ReadApiLogs"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readapilogsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadApiLogsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadApiLogsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadBillableDigest(
	readbillabledigestrequest ReadBillableDigestRequest,
) (
	response *POST_ReadBillableDigestResponses,
	err error,
) {
	path := client.service + "/ReadBillableDigest"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readbillabledigestrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadBillableDigestResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadBillableDigestResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadCatalog(
	readcatalogrequest ReadCatalogRequest,
) (
	response *POST_ReadCatalogResponses,
	err error,
) {
	path := client.service + "/ReadCatalog"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readcatalogrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadCatalogResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadCatalogResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadClientGateways(
	readclientgatewaysrequest ReadClientGatewaysRequest,
) (
	response *POST_ReadClientGatewaysResponses,
	err error,
) {
	path := client.service + "/ReadClientGateways"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readclientgatewaysrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadClientGatewaysResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadClientGatewaysResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadConsoleOutput(
	readconsoleoutputrequest ReadConsoleOutputRequest,
) (
	response *POST_ReadConsoleOutputResponses,
	err error,
) {
	path := client.service + "/ReadConsoleOutput"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readconsoleoutputrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadConsoleOutputResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadConsoleOutputResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadDhcpOptions(
	readdhcpoptionsrequest ReadDhcpOptionsRequest,
) (
	response *POST_ReadDhcpOptionsResponses,
	err error,
) {
	path := client.service + "/ReadDhcpOptions"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readdhcpoptionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadDhcpOptionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadDhcpOptionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadDirectLinkInterfaces(
	readdirectlinkinterfacesrequest ReadDirectLinkInterfacesRequest,
) (
	response *POST_ReadDirectLinkInterfacesResponses,
	err error,
) {
	path := client.service + "/ReadDirectLinkInterfaces"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readdirectlinkinterfacesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadDirectLinkInterfacesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadDirectLinkInterfacesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadDirectLinks(
	readdirectlinksrequest ReadDirectLinksRequest,
) (
	response *POST_ReadDirectLinksResponses,
	err error,
) {
	path := client.service + "/ReadDirectLinks"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readdirectlinksrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadDirectLinksResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadDirectLinksResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadImageExportTasks(
	readimageexporttasksrequest ReadImageExportTasksRequest,
) (
	response *POST_ReadImageExportTasksResponses,
	err error,
) {
	path := client.service + "/ReadImageExportTasks"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readimageexporttasksrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadImageExportTasksResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadImageExportTasksResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadImages(
	readimagesrequest ReadImagesRequest,
) (
	response *POST_ReadImagesResponses,
	err error,
) {
	path := client.service + "/ReadImages"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readimagesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadImagesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadImagesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadInternetServices(
	readinternetservicesrequest ReadInternetServicesRequest,
) (
	response *POST_ReadInternetServicesResponses,
	err error,
) {
	path := client.service + "/ReadInternetServices"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readinternetservicesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadInternetServicesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadInternetServicesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadKeypairs(
	readkeypairsrequest ReadKeypairsRequest,
) (
	response *POST_ReadKeypairsResponses,
	err error,
) {
	path := client.service + "/ReadKeypairs"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readkeypairsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadKeypairsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadKeypairsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadListenerRules(
	readlistenerrulesrequest ReadListenerRulesRequest,
) (
	response *POST_ReadListenerRulesResponses,
	err error,
) {
	path := client.service + "/ReadListenerRules"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readlistenerrulesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	utils.DebugRequest(req)
	resp, err := client.Do(req)
	if resp != nil {
		utils.DebugResponse(resp)
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadListenerRulesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadListenerRulesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadLoadBalancers(
	readloadbalancersrequest ReadLoadBalancersRequest,
) (
	response *POST_ReadLoadBalancersResponses,
	err error,
) {
	path := client.service + "/ReadLoadBalancers"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readloadbalancersrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadLoadBalancersResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadLoadBalancersResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadLocations(
	readlocationsrequest ReadLocationsRequest,
) (
	response *POST_ReadLocationsResponses,
	err error,
) {
	path := client.service + "/ReadLocations"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readlocationsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadLocationsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadLocationsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNatServices(
	readnatservicesrequest ReadNatServicesRequest,
) (
	response *POST_ReadNatServicesResponses,
	err error,
) {
	path := client.service + "/ReadNatServices"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnatservicesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNatServicesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNatServicesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNetAccessPointServices(
	readnetaccesspointservicesrequest ReadNetAccessPointServicesRequest,
) (
	response *POST_ReadNetAccessPointServicesResponses,
	err error,
) {
	path := client.service + "/ReadNetAccessPointServices"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnetaccesspointservicesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNetAccessPointServicesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNetAccessPointServicesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNetAccessPoints(
	readnetaccesspointsrequest ReadNetAccessPointsRequest,
) (
	response *POST_ReadNetAccessPointsResponses,
	err error,
) {
	path := client.service + "/ReadNetAccessPoints"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnetaccesspointsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNetAccessPointsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNetAccessPointsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNetPeerings(
	readnetpeeringsrequest ReadNetPeeringsRequest,
) (
	response *POST_ReadNetPeeringsResponses,
	err error,
) {
	path := client.service + "/ReadNetPeerings"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnetpeeringsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNetPeeringsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNetPeeringsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNets(
	readnetsrequest ReadNetsRequest,
) (
	response *POST_ReadNetsResponses,
	err error,
) {
	path := client.service + "/ReadNets"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnetsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNetsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNetsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadNics(
	readnicsrequest ReadNicsRequest,
) (
	response *POST_ReadNicsResponses,
	err error,
) {
	path := client.service + "/ReadNics"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readnicsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadNicsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadNicsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadPolicies(
	readpoliciesrequest ReadPoliciesRequest,
) (
	response *POST_ReadPoliciesResponses,
	err error,
) {
	path := client.service + "/ReadPolicies"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readpoliciesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadPoliciesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadPoliciesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadPrefixLists(
	readprefixlistsrequest ReadPrefixListsRequest,
) (
	response *POST_ReadPrefixListsResponses,
	err error,
) {
	path := client.service + "/ReadPrefixLists"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readprefixlistsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadPrefixListsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadPrefixListsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadProductTypes(
	readproducttypesrequest ReadProductTypesRequest,
) (
	response *POST_ReadProductTypesResponses,
	err error,
) {
	path := client.service + "/ReadProductTypes"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readproducttypesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadProductTypesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadProductTypesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadPublicCatalog(
	readpubliccatalogrequest ReadPublicCatalogRequest,
) (
	response *POST_ReadPublicCatalogResponses,
	err error,
) {
	path := client.service + "/ReadPublicCatalog"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readpubliccatalogrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadPublicCatalogResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadPublicCatalogResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadPublicIpRanges(
	readpubliciprangesrequest ReadPublicIpRangesRequest,
) (
	response *POST_ReadPublicIpRangesResponses,
	err error,
) {
	path := client.service + "/ReadPublicIpRanges"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readpubliciprangesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadPublicIpRangesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadPublicIpRangesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadPublicIps(
	readpublicipsrequest ReadPublicIpsRequest,
) (
	response *POST_ReadPublicIpsResponses,
	err error,
) {
	path := client.service + "/ReadPublicIps"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readpublicipsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadPublicIpsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadPublicIpsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadQuotas(
	readquotasrequest ReadQuotasRequest,
) (
	response *POST_ReadQuotasResponses,
	err error,
) {
	path := client.service + "/ReadQuotas"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readquotasrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadQuotasResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadQuotasResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadRegionConfig(
	readregionconfigrequest ReadRegionConfigRequest,
) (
	response *POST_ReadRegionConfigResponses,
	err error,
) {
	path := client.service + "/ReadRegionConfig"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readregionconfigrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadRegionConfigResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadRegionConfigResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadRegions(
	readregionsrequest ReadRegionsRequest,
) (
	response *POST_ReadRegionsResponses,
	err error,
) {
	path := client.service + "/ReadRegions"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readregionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadRegionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadRegionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadReservedVmOffers(
	readreservedvmoffersrequest ReadReservedVmOffersRequest,
) (
	response *POST_ReadReservedVmOffersResponses,
	err error,
) {
	path := client.service + "/ReadReservedVmOffers"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readreservedvmoffersrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadReservedVmOffersResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadReservedVmOffersResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadReservedVms(
	readreservedvmsrequest ReadReservedVmsRequest,
) (
	response *POST_ReadReservedVmsResponses,
	err error,
) {
	path := client.service + "/ReadReservedVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readreservedvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	utils.DebugRequest(req)
	resp, err := client.Do(req)
	if resp != nil {
		utils.DebugResponse(resp)
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadReservedVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadReservedVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadRouteTables(
	readroutetablesrequest ReadRouteTablesRequest,
) (
	response *POST_ReadRouteTablesResponses,
	err error,
) {
	path := client.service + "/ReadRouteTables"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readroutetablesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadRouteTablesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadRouteTablesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadSecurityGroups(
	readsecuritygroupsrequest ReadSecurityGroupsRequest,
) (
	response *POST_ReadSecurityGroupsResponses,
	err error,
) {
	path := client.service + "/ReadSecurityGroups"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readsecuritygroupsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadSecurityGroupsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadSecurityGroupsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadServerCertificates(
	readservercertificatesrequest ReadServerCertificatesRequest,
) (
	response *POST_ReadServerCertificatesResponses,
	err error,
) {
	path := client.service + "/ReadServerCertificates"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readservercertificatesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadServerCertificatesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadServerCertificatesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadSnapshotExportTasks(
	readsnapshotexporttasksrequest ReadSnapshotExportTasksRequest,
) (
	response *POST_ReadSnapshotExportTasksResponses,
	err error,
) {
	path := client.service + "/ReadSnapshotExportTasks"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readsnapshotexporttasksrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadSnapshotExportTasksResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadSnapshotExportTasksResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadSnapshots(
	readsnapshotsrequest ReadSnapshotsRequest,
) (
	response *POST_ReadSnapshotsResponses,
	err error,
) {
	path := client.service + "/ReadSnapshots"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readsnapshotsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadSnapshotsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadSnapshotsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadSubnets(
	readsubnetsrequest ReadSubnetsRequest,
) (
	response *POST_ReadSubnetsResponses,
	err error,
) {
	path := client.service + "/ReadSubnets"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readsubnetsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadSubnetsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadSubnetsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadSubregions(
	readsubregionsrequest ReadSubregionsRequest,
) (
	response *POST_ReadSubregionsResponses,
	err error,
) {
	path := client.service + "/ReadSubregions"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readsubregionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadSubregionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadSubregionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadTags(
	readtagsrequest ReadTagsRequest,
) (
	response *POST_ReadTagsResponses,
	err error,
) {
	path := client.service + "/ReadTags"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readtagsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadTagsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadTagsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadUserGroups(
	readusergroupsrequest ReadUserGroupsRequest,
) (
	response *POST_ReadUserGroupsResponses,
	err error,
) {
	path := client.service + "/ReadUserGroups"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readusergroupsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadUserGroupsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadUserGroupsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadUsers(
	readusersrequest ReadUsersRequest,
) (
	response *POST_ReadUsersResponses,
	err error,
) {
	path := client.service + "/ReadUsers"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readusersrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadUsersResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadUsersResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVirtualGateways(
	readvirtualgatewaysrequest ReadVirtualGatewaysRequest,
) (
	response *POST_ReadVirtualGatewaysResponses,
	err error,
) {
	path := client.service + "/ReadVirtualGateways"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvirtualgatewaysrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVirtualGatewaysResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVirtualGatewaysResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVmTypes(
	readvmtypesrequest ReadVmTypesRequest,
) (
	response *POST_ReadVmTypesResponses,
	err error,
) {
	path := client.service + "/ReadVmTypes"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvmtypesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVmTypesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVmTypesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVms(
	readvmsrequest ReadVmsRequest,
) (
	response *POST_ReadVmsResponses,
	err error,
) {
	path := client.service + "/ReadVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVmsHealth(
	readvmshealthrequest ReadVmsHealthRequest,
) (
	response *POST_ReadVmsHealthResponses,
	err error,
) {
	path := client.service + "/ReadVmsHealth"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvmshealthrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVmsHealthResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVmsHealthResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVmsState(
	readvmsstaterequest ReadVmsStateRequest,
) (
	response *POST_ReadVmsStateResponses,
	err error,
) {
	path := client.service + "/ReadVmsState"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvmsstaterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVmsStateResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVmsStateResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVolumes(
	readvolumesrequest ReadVolumesRequest,
) (
	response *POST_ReadVolumesResponses,
	err error,
) {
	path := client.service + "/ReadVolumes"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvolumesrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVolumesResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVolumesResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ReadVpnConnections(
	readvpnconnectionsrequest ReadVpnConnectionsRequest,
) (
	response *POST_ReadVpnConnectionsResponses,
	err error,
) {
	path := client.service + "/ReadVpnConnections"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(readvpnconnectionsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ReadVpnConnectionsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ReadVpnConnectionsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_RebootVms(
	rebootvmsrequest RebootVmsRequest,
) (
	response *POST_RebootVmsResponses,
	err error,
) {
	path := client.service + "/RebootVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(rebootvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_RebootVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &RebootVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_RegisterUserInUserGroup(
	registeruserinusergrouprequest RegisterUserInUserGroupRequest,
) (
	response *POST_RegisterUserInUserGroupResponses,
	err error,
) {
	path := client.service + "/RegisterUserInUserGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(registeruserinusergrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_RegisterUserInUserGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &RegisterUserInUserGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_RegisterVmsInLoadBalancer(
	registervmsinloadbalancerrequest RegisterVmsInLoadBalancerRequest,
) (
	response *POST_RegisterVmsInLoadBalancerResponses,
	err error,
) {
	path := client.service + "/RegisterVmsInLoadBalancer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(registervmsinloadbalancerrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_RegisterVmsInLoadBalancerResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &RegisterVmsInLoadBalancerResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_RejectNetPeering(
	rejectnetpeeringrequest RejectNetPeeringRequest,
) (
	response *POST_RejectNetPeeringResponses,
	err error,
) {
	path := client.service + "/RejectNetPeering"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(rejectnetpeeringrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_RejectNetPeeringResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &RejectNetPeeringResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 409:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code409 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_ResetAccountPassword(
	resetaccountpasswordrequest ResetAccountPasswordRequest,
) (
	response *POST_ResetAccountPasswordResponses,
	err error,
) {
	path := client.service + "/ResetAccountPassword"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(resetaccountpasswordrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_ResetAccountPasswordResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ResetAccountPasswordResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_SendResetPasswordEmail(
	sendresetpasswordemailrequest SendResetPasswordEmailRequest,
) (
	response *POST_SendResetPasswordEmailResponses,
	err error,
) {
	path := client.service + "/SendResetPasswordEmail"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(sendresetpasswordemailrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_SendResetPasswordEmailResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &SendResetPasswordEmailResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_StartVms(
	startvmsrequest StartVmsRequest,
) (
	response *POST_StartVmsResponses,
	err error,
) {
	path := client.service + "/StartVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(startvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_StartVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &StartVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_StopVms(
	stopvmsrequest StopVmsRequest,
) (
	response *POST_StopVmsResponses,
	err error,
) {
	path := client.service + "/StopVms"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(stopvmsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_StopVmsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &StopVmsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkInternetService(
	unlinkinternetservicerequest UnlinkInternetServiceRequest,
) (
	response *POST_UnlinkInternetServiceResponses,
	err error,
) {
	path := client.service + "/UnlinkInternetService"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkinternetservicerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	utils.DebugRequest(req)
	resp, err := client.Do(req)
	if resp != nil {
		utils.DebugResponse(resp)
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkInternetServiceResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkInternetServiceResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkNic(
	unlinknicrequest UnlinkNicRequest,
) (
	response *POST_UnlinkNicResponses,
	err error,
) {
	path := client.service + "/UnlinkNic"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinknicrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkNicResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkNicResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkPolicy(
	unlinkpolicyrequest UnlinkPolicyRequest,
) (
	response *POST_UnlinkPolicyResponses,
	err error,
) {
	path := client.service + "/UnlinkPolicy"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkpolicyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkPolicyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkPolicyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkPrivateIps(
	unlinkprivateipsrequest UnlinkPrivateIpsRequest,
) (
	response *POST_UnlinkPrivateIpsResponses,
	err error,
) {
	path := client.service + "/UnlinkPrivateIps"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkprivateipsrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkPrivateIpsResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkPrivateIpsResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkPublicIp(
	unlinkpubliciprequest UnlinkPublicIpRequest,
) (
	response *POST_UnlinkPublicIpResponses,
	err error,
) {
	path := client.service + "/UnlinkPublicIp"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkpubliciprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkPublicIpResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkPublicIpResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkRouteTable(
	unlinkroutetablerequest UnlinkRouteTableRequest,
) (
	response *POST_UnlinkRouteTableResponses,
	err error,
) {
	path := client.service + "/UnlinkRouteTable"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkroutetablerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkRouteTableResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkRouteTableResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkVirtualGateway(
	unlinkvirtualgatewayrequest UnlinkVirtualGatewayRequest,
) (
	response *POST_UnlinkVirtualGatewayResponses,
	err error,
) {
	path := client.service + "/UnlinkVirtualGateway"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkvirtualgatewayrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkVirtualGatewayResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkVirtualGatewayResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UnlinkVolume(
	unlinkvolumerequest UnlinkVolumeRequest,
) (
	response *POST_UnlinkVolumeResponses,
	err error,
) {
	path := client.service + "/UnlinkVolume"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(unlinkvolumerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UnlinkVolumeResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UnlinkVolumeResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateAccount(
	updateaccountrequest UpdateAccountRequest,
) (
	response *POST_UpdateAccountResponses,
	err error,
) {
	path := client.service + "/UpdateAccount"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateaccountrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateAccountResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateAccountResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateApiKey(
	updateapikeyrequest UpdateApiKeyRequest,
) (
	response *POST_UpdateApiKeyResponses,
	err error,
) {
	path := client.service + "/UpdateApiKey"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateapikeyrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateApiKeyResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateApiKeyResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateHealthCheck(
	updatehealthcheckrequest UpdateHealthCheckRequest,
) (
	response *POST_UpdateHealthCheckResponses,
	err error,
) {
	path := client.service + "/UpdateHealthCheck"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatehealthcheckrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateHealthCheckResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateHealthCheckResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateImage(
	updateimagerequest UpdateImageRequest,
) (
	response *POST_UpdateImageResponses,
	err error,
) {
	path := client.service + "/UpdateImage"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateimagerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateImageResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateImageResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateKeypair(
	updatekeypairrequest UpdateKeypairRequest,
) (
	response *POST_UpdateKeypairResponses,
	err error,
) {
	path := client.service + "/UpdateKeypair"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatekeypairrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateKeypairResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateKeypairResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateListenerRule(
	updatelistenerrulerequest UpdateListenerRuleRequest,
) (
	response *POST_UpdateListenerRuleResponses,
	err error,
) {
	path := client.service + "/UpdateListenerRule"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatelistenerrulerequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateListenerRuleResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateListenerRuleResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateLoadBalancer(
	updateloadbalancerrequest UpdateLoadBalancerRequest,
) (
	response *POST_UpdateLoadBalancerResponses,
	err error,
) {
	path := client.service + "/UpdateLoadBalancer"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateloadbalancerrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateLoadBalancerResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateLoadBalancerResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateNet(
	updatenetrequest UpdateNetRequest,
) (
	response *POST_UpdateNetResponses,
	err error,
) {
	path := client.service + "/UpdateNet"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatenetrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateNetResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateNetResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateNetAccessPoint(
	updatenetaccesspointrequest UpdateNetAccessPointRequest,
) (
	response *POST_UpdateNetAccessPointResponses,
	err error,
) {
	path := client.service + "/UpdateNetAccessPoint"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatenetaccesspointrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateNetAccessPointResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateNetAccessPointResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateNic(
	updatenicrequest UpdateNicRequest,
) (
	response *POST_UpdateNicResponses,
	err error,
) {
	path := client.service + "/UpdateNic"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatenicrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateNicResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateNicResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateRoute(
	updaterouterequest UpdateRouteRequest,
) (
	response *POST_UpdateRouteResponses,
	err error,
) {
	path := client.service + "/UpdateRoute"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updaterouterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateRouteResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateRouteResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateRoutePropagation(
	updateroutepropagationrequest UpdateRoutePropagationRequest,
) (
	response *POST_UpdateRoutePropagationResponses,
	err error,
) {
	path := client.service + "/UpdateRoutePropagation"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateroutepropagationrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateRoutePropagationResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateRoutePropagationResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateServerCertificate(
	updateservercertificaterequest UpdateServerCertificateRequest,
) (
	response *POST_UpdateServerCertificateResponses,
	err error,
) {
	path := client.service + "/UpdateServerCertificate"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateservercertificaterequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateServerCertificateResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateServerCertificateResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateSnapshot(
	updatesnapshotrequest UpdateSnapshotRequest,
) (
	response *POST_UpdateSnapshotResponses,
	err error,
) {
	path := client.service + "/UpdateSnapshot"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatesnapshotrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateSnapshotResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateSnapshotResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateUser(
	updateuserrequest UpdateUserRequest,
) (
	response *POST_UpdateUserResponses,
	err error,
) {
	path := client.service + "/UpdateUser"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateuserrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateUserResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateUserResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateUserGroup(
	updateusergrouprequest UpdateUserGroupRequest,
) (
	response *POST_UpdateUserGroupResponses,
	err error,
) {
	path := client.service + "/UpdateUserGroup"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updateusergrouprequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateUserGroupResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateUserGroupResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	default:
		break
	}
	return
}

//
func (client *Client) POST_UpdateVm(
	updatevmrequest UpdateVmRequest,
) (
	response *POST_UpdateVmResponses,
	err error,
) {
	path := client.service + "/UpdateVm"
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(updatevmrequest)
	req, err := http.NewRequest("POST", path, body)
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	req.Header = reqHeaders
	client.Sign(req, body.Bytes())
	if err != nil {
		return
	}
	utils.DebugRequest(req)
	resp, err := client.Do(req)
	if resp != nil {
		utils.DebugResponse(resp)
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, checkErrorResponse(resp)
	}
	response = &POST_UpdateVmResponses{}
	switch {
	case resp.StatusCode == 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &UpdateVmResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.OK = result
	case resp.StatusCode == 400:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code400 = result
	case resp.StatusCode == 401:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code401 = result
	case resp.StatusCode == 500:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		result := &ErrorResponse{}
		err = json.Unmarshal(body, result)
		if err != nil {
			return nil, err
		}
		response.Code500 = result
	default:
		break
	}
	return
}

func checkErrorResponse(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response error body %s", err)
	}

	reason, errFmt := fmtErrorResponse(body)
	if errFmt != nil {
		return fmt.Errorf("error formating error resonse %s", err)
	}

	return fmt.Errorf("error, status code %d, reason: %s", resp.StatusCode, reason)
}

func fmtErrorResponse(errBody []byte) (string, error) {
	result := &ErrorResponse{}
	err := json.Unmarshal(errBody, result)
	if err != nil {
		return "", err
	}

	errors, errPretty := json.MarshalIndent(result, "", "  ")
	if errPretty != nil {
		return "", err
	}

	return string(errors), nil
}

var _ OAPIClient = (*Client)(nil)
