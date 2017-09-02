// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pkcs12 implements some of PKCS#12.
//
// This implementation is distilled from https://tools.ietf.org/html/rfc7292
// and referenced documents. It is intended for decoding P12/PFX-stored
// certificates and keys for use with the crypto/tls package.
package pkcs12

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
)

var (
	oidDataContentType          = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 7, 1})
	oidEncryptedDataContentType = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 7, 6})

	oidFriendlyName     = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 9, 20})
	oidLocalKeyID       = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 9, 21})
	oidMicrosoftCSPName = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 4, 1, 311, 17, 1})

	localKeyId = []byte{0x01, 0x00, 0x00, 0x00}
)

type pfxPdu struct {
	Version  int
	AuthSafe contentInfo
	MacData  macData `asn1:"optional"`
}

type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"tag:0,explicit,optional"`
}

type encryptedData struct {
	Version              int
	EncryptedContentInfo encryptedContentInfo
}

type encryptedContentInfo struct {
	ContentType                asn1.ObjectIdentifier
	ContentEncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedContent           []byte `asn1:"tag:0,optional"`
}

func (i encryptedContentInfo) Algorithm() pkix.AlgorithmIdentifier {
	return i.ContentEncryptionAlgorithm
}

func (i encryptedContentInfo) Data() []byte { return i.EncryptedContent }

type safeBag struct {
	Id         asn1.ObjectIdentifier
	Value      asn1.RawValue     `asn1:"tag:0,explicit"`
	Attributes []pkcs12Attribute `asn1:"set,optional"`
}

type pkcs12Attribute struct {
	Id    asn1.ObjectIdentifier
	Value asn1.RawValue `asn1:"set"`
}

type encryptedPrivateKeyInfo struct {
	AlgorithmIdentifier pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

func (i encryptedPrivateKeyInfo) Algorithm() pkix.AlgorithmIdentifier {
	return i.AlgorithmIdentifier
}

func (i encryptedPrivateKeyInfo) Data() []byte {
	return i.EncryptedData
}

// PEM block types
const (
	certificateType = "CERTIFICATE"
	privateKeyType  = "PRIVATE KEY"
)

// unmarshal calls asn1.Unmarshal, but also returns an error if there is any
// trailing data after unmarshaling.
func unmarshal(in []byte, out interface{}) error {
	trailing, err := asn1.Unmarshal(in, out)
	if err != nil {
		return err
	}
	if len(trailing) != 0 {
		return errors.New("pkcs12: trailing data found")
	}
	return nil
}

// ConvertToPEM converts all "safe bags" contained in pfxData to PEM blocks.
func ToPEM(pfxData []byte, password string) ([]*pem.Block, error) {
	encodedPassword, err := bmpString(password)
	if err != nil {
		return nil, ErrIncorrectPassword
	}

	bags, encodedPassword, err := getSafeContents(pfxData, encodedPassword)

	blocks := make([]*pem.Block, 0, len(bags))
	for _, bag := range bags {
		block, err := convertBag(&bag, encodedPassword)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func convertBag(bag *safeBag, password []byte) (*pem.Block, error) {
	block := &pem.Block{
		Headers: make(map[string]string),
	}

	for _, attribute := range bag.Attributes {
		k, v, err := convertAttribute(&attribute)
		if err != nil {
			return nil, err
		}
		block.Headers[k] = v
	}

	switch {
	case bag.Id.Equal(oidCertBag):
		block.Type = certificateType
		certsData, err := decodeCertBag(bag.Value.Bytes)
		if err != nil {
			return nil, err
		}
		block.Bytes = certsData
	case bag.Id.Equal(oidPKCS8ShroudedKeyBag):
		block.Type = privateKeyType

		key, err := decodePkcs8ShroudedKeyBag(bag.Value.Bytes, password)
		if err != nil {
			return nil, err
		}

		switch key := key.(type) {
		case *rsa.PrivateKey:
			block.Bytes = x509.MarshalPKCS1PrivateKey(key)
		case *ecdsa.PrivateKey:
			block.Bytes, err = x509.MarshalECPrivateKey(key)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	default:
		return nil, errors.New("don't know how to convert a safe bag of type " + bag.Id.String())
	}
	return block, nil
}

func convertAttribute(attribute *pkcs12Attribute) (key, value string, err error) {
	isString := false

	switch {
	case attribute.Id.Equal(oidFriendlyName):
		key = "friendlyName"
		isString = true
	case attribute.Id.Equal(oidLocalKeyID):
		key = "localKeyId"
	case attribute.Id.Equal(oidMicrosoftCSPName):
		// This key is chosen to match OpenSSL.
		key = "Microsoft CSP Name"
		isString = true
	default:
		return "", "", errors.New("pkcs12: unknown attribute with OID " + attribute.Id.String())
	}

	if isString {
		if err := unmarshal(attribute.Value.Bytes, &attribute.Value); err != nil {
			return "", "", err
		}
		if value, err = decodeBMPString(attribute.Value.Bytes); err != nil {
			return "", "", err
		}
	} else {
		var id []byte
		if err := unmarshal(attribute.Value.Bytes, &id); err != nil {
			return "", "", err
		}
		value = hex.EncodeToString(id)
	}

	return key, value, nil
}

// Decode extracts a certificate and private key from pfxData. This function
// assumes that there is only one certificate and only one private key in the
// pfxData.
func Decode(pfxData []byte, password string) (privateKey interface{}, certificate *x509.Certificate, err error) {
	encodedPassword, err := bmpString(password)
	if err != nil {
		return nil, nil, err
	}

	bags, encodedPassword, err := getSafeContents(pfxData, encodedPassword)
	if err != nil {
		return nil, nil, err
	}

	if len(bags) != 2 {
		err = errors.New("pkcs12: expected exactly two safe bags in the PFX PDU")
		return
	}

	for _, bag := range bags {
		switch {
		case bag.Id.Equal(oidCertBag):
			if certificate != nil {
				err = errors.New("pkcs12: expected exactly one certificate bag")
			}

			certsData, err := decodeCertBag(bag.Value.Bytes)
			if err != nil {
				return nil, nil, err
			}
			certs, err := x509.ParseCertificates(certsData)
			if err != nil {
				return nil, nil, err
			}
			if len(certs) != 1 {
				err = errors.New("pkcs12: expected exactly one certificate in the certBag")
				return nil, nil, err
			}
			certificate = certs[0]

		case bag.Id.Equal(oidPKCS8ShroudedKeyBag):
			if privateKey != nil {
				err = errors.New("pkcs12: expected exactly one key bag")
			}
			if privateKey, err = decodePkcs8ShroudedKeyBag(bag.Value.Bytes, encodedPassword); err != nil {
				return nil, nil, err
			}
		}
	}

	if certificate == nil {
		return nil, nil, errors.New("pkcs12: certificate missing")
	}
	if privateKey == nil {
		return nil, nil, errors.New("pkcs12: private key missing")
	}

	return
}

func getLocalKeyId(id []byte) (attribute pkcs12Attribute, err error) {
	octetString := asn1.RawValue{Tag: 4, Class: 0, IsCompound: false, Bytes: id}
	bytes, err := asn1.Marshal(octetString)
	if err != nil {
		return
	}

	attribute = pkcs12Attribute{
		Id:    oidLocalKeyID,
		Value: asn1.RawValue{Tag: 17, Class: 0, IsCompound: true, Bytes: bytes},
	}

	return attribute, nil
}

func convertToRawVal(val interface{}) (raw asn1.RawValue, err error) {
	bytes, err := asn1.Marshal(val)
	if err != nil {
		return
	}

	_, err = asn1.Unmarshal(bytes, &raw)
	return raw, nil
}

func makeSafeBags(oid asn1.ObjectIdentifier, value []byte) ([]safeBag, error) {
	attribute, err := getLocalKeyId(localKeyId)

	if err != nil {
		return nil, EncodeError("local key id: " + err.Error())
	}

	bag := make([]safeBag, 1)
	bag[0] = safeBag{
		Id:         oid,
		Value:      asn1.RawValue{Tag: 0, Class: 2, IsCompound: true, Bytes: value},
		Attributes: []pkcs12Attribute{attribute},
	}

	return bag, nil
}

func makeCertBagContentInfo(derBytes []byte) (*contentInfo, error) {
	certBag1 := certBag{
		Id:   oidCertTypeX509Certificate,
		Data: derBytes,
	}

	bytes, err := asn1.Marshal(certBag1)
	if err != nil {
		return nil, EncodeError("encoding cert bag: " + err.Error())
	}

	certSafeBags, err := makeSafeBags(oidCertBag, bytes)
	if err != nil {
		return nil, EncodeError("safe bags: " + err.Error())
	}

	return makeContentInfo(certSafeBags)
}

func makeShroudedKeyBagContentInfo(privateKey interface{}, password []byte) (*contentInfo, error) {
	shroudedKeyBagBytes, err := encodePkcs8ShroudedKeyBag(privateKey, password)
	if err != nil {
		return nil, EncodeError("encode PKCS#8 shrouded key bag: " + err.Error())
	}

	safeBags, err := makeSafeBags(oidPKCS8ShroudedKeyBag, shroudedKeyBagBytes)
	if err != nil {
		return nil, EncodeError("safe bags: " + err.Error())
	}

	return makeContentInfo(safeBags)
}

func makeContentInfo(val interface{}) (*contentInfo, error) {
	fullBytes, err := asn1.Marshal(val)
	if err != nil {
		return nil, EncodeError("contentInfo raw value marshal: " + err.Error())
	}

	octetStringVal := asn1.RawValue{Tag: 4, Class: 0, IsCompound: false, Bytes: fullBytes}
	octetStringFullBytes, err := asn1.Marshal(octetStringVal)
	if err != nil {
		return nil, EncodeError("raw contentInfo to octet string: " + err.Error())
	}

	contentInfo := contentInfo{ContentType: oidDataContentType}
	contentInfo.Content = asn1.RawValue{Tag: 0, Class: 2, IsCompound: true, Bytes: octetStringFullBytes}

	return &contentInfo, nil
}

func makeContentInfos(derBytes []byte, privateKey interface{}, password []byte) ([]contentInfo, error) {
	shroudedKeyContentInfo, err := makeShroudedKeyBagContentInfo(privateKey, password)
	if err != nil {
		return nil, EncodeError("shrouded key content info: " + err.Error())
	}

	certBagContentInfo, err := makeCertBagContentInfo(derBytes)
	if err != nil {
		return nil, EncodeError("cert bag content info: " + err.Error())
	}

	contentInfos := make([]contentInfo, 2)
	contentInfos[0] = *shroudedKeyContentInfo
	contentInfos[1] = *certBagContentInfo

	return contentInfos, nil
}

func makeSalt(saltByteCount int) ([]byte, error) {
	salt := make([]byte, saltByteCount)
	_, err := io.ReadFull(rand.Reader, salt)
	return salt, err
}

// Encode converts a certificate and a private key to the PKCS#12 byte stream format.
//
// derBytes is a DER encoded certificate.
// privateKey is an RSA
func Encode(derBytes []byte, privateKey interface{}, password string) (pfxBytes []byte, err error) {
	secret, err := bmpString(password)
	if err != nil {
		return nil, ErrIncorrectPassword
	}

	contentInfos, err := makeContentInfos(derBytes, privateKey, secret)
	if err != nil {
		return nil, err
	}

	// Marhsal []contentInfo so we can re-constitute the byte stream that will
	// be suitable for computing the MAC
	bytes, err := asn1.Marshal(contentInfos)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an asn1.RawValue so, we can compute the MAC against the .Bytes
	var contentInfosRaw asn1.RawValue
	err = unmarshal(bytes, &contentInfosRaw)
	if err != nil {
		return nil, err
	}

	authSafeContentInfo, err := makeContentInfo(contentInfosRaw)
	if err != nil {
		return nil, EncodeError("authSafe content info: " + err.Error())
	}

	salt, err := makeSalt(pbeSaltSizeBytes)
	if err != nil {
		return nil, EncodeError("salt value: " + err.Error())
	}

	// Compute the MAC for marshaled bytes of contentInfos, which includes the
	// cert bag, and the shrouded key bag.
	digest := computeMac(contentInfosRaw.FullBytes, pbeIterationCount, salt, secret)

	pfx := pfxPdu{
		Version:  3,
		AuthSafe: *authSafeContentInfo,
		MacData: macData{
			Iterations: pbeIterationCount,
			MacSalt:    salt,
			Mac: digestInfo{
				Algorithm: pkix.AlgorithmIdentifier{
					Algorithm: oidSHA1,
				},
				Digest: digest,
			},
		},
	}

	bytes, err = asn1.Marshal(pfx)
	if err != nil {
		return nil, EncodeError("marshal PFX PDU: " + err.Error())
	}

	return bytes, err
}

func getSafeContents(p12Data, password []byte) (bags []safeBag, updatedPassword []byte, err error) {
	pfx := new(pfxPdu)

	if err := unmarshal(p12Data, pfx); err != nil {
		return nil, nil, errors.New("pkcs12: error reading P12 data: " + err.Error())
	}

	if pfx.Version != 3 {
		return nil, nil, NotImplementedError("can only decode v3 PFX PDU's")
	}

	if !pfx.AuthSafe.ContentType.Equal(oidDataContentType) {
		return nil, nil, NotImplementedError("only password-protected PFX is implemented")
	}

	// unmarshal the explicit bytes in the content for type 'data'
	if err := unmarshal(pfx.AuthSafe.Content.Bytes, &pfx.AuthSafe.Content); err != nil {
		return nil, nil, err
	}

	if len(pfx.MacData.Mac.Algorithm.Algorithm) == 0 {
		return nil, nil, errors.New("pkcs12: no MAC in data")
	}

	if err := verifyMac(&pfx.MacData, pfx.AuthSafe.Content.Bytes, password); err != nil {
		if err == ErrIncorrectPassword && len(password) == 2 && password[0] == 0 && password[1] == 0 {
			// some implementations use an empty byte array
			// for the empty string password try one more
			// time with empty-empty password
			password = nil
			err = verifyMac(&pfx.MacData, pfx.AuthSafe.Content.Bytes, password)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	var authenticatedSafe []contentInfo
	if err := unmarshal(pfx.AuthSafe.Content.Bytes, &authenticatedSafe); err != nil {
		return nil, nil, err
	}

	if len(authenticatedSafe) != 2 {
		return nil, nil, NotImplementedError("expected exactly two items in the authenticated safe")
	}

	for _, ci := range authenticatedSafe {
		var data []byte

		switch {
		case ci.ContentType.Equal(oidDataContentType):
			if err := unmarshal(ci.Content.Bytes, &data); err != nil {
				return nil, nil, err
			}
		case ci.ContentType.Equal(oidEncryptedDataContentType):
			var encryptedData encryptedData
			if err := unmarshal(ci.Content.Bytes, &encryptedData); err != nil {
				return nil, nil, err
			}
			if encryptedData.Version != 0 {
				return nil, nil, NotImplementedError("only version 0 of EncryptedData is supported")
			}
			if data, err = pbDecrypt(encryptedData.EncryptedContentInfo, password); err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, NotImplementedError("only data and encryptedData content types are supported in authenticated safe")
		}

		var safeContents []safeBag
		if err := unmarshal(data, &safeContents); err != nil {
			return nil, nil, err
		}

		bags = append(bags, safeContents...)
	}

	return bags, password, nil
}
