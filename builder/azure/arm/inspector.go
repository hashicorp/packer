package arm

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/mitchellh/packer/builder/azure/common/logutil"
	"io"
)

func chop(data []byte, maxlen int64) string {
	s := string(data)
	if int64(len(s)) > maxlen {
		s = s[:maxlen] + "..."
	}
	return s
}

func handleBody(body io.ReadCloser, maxlen int64) (io.ReadCloser, string) {
	if body == nil {
		return nil, ""
	}

	defer body.Close()

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, ""
	}

	return ioutil.NopCloser(bytes.NewReader(b)), chop(b, maxlen)
}

func withInspection(maxlen int64) autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			body, bodyString := handleBody(r.Body, maxlen)
			r.Body = body

			log.Print("Azure request", logutil.Fields{
				"method":  r.Method,
				"request": r.URL.String(),
				"body":    bodyString,
			})
			return p.Prepare(r)
		})
	}
}

func byInspecting(maxlen int64) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			body, bodyString := handleBody(resp.Body, maxlen)
			resp.Body = body

			log.Print("Azure response", logutil.Fields{
				"status":          resp.Status,
				"method":          resp.Request.Method,
				"request":         resp.Request.URL.String(),
				"x-ms-request-id": azure.ExtractRequestID(resp),
				"body":            bodyString,
			})
			return r.Respond(resp)
		})
	}
}
