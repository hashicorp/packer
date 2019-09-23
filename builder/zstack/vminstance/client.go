package vminstance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer"
)

type zClient struct {
	httpClient     *http.Client
	baseUrl        string
	accessKey      string
	keySecret      string
	ui             packer.Ui
	state_timeout  time.Duration
	create_timeout time.Duration
	replace        []string
	packer_tag     bool
}

const (
	apipath = "/zstack"
)

func Client(accessKey, keySecret, baseUrl string) (c *zClient) {
	c = &zClient{}
	c.httpClient = &http.Client{}
	c.baseUrl = baseUrl
	c.accessKey = accessKey
	c.keySecret = keySecret
	return
}

func (c *zClient) Get(path string, params map[string]string) (result map[string]interface{}, err error) {
	request, err := c.getRequest("GET", c.baseUrl, path, "", nil)
	if err != nil {
		return nil, err
	}
	c.ui.Say(fmt.Sprintf("request to: %s%s, method: GET", request.URL.Host, request.URL.RequestURI()))
	rsp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return c.getResponse(rsp, err)
}

func (c *zClient) SendRequest(path, method string, pk string, params map[string]interface{}) (result map[string]interface{}, err error) {
	request, err := c.getRequest(method, c.baseUrl, path, pk, params)

	if err != nil {
		return nil, err
	}
	c.ui.Say(fmt.Sprintf("request to: %s%s, method: %s", request.URL.Host, request.URL.RequestURI(), method))
	rsp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return c.getResponse(rsp, err)
}

func postQueryRequest(request *http.Request, conditions map[string]string) {
	q := request.URL.Query()
	q.Add("replyWithCount", "true")
	for k, v := range conditions {
		q.Add("q", fmt.Sprintf("%s=%s", k, v))
	}
	request.URL.RawQuery = q.Encode()
}

func (c *zClient) Query(path string, conditions map[string]string) (result map[string]interface{}, err error) {
	request, err := c.getRequest("QUERY", c.baseUrl, path, "", nil)
	if err != nil {
		return nil, err
	}
	// replyWithCount=true&q=type!=VCenter&q=__systemTag__!?=remote,remotebackup,aliyun,onlybackup&q=status=Connected
	postQueryRequest(request, conditions)
	rsp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return c.getResponse(rsp, err)
}

func (c *zClient) Send(path, method string, params map[string]interface{}) (result map[string]interface{}, err error) {
	return c.SendRequest(path, method, "params", params)
}

func (c *zClient) pollResult(location string) (map[string]interface{}, error) {
	httpCode := map[int]string{
		202: "wait",
		200: "succeed",
		503: "failed",
	}

	interval, _ := time.ParseDuration("100ms")
	pollerr := make(chan error)
	result := make(chan map[string]interface{})
	go func() {
		for {
			request, err := http.NewRequest("GET", location, nil)
			if err != nil {
				pollerr <- err
				return
			}
			// c.ui.Say(fmt.Sprintf("polling: %s%s", request.URL.Host, request.URL.RequestURI()))
			fmt.Printf("%v\n", request)
			rsp, err := c.httpClient.Do(request)
			if err != nil {
				pollerr <- err
				return
			}
			defer rsp.Body.Close()
			switch {
			case httpCode[rsp.StatusCode] == "wait":
				time.Sleep(interval)
			case httpCode[rsp.StatusCode] == "succeed":
				var r map[string]interface{}
				resp_body, err := ioutil.ReadAll(rsp.Body)
				if err == nil {
					err = json.Unmarshal(resp_body, &r)
				}
				if err != nil {
					pollerr <- err
				} else {
					result <- r
				}
				return
			case httpCode[rsp.StatusCode] == "failed":
				resp_body, err := ioutil.ReadAll(rsp.Body)
				if err != nil {
					pollerr <- err
				} else {
					pollerr <- fmt.Errorf(string(resp_body))
				}
				return
			default:
				resp_body, _ := ioutil.ReadAll(rsp.Body)
				pollerr <- fmt.Errorf("http error: %d, %s", rsp.StatusCode, string(resp_body))
				return
			}
		}
	}()

	select {
	case err := <-pollerr:
		return nil, err
	case r := <-result:
		return formatRsp(r)
	case <-time.After(c.create_timeout):
		return nil, fmt.Errorf("polling an API result time out after %v ms", c.create_timeout.Nanoseconds()/1000/1000)
	}
}

func formatRsp(result map[string]interface{}) (map[string]interface{}, error) {
	// if result["success"] == false {
	// 	return result, fmt.Errorf("%v", result["error"])
	// }
	return result, nil
}

func (c *zClient) getResponse(rsp *http.Response, err error) (map[string]interface{}, error) {
	resp_body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return nil, fmt.Errorf("get response status code: %d, messages: %s", rsp.StatusCode, string(resp_body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp_body, &result)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode == 200 || rsp.StatusCode == 204 {
		return formatRsp(result)
	}

	if rsp.StatusCode == 202 {
		// polling result
		location := result["location"]
		if location == "" {
			return nil, fmt.Errorf("location cannot find in response.")
		}
		if len(c.replace) > 0 {
			return c.pollResult(strings.Replace(location.(string), c.replace[0], c.replace[1], -1))
		} else {
			return c.pollResult(location.(string))
		}
	}
	return nil, fmt.Errorf("[Internal Error] the server returns an unknown status code[%d], body[%s]", rsp.StatusCode, string(resp_body))
}

func (c *zClient) getRequest(method, baseUrl, path, pk string, params map[string]interface{}) (*http.Request, error) {
	var requestBody interface{}
	var api_url string
	if method == "GET" {
		api_url = fmt.Sprintf("%s%s%s", baseUrl, apipath, path)
		first := true
		for k, v := range params {
			if first {
				api_url = fmt.Sprintf("%s?%s=%v", api_url, k, v)
				first = false
			} else {
				api_url = fmt.Sprintf("%s&%s=%v", api_url, k, v)
			}

		}
	} else if method == "QUERY" {
		api_url = fmt.Sprintf("%s%s%s", baseUrl, apipath, path)
		method = "GET"
	} else {
		pBody := make(map[string]interface{})
		for k, v := range params {
			pBody[k] = v
		}
		requestBody = map[string]interface{}{
			pk: pBody,
		}
		api_url = fmt.Sprintf("%s%s%s", baseUrl, apipath, path)
	}

	d := time.Now().Format("Mon, 02 Jan 2006 15:04:05 CST")
	var bodyData io.Reader
	if requestBody != nil {
		data, _ := json.Marshal(requestBody)
		bodyData = bytes.NewBuffer(data)
	}
	request, err := http.NewRequest(method, api_url, bodyData)

	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json;charset=utf-8")
	if path != "/v1/management-nodes/actions" {
		request.Header.Add("Authorization", SignZStack(c.accessKey, c.keySecret, method, d, strings.Split(path, "?")[0]))
	}
	request.Header.Add("Date", d)
	return request, nil
}
