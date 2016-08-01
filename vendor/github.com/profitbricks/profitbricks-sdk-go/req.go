package profitbricks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/profitbricks/profitbricks-sdk-go/model"
)

//FullHeader is the standard header to include with all http requests except is_patch and is_command
const FullHeader = "application/vnd.profitbricks.resource+json"

//PatchHeader is used with is_patch .
const PatchHeader = "application/vnd.profitbricks.partial-properties+json"

//CommandHeader is used with is_command
const CommandHeader = "application/x-www-form-urlencoded"

var Depth = "5"

// SetDepth is used to set Depth
func SetDepth(newdepth string) string {
	Depth = newdepth
	return Depth
}

// mk_url  either:
// returns the path (if it`s a full url)
//			 or
//	returns	Endpoint+ path .
func mk_url(path string) string {
	if strings.HasPrefix(path, "http") {
		path := strings.Replace(path, "https://api.profitbricks.com/rest/v2", Endpoint, 1)
		return path
	}
	if strings.HasPrefix(path, "<base>") {
		path := strings.Replace(path, "<base>", Endpoint, 1)
		return path
	}
	url := Endpoint + path
	return url
}

func do(req *http.Request) Resp {
	client := &http.Client{}
	req.SetBasicAuth(Username, Passwd)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	var R Resp
	R.Req = resp.Request
	R.Body = resp_body
	R.Headers = resp.Header
	R.StatusCode = resp.StatusCode
	return R
}

// is_delete performs an http.NewRequest DELETE  and returns a Resp struct
func is_delete(path string) Resp {
	url := mk_url(path)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return do(req)
}

// is_list performs an http.NewRequest GET and returns a Collection struct
func is_list(path string) Collection {
	url := mk_url(path) + `?depth=` + Depth
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return toCollection(do(req))
}

// is_get performs an http.NewRequest GET and returns an Instance struct
func is_get(path string) Instance {
	url := mk_url(path) + `?depth=` + Depth
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return toInstance(do(req))
}

// is_patch performs an http.NewRequest PATCH and returns an Instance struct
func is_patch(path string, jason []byte) Instance {
	url := mk_url(path)
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jason))
	req.Header.Add("Content-Type", PatchHeader)
	return toInstance(do(req))
}

// is_put performs an http.NewRequest PUT and returns an Instance struct
func is_put(path string, jason []byte) Instance {
	url := mk_url(path)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jason))
	req.Header.Add("Content-Type", FullHeader)
	return toInstance(do(req))
}

// is_post performs an http.NewRequest POST and returns an Instance struct
func is_post(path string, jason []byte) Instance {
	url := mk_url(path)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jason))
	req.Header.Add("Content-Type", FullHeader)
	return toInstance(do(req))
}

func is_composite_post(path string, jason []byte) model.Datacenter {
	url := mk_url(path)+ `?depth=` + Depth
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jason))
	req.Header.Add("Content-Type", FullHeader)
	return toDataCenter(do(req))
}

// is_command performs an http.NewRequest POST and returns a Resp struct
func is_command(path string, jason string) Resp {
	url := mk_url(path)
	body := json.RawMessage(jason)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", CommandHeader)
	return do(req)
}
