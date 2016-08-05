package pkcs12

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"testing"
)

func decodePkcs8ShroudedKeyBag(asn1Data, password []byte) (privateKey interface{}, err error) {
	pkinfo := new(encryptedPrivateKeyInfo)
	if _, err = asn1.Unmarshal(asn1Data, pkinfo); err != nil {
		err = fmt.Errorf("error decoding PKCS8 shrouded key bag: %v", err)
		return nil, err
	}

	pkData, err := pbDecrypt(pkinfo, password)
	if err != nil {
		err = fmt.Errorf("error decrypting PKCS8 shrouded key bag: %v", err)
		return
	}

	rv := new(asn1.RawValue)
	if _, err = asn1.Unmarshal(pkData, rv); err != nil {
		err = fmt.Errorf("could not decode decrypted private key data")
	}

	if privateKey, err = x509.ParsePKCS8PrivateKey(pkData); err != nil {
		err = fmt.Errorf("error parsing PKCS8 private key: %v", err)
		return nil, err
	}
	return
}

// Assert the default algorithm parameters are in the correct order,
// and default to the correct value.  Defaults are based on OpenSSL.
//  1. IterationCount, defaults to 2,048 long.
//  2. Salt, is 8 bytes long.
func TestDefaultAlgorithmParametersPkcs8ShroudedKeyBag(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	password := []byte("sesame")
	bytes, err := encodePkcs8ShroudedKeyBag(privateKey, password)
	if err != nil {
		t.Fatalf("failed to encode PKCS#8 shrouded key bag: %s", err)
	}

	var pkinfo encryptedPrivateKeyInfo
	rest, err := asn1.Unmarshal(bytes, &pkinfo)
	if err != nil {
		t.Fatalf("failed to unmarshal encryptedPrivateKeyInfo %s", err)
	}

	if len(rest) != 0 {
		t.Fatalf("unexpected trailing bytes of len=%d, bytes=%x", len(rest), rest)
	}

	var params pbeParams
	rest, err = asn1.Unmarshal(pkinfo.GetAlgorithm().Parameters.FullBytes, &params)
	if err != nil {
		t.Fatalf("failed to unmarshal encryptedPrivateKeyInfo %s", err)
	}

	if len(rest) != 0 {
		t.Fatalf("unexpected trailing bytes of len=%d, bytes=%x", len(rest), rest)
	}

	if params.Iterations != pbeIterationCount {
		t.Errorf("expected iteration count to be %d, but actual=%d", pbeIterationCount, params.Iterations)
	}
	if len(params.Salt) != pbeSaltSizeBytes {
		t.Errorf("expected the number of salt bytes to be %d, but actual=%d", pbeSaltSizeBytes, len(params.Salt))
	}
}

func TestRoundTripPkcs8ShroudedKeyBag(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	password := []byte("sesame")
	bytes, err := encodePkcs8ShroudedKeyBag(privateKey, password)
	if err != nil {
		t.Fatalf("failed to encode PKCS#8 shrouded key bag: %s", err)
	}

	key, err := decodePkcs8ShroudedKeyBag(bytes, password)
	if err != nil {
		t.Fatalf("failed to decode PKCS#8 shrouded key bag: %s", err)
	}

	actualPrivateKey := key.(*rsa.PrivateKey)
	if actualPrivateKey.D.Cmp(privateKey.D) != 0 {
		t.Fatal("failed to round-trip rsa.PrivateKey.D")
	}
}
