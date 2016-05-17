package winrm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// KeyType for RSA/ECDSA
type KeyType uint8

// CertConfig struct for
// generating base certificates x509
// for ssl/tls auth
type CertConfig struct {
	// Subject identifies the entity associated with the public
	// key stored in the subject public key fields.
	Subject pkix.Name
	// ValidFrom creation date formated as Feb 16 10:04:20 2016
	ValidFrom time.Time
	// ValidFor the duration that certificate will be valid for
	// The validity period for a certificate is the period of time
	// form notBefore through notAfter, inclusive.
	//	Note this fileds will be parsed into the notAfter,notBefore
	//	x509 cert fields
	ValidFor time.Duration
	// SizeT can represent the size of bytes RSA cert if you specify in the
	SizeT int
	// Method to be RSA or it can be ECDSA flag if you specify the ECDSA flag
	Method KeyType
}

// definition flags for specific elliptic curves formats
const (
	// P224 curve which implements P-224 (see FIPS 186-3, section D.2.2)
	P224 = iota
	// P256 curve which implements P-256 (see FIPS 186-3, section D.2.3)
	P256
	// P384 curve which implements P-384 (see FIPS 186-3, section D.2.4)
	P384
	// P521 curve which implements P-521 (see FIPS 186-3, section D.2.5)
	P521

	// definition flags for specific methods that the cert will be generated
	// RSA method
	RSA KeyType = iota
	// ECDSA method
	ECDSA
)

// OtherName type for asn1 encoding
type OtherName struct {
	A string `asn1:"utf8"`
}

// GeneralName type for asn1 encoding
type GeneralName struct {
	OID       asn1.ObjectIdentifier
	OtherName `asn1:"tag:0"`
}

// GeneralNames type for asn1 encoding
type GeneralNames struct {
	GeneralName `asn1:"tag:0"`
}

var (
	// https://support.microsoft.com/en-us/kb/287547
	//  szOID_NT_PRINCIPAL_NAME 1.3.6.1.4.1.311.20.2.3
	szOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 20, 2, 3}
	// http://www.umich.edu/~x509/ssleay/asn1-oids.html
	// 2 5 29 17  subjectAltName
	subjAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
)

//  getUPNExtensionValue returns marsheled asn1 encoded info
func getUPNExtensionValue(subject pkix.Name) ([]byte, error) {
	// returns the ASN.1 encoding of val
	// in addition to the struct tags recognized
	// we used:
	// utf8 => causes string to be marsheled as ASN.1, UTF8 strings
	// tag:x => specifies the ASN.1 tag number; imples ASN.1 CONTEXT SPECIFIC
	return asn1.Marshal(GeneralNames{
		GeneralName: GeneralName{
			// init our ASN.1 object identifier
			OID: szOID,
			OtherName: OtherName{
				A: subject.CommonName,
			},
		},
	})
}

// NewCert generates a new x509 certificates based on the CertConfig passed
func NewCert(config CertConfig) (string, string, error) {
	var (
		err       error
		notAfter  time.Time
		notBefore time.Time
		// PrivateKey that can be:
		// type e.g *rsa.PrivateKey
		// or *ecdsa.PrivateKey
		PrivateKey interface{}
	)
	// parse the time vars from the config into notAfter and notBefore
	notBefore = config.ValidFrom
	notAfter = notBefore.Add(config.ValidFor)

	// choose the method to generate the certificate
	switch config.Method {
	case RSA:
		PrivateKey, err = rsa.GenerateKey(rand.Reader, config.SizeT)
		if err != nil {
			return "", "", fmt.Errorf("Can't generate rsa private key")
		}

	// if the ECDSA byte is set
	case ECDSA:
		PrivateKey, err = genKeyEcdsa(config.SizeT)
		if err != nil {
			return "", "", fmt.Errorf("Can't generate ecdsa private key")
		}
	default:
		return "", "", fmt.Errorf("Unknown method type to use")
	}

	// get asn1 encoded info of the subject pkix
	value, err := getUPNExtensionValue(config.Subject)
	if err != nil {
		return "", "", fmt.Errorf("Can't marshal asn1 encoded")
	}

	// A serial number can be up to 20 octets in size.
	// https://tools.ietf.org/html/rfc5280#section-4.1.2.2
	serialNum, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 8*20))
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate serial number: %s", err.Error())
	}

	// make a new x509 temaplte certificate
	// entry point for the generation process
	template := x509.Certificate{
		SerialNumber: serialNum,
		Subject:      config.Subject,
		// when extenstions are used, as expected in this profile,
		// versions MUST be 3(value is 2).
		Version:     2,
		NotBefore:   config.ValidFrom,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		ExtraExtensions: []pkix.Extension{
			{
				Id:       subjAltName,
				Critical: false,
				Value:    value,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, getPublicKey(PrivateKey), PrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate cert")
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	privPem, err := exportPrivKeyToPem(PrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("Failed to export priv key")
	}

	return string(certPem), string(privPem), nil
}

// getPublicKey fetch public key with assertion
func getPublicKey(p interface{}) interface{} {
	switch t := p.(type) {
	case *rsa.PrivateKey:
		return t.Public()
	case *ecdsa.PrivateKey:
		return t.Public()
	default:
		return nil
	}
}

// choose the format for the key and generate it
func genKeyEcdsa(szT int) (*ecdsa.PrivateKey, error) {
	switch szT {
	case P224:
		return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case P256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case P384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case P521:
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("Can't generate private key, please specify a format")
	}

}

// ExportPrivKeyToPem exports private key from the previous generation
func exportPrivKeyToPem(key interface{}) ([]byte, error) {

	var pemBlock *pem.Block

	switch k := key.(type) {
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("Unable to export ecdsa private key")
		}
		pemBlock = &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: b,
		}
	default:
		return nil, fmt.Errorf("The private key is not RSA or ECDSA")
	}

	return pem.EncodeToMemory(pemBlock), nil
}
