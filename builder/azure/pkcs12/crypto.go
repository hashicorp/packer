// implementation of https://tools.ietf.org/html/rfc2898#section-6.1.2

package pkcs12

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"io"

	"github.com/mitchellh/packer/builder/azure/pkcs12/rc2"
)

const (
	pbeWithSHAAnd3KeyTripleDESCBC = "pbeWithSHAAnd3-KeyTripleDES-CBC"
	pbewithSHAAnd40BitRC2CBC      = "pbewithSHAAnd40BitRC2-CBC"
)

const (
	pbeIterationCount = 2048
	pbeSaltSizeBytes  = 8
)

var (
	oidPbeWithSHAAnd3KeyTripleDESCBC = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 12, 1, 3}
	oidPbewithSHAAnd40BitRC2CBC      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 12, 1, 6}
)

var algByOID = map[string]string{
	oidPbeWithSHAAnd3KeyTripleDESCBC.String(): pbeWithSHAAnd3KeyTripleDESCBC,
	oidPbewithSHAAnd40BitRC2CBC.String():      pbewithSHAAnd40BitRC2CBC,
}

var blockcodeByAlg = map[string]func(key []byte) (cipher.Block, error){
	pbeWithSHAAnd3KeyTripleDESCBC: des.NewTripleDESCipher,
	pbewithSHAAnd40BitRC2CBC: func(key []byte) (cipher.Block, error) {
		return rc2.New(key, len(key)*8)
	},
}

type pbeParams struct {
	Salt       []byte
	Iterations int
}

func pad(src []byte, blockSize int) []byte {
	paddingLength := blockSize - len(src)%blockSize
	paddingText := bytes.Repeat([]byte{byte(paddingLength)}, paddingLength)
	return append(src, paddingText...)
}

func pbEncrypt(plainText, salt, password []byte, iterations int) (cipherText []byte, err error) {
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, errors.New("pkcs12: failed to create a random salt value: " + err.Error())
	}

	key := deriveKeyByAlg[pbeWithSHAAnd3KeyTripleDESCBC](salt, password, iterations)
	iv := deriveIVByAlg[pbeWithSHAAnd3KeyTripleDESCBC](salt, password, iterations)

	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, errors.New("pkcs12: failed to create a block cipher: " + err.Error())
	}

	paddedPlainText := pad(plainText, block.BlockSize())

	encrypter := cipher.NewCBCEncrypter(block, iv)
	cipherText = make([]byte, len(paddedPlainText))
	encrypter.CryptBlocks(cipherText, paddedPlainText)

	return cipherText, nil
}

type decryptable interface {
	GetAlgorithm() pkix.AlgorithmIdentifier
	GetData() []byte
}
