package pkcs12

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
)

//see https://tools.ietf.org/html/rfc7292#appendix-D
var (
	oidPkcs8ShroudedKeyBagType = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 12, 10, 1, 2}
	oidCertBagType             = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 12, 10, 1, 3}

	oidCertTypeX509Certificate = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 22, 1}
)

type certBag struct {
	Id   asn1.ObjectIdentifier
	Data []byte `asn1:"tag:0,explicit"`
}

func getAlgorithmParams(salt []byte, iterations int) (asn1.RawValue, error) {
	params := pbeParams{
		Salt:       salt,
		Iterations: iterations,
	}

	return convertToRawVal(params)
}

func encodePkcs8ShroudedKeyBag(privateKey interface{}, password []byte) (bytes []byte, err error) {
	privateKeyBytes, err := marshalPKCS8PrivateKey(privateKey)

	if err != nil {
		return nil, errors.New("pkcs12: error encoding PKCS#8 private key: " + err.Error())
	}

	salt, err := makeSalt(pbeSaltSizeBytes)
	if err != nil {
		return nil, errors.New("pkcs12: error creating PKCS#8 salt: " + err.Error())
	}

	pkData, err := pbEncrypt(privateKeyBytes, salt, password, pbeIterationCount)
	if err != nil {
		return nil, errors.New("pkcs12: error encoding PKCS#8 shrouded key bag when encrypting cert bag: " + err.Error())
	}

	params, err := getAlgorithmParams(salt, pbeIterationCount)
	if err != nil {
		return nil, errors.New("pkcs12: error encoding PKCS#8 shrouded key bag algorithm's parameters: " + err.Error())
	}

	pkinfo := encryptedPrivateKeyInfo{
		AlgorithmIdentifier: pkix.AlgorithmIdentifier{
			Algorithm:  oidPbeWithSHAAnd3KeyTripleDESCBC,
			Parameters: params,
		},
		EncryptedData: pkData,
	}

	bytes, err = asn1.Marshal(pkinfo)
	if err != nil {
		return nil, errors.New("pkcs12: error encoding PKCS#8 shrouded key bag: " + err.Error())
	}

	return bytes, err
}
