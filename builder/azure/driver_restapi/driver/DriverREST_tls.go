// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver

import (
	"github.com/mitchellh/packer/builder/azure/driver_restapi/mod/pkg/net/http"
	"fmt"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/mod/pkg/crypto/tls"
	"log"
	"io"
	"io/ioutil"
	"encoding/xml"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/settings"
	"regexp"
)

type DriverRest_tls struct {
	httpClient *http.Client
}

func NewTlsDriver(pem []byte) (IDriverRest, error) {
	var cert tls.Certificate
	var err error

	cert, err = tls.X509KeyPair(pem, pem)

	if err != nil {
		return nil, err
	}

	tr := &http.Transport {
		TLSClientConfig: &tls.Config { Certificates : []tls.Certificate { cert } },
	}

	client := &http.Client { Transport: tr }

	tlsDriver := &DriverRest_tls{ httpClient: client}

	return tlsDriver, nil
}

// Exec executes REST request
func (d *DriverRest_tls) Exec(verb string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
//	var err error
	var req *http.Request

	const errorIgnoreLimit = 10
	errorIgnoreCount1 := 0
	errorIgnoreCount2 := 0

	for {
		req, err = http.NewRequest(verb, url, body)

		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err = d.httpClient.Do(req)

		if err != nil {
			return nil, err
		}

		log.Printf("Exec response: %v\n", resp)

		statusCode := resp.StatusCode

		if 	statusCode>=400 && statusCode<= 505 {

			defer resp.Body.Close()
			errXml := new (ErrorXml)

			var respBody []byte
			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			if settings.LogRawResponseError {
				log.Printf("Response raw error:\n%s\n", string(respBody))
			}

			err = xml.Unmarshal(respBody, errXml)
			if err != nil {
				return nil, err
			}

			errString := errXml.Message

			pattern := "Request needs to have a x-ms-version header"
			// Sometimes server returns strange error - ignore it
			match, _ := regexp.MatchString(pattern, errString)
			if match {
				log.Println("Exec ignoring error: " + errString)
				errorIgnoreCount1++
				if errorIgnoreCount1 == errorIgnoreLimit {
					return nil, fmt.Errorf("Remote server returned error: '%s' (%d times).", errString, errorIgnoreCount1)
				}
				continue
			}

			pattern = "The server encountered an internal error. Please retry the request"
			match, _ = regexp.MatchString(pattern, errString)
			if match {
				log.Println("Exec ignoring error: " + errString)
				errorIgnoreCount2++
				if errorIgnoreCount2 == errorIgnoreLimit {
					return nil, fmt.Errorf("Remote server returned error: '%s' (%d times).", errString, errorIgnoreCount2)
				}
				continue
			}

			err = fmt.Errorf("Remote server returned error: %s", errXml.Message)

			return nil, err
		}

		if statusCode == 307 { // Temporary Redirect
			redirectUrl , ok := resp.Header["Location"]
			if !ok {
				return nil, fmt.Errorf("%s %s", "Failed to redirect:", "header key 'Location' wasn't found")
			}
			log.Printf("Redirecting: '%s' --> '%s'", url, redirectUrl[0])
			url = redirectUrl[0]
			continue

		}

		break
	}

	return resp, err
}

type ErrorXml struct {
	Code string		`xml:"Code"`
	Message string	`xml:"Message"`
}





