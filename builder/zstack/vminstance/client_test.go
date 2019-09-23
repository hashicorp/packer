package vminstance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func createClient() (client *zClient) {
	config := BuilderConfigTest()
	client = &zClient{
		httpClient: &http.Client{},
		baseUrl:    config["base_url"].(string),
		accessKey:  config["access_key"].(string),
		keySecret:  config["key_secret"].(string),
	}
	return
}
func TestClient_ImplementsClient(t *testing.T) {
	client := createClient()
	if client.accessKey == "" || client.keySecret == "" || client.baseUrl == "" {
		t.Fatalf("initial zClient failed")
	}
}

func validateGetRequest(req *http.Request, expected string, params []string) error {
	url := fmt.Sprintf("%s%s", req.URL.Host, req.URL.RequestURI())
	if !strings.HasPrefix(url, expected) {
		return fmt.Errorf("url[%s] not start with [%s]", url, expected)
	}
	if len(params) == 0 {
		return nil
	}
	suffix := url[len(expected):]
	if !strings.HasPrefix(suffix, "?") {
		return fmt.Errorf("url[%s] need ?", url)
	}
	l := len(params)
	if len(strings.Split(suffix, "&")) != l {
		return fmt.Errorf("url[%s], %d params need %d &, but got %d", url, l, l-1, len(strings.Split(suffix, "&")))
	}
	for _, p := range params {
		if !strings.Contains(suffix, p) {
			return fmt.Errorf("url[%s] not include [%s]", url, p)
		}
	}
	return nil
}

func validateQueryRequest(req *http.Request, expected string, params map[string]string) error {
	url := fmt.Sprintf("%s%s", req.URL.Host, req.URL.RequestURI())
	if url != expected {
		return fmt.Errorf("expected: [%s], but got: [%s]", expected, url)
	}
	postQueryRequest(req, params)
	rawQuery := fmt.Sprintf("%v", req.URL.RawQuery)
	if !strings.Contains(rawQuery, "replyWithCount=true") {
		return fmt.Errorf("rawQuery[%s] need 'replyWithCount=true'", rawQuery)
	}
	l := len(params)
	if l == 0 {
		return nil
	}
	q := strings.Split(rawQuery, "&")
	if len(q) != len(params)+1 {
		return fmt.Errorf("rawQuery[%s], %d params need %d &, but got %d", rawQuery, l, l, len(q))
	}
	for k, v := range params {
		p := fmt.Sprintf("%s3D%s", k, v)
		if strings.Contains(rawQuery, p) {
			return fmt.Errorf("rawQuery[%s] not include [%s]", rawQuery, p)
		}
	}

	return nil
}

func validateSendRequest(req *http.Request, expectUrl string, params map[string]interface{}) error {
	url := fmt.Sprintf("%s%s", req.URL.Host, req.URL.RequestURI())
	if url != expectUrl {
		return fmt.Errorf("expected: [%s], but got: [%s]", expectUrl, url)
	}
	s, _ := ioutil.ReadAll(req.Body)
	var r map[string]map[string]interface{}
	err := json.Unmarshal(s, &r)
	if err != nil {
		return err
	}
	if r["params"] == nil {
		return fmt.Errorf("body should contains params")
	}
	for k, v := range params {
		// convert float64 to int
		if reflect.Float64 == reflect.TypeOf(r["params"][k]).Kind() && reflect.Int == reflect.TypeOf(v).Kind() {
			r["params"][k], _ = strconv.Atoi(fmt.Sprintf("%v", r["params"][k]))
		}

		if r["params"][k] != v {
			return fmt.Errorf("%s, %s, %v", r["params"][k], k, v)
			// return fmt.Errorf("need: %v, but got body: %v", params, r["params"])
		}
	}
	return nil
}

func testRequestUrl(method, path string) (err error) {
	client := createClient()
	expected := fmt.Sprintf("%s%s%s", strings.Split(client.baseUrl, "http://")[1], apipath, path)
	params := map[string]interface{}{
		"k1": "v1",
		"k2": 2,
		"k3": true,
	}
	request, err := client.getRequest(method, client.baseUrl, path, "params", params)
	if err != nil {
		return
	}
	if method == "GET" {
		p := []string{
			"k1=v1",
			"k2=2",
			"k3=true",
		}
		err = validateGetRequest(request, expected, p)
	} else if method == "QUERY" {
		p := map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "true",
		}
		err = validateQueryRequest(request, expected, p)
	} else {
		err = validateSendRequest(request, expected, params)
	}

	if err != nil {
		return err
	}
	// Header include: Authorization, Content-Type, Date
	if len(request.Header) != 3 {
		err = fmt.Errorf("%v", request.Header)
	}
	return
}
func TestClient_GetRequest(t *testing.T) {
	err := testRequestUrl("GET", "/getvm")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestClient_QueryRequest(t *testing.T) {
	err := testRequestUrl("QUERY", "/queryvm")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestClient_SendRequest(t *testing.T) {
	err := testRequestUrl("POST", "/createvm")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func generateBodys() [][]byte {
	bodies := [][]byte{}
	rs := []map[string]interface{}{
		{
			"error": map[string]interface{}{
				"code":       "LICENSE.1000",
				"descripton": "Expired license",
				"details":    "the license has been expired, please renew it",
			},
			"success": false,
		},
		{
			"inventory": map[string]interface{}{
				"uuid":  "f42665b500064582865525b43383f87c",
				"name":  "vm1",
				"state": "Running",
			},
			"success": true,
		},
	}

	for _, r := range rs {
		body, _ := json.Marshal(r)
		bodies = append(bodies, body)
	}

	return bodies
}
func TestClient_GetResponse(t *testing.T) {
	client := createClient()

	bodies := generateBodys()
	cases := []struct {
		Body       io.ReadCloser
		StatusCode int
		Success    bool
	}{
		{
			ioutil.NopCloser(bytes.NewReader(bodies[0])),
			100,
			false,
		},
		{
			ioutil.NopCloser(bytes.NewReader(bodies[0])),
			300,
			false,
		},
		{
			ioutil.NopCloser(bytes.NewReader(bodies[0])),
			206,
			false,
		},
		{
			ioutil.NopCloser(bytes.NewReader(bodies[0])),
			200,
			true,
		},
		{
			ioutil.NopCloser(bytes.NewReader(bodies[0])),
			204,
			true,
		},
		{
			ioutil.NopCloser(bytes.NewReader(bodies[1])),
			200,
			true,
		},
	}

	for _, c := range cases {
		rsp := &http.Response{
			Body:       c.Body,
			StatusCode: c.StatusCode,
		}
		rsp_body, err := client.getResponse(rsp, nil)
		if err == nil && !c.Success {
			t.Fatalf("expected false, but got success")
		} else {
			// if success, check bodies
			s, _ := ioutil.ReadAll(c.Body)
			var r map[string]interface{}
			json.Unmarshal(s, &r)
			for k, v := range r {
				if v != rsp_body[k] {
					t.Fatalf("%s, %v, %v", k, v, rsp_body[k])
				}
			}
		}
	}
}
