package pkcs12

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
)

// pkcs8 reflects an ASN.1, PKCS#8 PrivateKey. See
// ftp://ftp.rsasecurity.com/pub/pkcs/pkcs-8/pkcs-8v1_2.asn
// and RFC5208.
type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
	// optional attributes omitted.
}

var (
	oidPublicKeyRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

	nullAsn = asn1.RawValue{Tag: 5}
)

// marshalPKCS8PrivateKey converts a private key to PKCS#8 encoded form.
// See http://www.rsa.com/rsalabs/node.asp?id=2130 and RFC5208.
func marshalPKCS8PrivateKey(key interface{}) ([]byte, error) {
	pkcs := pkcs8{
		Version: 0,
	}

	switch key := key.(type) {
	case *rsa.PrivateKey:
		pkcs.Algo = pkix.AlgorithmIdentifier{
			Algorithm:  oidPublicKeyRSA,
			Parameters: nullAsn,
		}
		pkcs.PrivateKey = x509.MarshalPKCS1PrivateKey(key)
	case *ecdsa.PrivateKey:
		bytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, errors.New("x509: failed to marshal to PKCS#8: " + err.Error())
		}

		pkcs.Algo = pkix.AlgorithmIdentifier{
			Algorithm:  oidPublicKeyECDSA,
			Parameters: nullAsn,
		}
		pkcs.PrivateKey = bytes
	default:
		return nil, errors.New("x509: PKCS#8 only RSA and ECDSA private keys supported")
	}

	bytes, err := asn1.Marshal(pkcs)
	if err != nil {
		return nil, errors.New("x509: failed to marshal to PKCS#8: " + err.Error())
	}

	return bytes, nil
}
