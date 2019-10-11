package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// DebugRequest ...
func DebugRequest(req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("[DEBUG] Request\n%s", string(requestDump))
}

// DebugResponse ...
func DebugResponse(res *http.Response) {
	responseDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("[DEBUG] Response\n%s", string(responseDump))
}
