// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package response

import (
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseServiceCertificateList(body io.ReadCloser) (*model.ServiceCertificateList, error ) {
	data, err := toModel(body, &model.ServiceCertificateList{})

	if err != nil {
		return nil, err
	}
	m := data.(*model.ServiceCertificateList)

	return m, nil
}
