package arm

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type azureErrorDetails struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []azureErrorDetails `json:"details"`
}

type azureErrorResponse struct {
	ErrorDetails azureErrorDetails `json:"error"`
}

func newAzureErrorResponse(s string) *azureErrorResponse {
	var errorResponse azureErrorResponse
	err := json.Unmarshal([]byte(s), &errorResponse)
	if err == nil {
		return &errorResponse
	}

	return nil
}

func (e *azureErrorDetails) isEmpty() bool {
	return e.Code == ""
}

func (e *azureErrorResponse) isEmpty() bool {
	return e.ErrorDetails.isEmpty()
}

func (e *azureErrorResponse) Error() string {
	var buf bytes.Buffer
	//buf.WriteString("-=-=- ERROR -=-=-")
	formatAzureErrorResponse(e.ErrorDetails, &buf, "")
	//buf.WriteString("-=-=- ERROR -=-=-")
	return buf.String()
}

// format a Azure Error Response by recursing through the JSON structure.
//
// Errors may contain nested errors, which are JSON documents that have been
// serialized and escaped.  Keep following this nesting all the way down...
func formatAzureErrorResponse(error azureErrorDetails, buf *bytes.Buffer, indent string) {
	if error.isEmpty() {
		return
	}

	buf.WriteString(fmt.Sprintf("ERROR: %s-> %s : %s\n", indent, error.Code, error.Message))
	for _, x := range error.Details {
		newIndent := fmt.Sprintf("%s  ", indent)

		var aer azureErrorResponse
		err := json.Unmarshal([]byte(x.Message), &aer)
		if err == nil {
			buf.WriteString(fmt.Sprintf("ERROR: %s-> %s\n", newIndent, x.Code))
			formatAzureErrorResponse(aer.ErrorDetails, buf, newIndent)
		} else {
			buf.WriteString(fmt.Sprintf("ERROR: %s-> %s : %s\n", newIndent, x.Code, x.Message))
		}
	}
}
