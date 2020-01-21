package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// allow override for unit tests
var getSubscriptionFromIMDS = _getSubscriptionFromIMDS

func _getSubscriptionFromIMDS() (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute", nil)
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", "2017-08-01")

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	result := map[string]string{}
	err = json.Unmarshal(resp_body, &result)
	if err != nil {
		return "", err
	}

	return result["subscriptionId"], nil
}
