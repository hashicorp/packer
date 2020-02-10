package vagrantcloud

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVagranCloudErrors(t *testing.T) {
	testCases := []struct {
		resp     string
		expected string
	}{
		{`{"Status":"422 Unprocessable Entity", "StatusCode":422, "errors":[]}`, ""},
		{`{"Status":"404 Artifact not found", "StatusCode":404, "errors":["error1", "error2"]}`, "error1. error2"},
		{`{"StatusCode":403, "errors":[{"message":"Bad credentials"}]}`, "message Bad credentials"},
		{`{"StatusCode":500, "errors":[["error in unexpected format"]]}`, "[error in unexpected format]"},
	}

	for _, tc := range testCases {
		var cloudErrors VagrantCloudErrors
		err := json.NewDecoder(strings.NewReader(tc.resp)).Decode(&cloudErrors)
		if err != nil {
			t.Errorf("failed to decode error response: %s", err)
		}
		if got := cloudErrors.FormatErrors(); got != tc.expected {
			t.Errorf("failed to get expected response; expected %q, got %q", tc.expected, got)
		}
	}
}
