// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package cert

import (
	"fmt"
	"crypto/x509"
	"crypto/sha1"
	"io/ioutil"
	"encoding/pem"
)

func GetThumbprint(pemPath string)  (string, error){
	certs, err := loadX509Pem(pemPath)
	if err != nil {
		return "", err
	}

	for _, cert := range(certs) {
		tp := calcThumbprint(cert)
		return tp.toString(), nil
	}
	
	return "", fmt.Errorf("no cert found")
}

type Thumbprint []byte

const CERTIFICATE = "CERTIFICATE"

func calcThumbprint(cert *x509.Certificate) Thumbprint {
	h := sha1.New()
	h.Write(cert.Raw)
	return Thumbprint(h.Sum(nil))
}

func loadX509Pem(pemPath string) ([]*x509.Certificate, error) {
	pemData, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, err
	}
	return x509Pem(pemData)
}

func x509Pem(pemData []byte) (certs []*x509.Certificate, err error) {
	var block *pem.Block
//	var pemBytes []byte
	for {
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}
		if block.Type != CERTIFICATE {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func (t Thumbprint) toString() string {
	return fmt.Sprintf("%X", t)
}
