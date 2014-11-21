// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package response

import (
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseResourceExtensionList(body io.ReadCloser) (*model.ResourceExtensionList, error ) {
	data, err := toModel(body, &model.ResourceExtensionList{})

	if err != nil {
		return nil, err
	}
	m := data.(*model.ResourceExtensionList)

	return m, nil
}
