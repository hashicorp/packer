// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing Service API
//
// API for the Load Balancing Service
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateCertificateDetails The configuration details for adding a certificate bundle to a listener.
// For more information on SSL certficate configuration, see
// Managing SSL Certificates (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingcertificates.htm).
type CreateCertificateDetails struct {

	// A friendly name for the certificate bundle. It must be unique and it cannot be changed.
	// Valid certificate bundle names include only alphanumeric characters, dashes, and underscores.
	// Certificate bundle names cannot contain spaces. Avoid entering confidential information.
	// Example: `example_certificate_bundle`
	CertificateName *string `mandatory:"true" json:"certificateName"`

	// The Certificate Authority certificate, or any interim certificate, that you received from your SSL certificate provider.
	// Example:
	//     -----BEGIN CERTIFICATE-----
	//     MIIEczCCA1ugAwIBAgIBADANBgkqhkiG9w0BAQQFAD..AkGA1UEBhMCR0Ix
	//     EzARBgNVBAgTClNvbWUtU3RhdGUxFDASBgNVBAoTC0..0EgTHRkMTcwNQYD
	//     VQQLEy5DbGFzcyAxIFB1YmxpYyBQcmltYXJ5IENlcn..XRpb24gQXV0aG9y
	//     aXR5MRQwEgYDVQQDEwtCZXN0IENBIEx0ZDAeFw0wMD..TUwMTZaFw0wMTAy
	//     ...
	//     -----END CERTIFICATE-----
	CaCertificate *string `mandatory:"false" json:"caCertificate"`

	// A passphrase for encrypted private keys. This is needed only if you created your certificate with a passphrase.
	Passphrase *string `mandatory:"false" json:"passphrase"`

	// The SSL private key for your certificate, in PEM format.
	// Example:
	//     -----BEGIN RSA PRIVATE KEY-----
	//     jO1O1v2ftXMsawM90tnXwc6xhOAT1gDBC9S8DKeca..JZNUgYYwNS0dP2UK
	//     tmyN+XqVcAKw4HqVmChXy5b5msu8eIq3uc2NqNVtR..2ksSLukP8pxXcHyb
	//     +sEwvM4uf8qbnHAqwnOnP9+KV9vds6BaH1eRA4CHz..n+NVZlzBsTxTlS16
	//     /Umr7wJzVrMqK5sDiSu4WuaaBdqMGfL5hLsTjcBFD..Da2iyQmSKuVD4lIZ
	//     ...
	//     -----END RSA PRIVATE KEY-----
	PrivateKey *string `mandatory:"false" json:"privateKey"`

	// The public certificate, in PEM format, that you received from your SSL certificate provider.
	// Example:
	//     -----BEGIN CERTIFICATE-----
	//     MIIC2jCCAkMCAg38MA0GCSqGSIb3DQEBBQUAMIGbM..QswCQYDVQQGEwJKU
	//     A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxE..TAPBgNVBAoTCEZyY
	//     MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWB..gNVBAMTD0ZyYW5rN
	//     YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmc..mFuazRkZC5jb20wH
	//     ...
	//     -----END CERTIFICATE-----
	PublicCertificate *string `mandatory:"false" json:"publicCertificate"`
}

func (m CreateCertificateDetails) String() string {
	return common.PointerString(m)
}
