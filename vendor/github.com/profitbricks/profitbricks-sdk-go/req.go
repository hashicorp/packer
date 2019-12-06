package profitbricks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

//FullHeader is the standard header to include with all http requests except is_patch and is_command
const FullHeader = "application/json"

var AgentHeader = "profitbricks-sdk-go/3.0.1"

//PatchHeader is used with is_patch .
const PatchHeader = "application/json"

//CommandHeader is used with is_command
const CommandHeader = "application/x-www-form-urlencoded"

var Depth = "5"
var Pretty = true

// SetDepth is used to set Depth
func SetDepth(newdepth string) string {
	Depth = newdepth
	return Depth
}

// SetDepth is used to set Depth
func SetPretty(pretty bool) bool {
	Pretty = pretty
	return Pretty
}

// mk_url  either:
// returns the path (if it`s a full url)
//			 or
//	returns	Endpoint+ path .
func mk_url(path string) string {
	if strings.HasPrefix(path, "http") {
		//REMOVE AFTER TESTING
		//FIXME @jasmin Is this still relevant?
		path := strings.Replace(path, "https://api.profitbricks.com/cloudapi/v3", Endpoint, 1)
		// END REMOVE
		return path
	}
	if strings.HasPrefix(path, "<base>") {
		//REMOVE AFTER TESTING
		path := strings.Replace(path, "<base>", Endpoint, 1)
		return path
	}
	url := Endpoint + path
	return url
}

func do(req *http.Request) Resp {
	client := &http.Client{}
	req.SetBasicAuth(Username, Passwd)
	req.Header.Add("User-Agent", AgentHeader)
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
	req.Header.Add("User-Agent", AgentHeader)
	return do(req)
}

// is_command performs an http.NewRequest POST and returns a Resp struct
func is_command(path string, jason string) Resp {
	url := mk_url(path)
	body := json.RawMessage(jason)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", CommandHeader)
	req.Header.Add("User-Agent", AgentHeader)
	return do(req)
}
