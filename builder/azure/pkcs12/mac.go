package pkcs12

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/x509/pkix"
	"encoding/asn1"
)

var (
	oidSha1Algorithm = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
)

type macData struct {
	Mac        digestInfo
	MacSalt    []byte
	Iterations int `asn1:"optional,default:1"`
}

// from PKCS#7:
type digestInfo struct {
	Algorithm pkix.AlgorithmIdentifier
	Digest    []byte
}

func computeMac(message []byte, iterations int, salt, password []byte) []byte {
	key := pbkdf(sha1Sum, 20, 64, salt, password, iterations, 3, 20)

	mac := hmac.New(sha1.New, key)
	mac.Write(message)

	return mac.Sum(nil)
}
