package pkcs12

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"testing"
	"time"

	gopkcs12 "golang.org/x/crypto/pkcs12"
)

func TestPfxRoundTriRsa(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal(err.Error())
	}

	key := testPfxRoundTrip(t, privateKey)

	actualPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		t.Fatal("failed to decode private key")
	}

	if privateKey.D.Cmp(actualPrivateKey.D) != 0 {
		t.Errorf("priv.D")
	}
}

func TestPfxRoundTriEcdsa(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}

	key := testPfxRoundTrip(t, privateKey)

	actualPrivateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("failed to decode private key")
	}

	if privateKey.D.Cmp(actualPrivateKey.D) != 0 {
		t.Errorf("priv.D")
	}
}

func testPfxRoundTrip(t *testing.T, privateKey interface{}) interface{} {
	certificateBytes, err := newCertificate("hostname", privateKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	bytes, err := Encode(certificateBytes, privateKey, "sesame")
	if err != nil {
		t.Fatal(err.Error())
	}

	key, _, err := gopkcs12.Decode(bytes, "sesame")
	if err != nil {
		t.Fatalf(err.Error())
	}

	return key
}

func newCertificate(hostname string, privateKey interface{}) ([]byte, error) {
	t, _ := time.Parse("2006-01-02", "2016-01-01")
	notBefore := t
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		err := fmt.Errorf("Failed to Generate Serial Number: %v", err)
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Issuer: pkix.Name{
			CommonName: hostname,
		},
		Subject: pkix.Name{
			CommonName: hostname,
		},

		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	var publicKey interface{}
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = key.Public()
	case *ecdsa.PrivateKey:
		publicKey = key.Public()
	default:
		panic(fmt.Sprintf("unsupported private key type: %T", privateKey))
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to Generate derBytes: " + err.Error())
	}

	return derBytes, nil
}
