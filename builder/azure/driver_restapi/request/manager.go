// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"io"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/mod/pkg/net/http"
	restapi "github.com/mitchellh/packer/builder/azure/driver_restapi/driver"
	"fmt"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"time"
	"log"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/settings"
)

type Data struct {
	Verb string
	Uri string
	Headers map[string]string
	Body io.Reader
}

type Manager struct {
	SubscrId string
	Driver restapi.IDriverRest
}

func (m *Manager) Execute(req *Data) (resp *http.Response, err error) {
	if settings.LogRequestData {
		log.Printf("Manager.Execute Request Data:\n %v", req)
	}

	resp, err = m.Driver.Exec(req.Verb, req.Uri, req.Headers, req.Body)
	return
}

func (m *Manager) ExecuteSync(req *Data) (error) {
	if settings.LogRequestData {
		log.Printf("Manager.ExecuteSync Request Data:\n %v", req)
	}

	resp, err := m.Driver.Exec(req.Verb, req.Uri, req.Headers, req.Body)

	if err != nil {
		return err
	}

	errorMsg := "Manager.ExecuteSync: %s"

	reqId, ok := resp.Header["X-Ms-Request-Id"]
	if !ok {
		return fmt.Errorf(errorMsg, "header key 'X-Ms-Request-Id' wasn't found")
	}

	count := 60
	var duration time.Duration = 15
	sleepTime := time.Second * duration

	for count != 0 {
		requestData := m.GetOperationStatus(reqId[0])
		resp, err := m.Execute(requestData)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			return err
		}

		operation, err := response.ParseOperation(resp.Body)
		log.Println(fmt.Sprintf("operation: %v", operation))

		if operation.Status == "Succeeded" {
			break;
		}

		if operation.Status == "Failed" {
			return fmt.Errorf(errorMsg, operation.Error.Message)
		}

		// InProgress
		log.Println(fmt.Sprintf("Waiting for another %v seconds...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if(count == 0){
		err := fmt.Errorf(errorMsg, "timeout")
		return err
	}

	return nil
}



