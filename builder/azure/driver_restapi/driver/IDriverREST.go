// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver

import (
	"github.com/mitchellh/packer/builder/azure/driver_restapi/mod/pkg/net/http"
	"io"
)

// A driver is able to talk to Azure via REST API and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the HyperV builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.

type IDriverRest interface {

	// Exec executes
	Exec(verb string, url string, headers  map[string]string, body io.Reader) (*http.Response, error)
}


