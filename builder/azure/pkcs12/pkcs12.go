// Package pkcs12 provides some implementations of PKCS#12.
//
// This implementation is distilled from https://tools.ietf.org/html/rfc7292 and referenced documents.
// It is intended for decoding P12/PFX-stored certificate+key for use with the crypto/tls package.
package pkcs12

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"io"
)

var (
	oidLocalKeyID      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 21}
	oidDataContentType = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}

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

func (i encryptedContentInfo) GetAlgorithm() pkix.AlgorithmIdentifier {
	return i.ContentEncryptionAlgorithm
}

func (i encryptedContentInfo) GetData() []byte { return i.EncryptedContent }

type safeBag struct {
	Id         asn1.ObjectIdentifier
	Value      asn1.RawValue     `asn1:"tag:0,explicit"`
	Attributes []pkcs12Attribute `asn1:"set,optional"`
}

type pkcs12Attribute struct {
	Id    asn1.ObjectIdentifier
	Value asn1.RawValue `ans1:"set"`
}

type encryptedPrivateKeyInfo struct {
	AlgorithmIdentifier pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

func (i encryptedPrivateKeyInfo) GetAlgorithm() pkix.AlgorithmIdentifier { return i.AlgorithmIdentifier }
func (i encryptedPrivateKeyInfo) GetData() []byte                        { return i.EncryptedData }

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

	certSafeBags, err := makeSafeBags(oidCertBagType, bytes)
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

	safeBags, err := makeSafeBags(oidPkcs8ShroudedKeyBagType, shroudedKeyBagBytes)
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
func Encode(derBytes []byte, privateKey interface{}, password string) ([]byte, error) {
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
					Algorithm: oidSha1Algorithm,
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
