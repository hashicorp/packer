// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package cert

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"bytes"
//	"os"
)

func ToX509(base64X509 string) (*x509.Certificate, error) {
	return nil, fmt.Errorf("not implemented")
}

func ToX509File(base64X509, certPath string) (error) {

	var err error
	var errMsg = "ToX509File error %s"

//	if _, err = os.Stat(certPath); err != nil {
//		return fmt.Errorf(errMsg, "check the certPath is correct.")
//	}

	var lineLen int = 64
	var output bytes.Buffer

	dataLen := len(base64X509)

	output.WriteString("-----BEGIN CERTIFICATE-----\n")
	for offset := 0; offset < dataLen; offset += lineLen {
		remaining := dataLen - offset
		if remaining < lineLen {
			output.WriteString(base64X509[offset:offset+remaining]+"\n")
		} else {
			output.WriteString(base64X509[offset:offset+lineLen]+"\n")
		}
	}
	output.WriteString("-----END CERTIFICATE-----\n")

	err = ioutil.WriteFile(certPath, output.Bytes(), 0700)

	if err != nil {
		return fmt.Errorf(errMsg, err.Error())
	}

	return nil
}

