package oci

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNewConfigMissingFile(t *testing.T) {
	// WHEN
	_, err := LoadConfigsFromFile("some/invalid/path")

	// THEN

	if err == nil {
		t.Error("Expected missing file error")
	}
}

func TestNewConfigDefaultOnly(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	// Load configs
	cfgs, err := LoadConfigsFromFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	// THEN

	if _, ok := cfgs["DEFAULT"]; !ok {
		t.Fatal("Expected DEFAULT config to exist in map")
	}
}

func TestNewConfigDefaultsPopulated(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	admin := cfg.Section("ADMIN")
	admin.NewKey("user", "ocid1.user.oc1..bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	admin.NewKey("fingerprint", "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11")

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	cfgs, err := LoadConfigsFromFile(f.Name())
	adminConfig, ok := cfgs["ADMIN"]

	// THEN

	if !ok {
		t.Fatal("Expected ADMIN config to exist in map")
	}

	if adminConfig.Region != "us-ashburn-1" {
		t.Errorf("Expected 'us-ashburn-1', got '%s'", adminConfig.Region)
	}
}

func TestNewConfigDefaultsOverridden(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	admin := cfg.Section("ADMIN")
	admin.NewKey("user", "ocid1.user.oc1..bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	admin.NewKey("fingerprint", "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11")

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	cfgs, err := LoadConfigsFromFile(f.Name())
	adminConfig, ok := cfgs["ADMIN"]

	// THEN

	if !ok {
		t.Fatal("Expected ADMIN config to exist in map")
	}

	if adminConfig.Fingerprint != "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11" {
		t.Errorf("Expected fingerprint '11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11', got '%s'",
			adminConfig.Fingerprint)
	}
}

func TestParseEncryptedPrivateKeyValidPassword(t *testing.T) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		t.Fatalf("Unexpected generating RSA key: %+v", err)
	}
	publicKey := priv.PublicKey

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)

	blockType := "RSA PRIVATE KEY"
	password := []byte("password")
	cipherType := x509.PEMCipherAES256

	// Encrypt priv with password
	encryptedPEMBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		blockType,
		privDer,
		password,
		cipherType)
	if err != nil {
		t.Fatalf("Unexpected error encryting PEM block: %+v", err)
	}

	// Parse private key
	key, err := ParsePrivateKey(pem.EncodeToMemory(encryptedPEMBlock), password)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	// Check we get the same key back
	if !reflect.DeepEqual(publicKey, key.PublicKey) {
		t.Errorf("expected public key of encrypted and decrypted key to match")
	}
}

func TestParseEncryptedPrivateKeyPKCS8(t *testing.T) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		t.Fatalf("Unexpected generating RSA key: %+v", err)
	}
	publicKey := priv.PublicKey

	// Implements x509.MarshalPKCS8PrivateKey which is not included in the
	// standard library.
	pkey := struct {
		Version             int
		PrivateKeyAlgorithm []asn1.ObjectIdentifier
		PrivateKey          []byte
	}{
		Version:             0,
		PrivateKeyAlgorithm: []asn1.ObjectIdentifier{{1, 2, 840, 113549, 1, 1, 1}},
		PrivateKey:          x509.MarshalPKCS1PrivateKey(priv),
	}
	privDer, err := asn1.Marshal(pkey)
	if err != nil {
		t.Fatalf("Unexpected marshaling RSA key: %+v", err)
	}

	blockType := "RSA PRIVATE KEY"
	password := []byte("password")
	cipherType := x509.PEMCipherAES256

	// Encrypt priv with password
	encryptedPEMBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		blockType,
		privDer,
		password,
		cipherType)
	if err != nil {
		t.Fatalf("Unexpected error encryting PEM block: %+v", err)
	}

	// Parse private key
	key, err := ParsePrivateKey(pem.EncodeToMemory(encryptedPEMBlock), password)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	// Check we get the same key back
	if !reflect.DeepEqual(publicKey, key.PublicKey) {
		t.Errorf("expected public key of encrypted and decrypted key to match")
	}
}

func TestParseEncryptedPrivateKeyInvalidPassword(t *testing.T) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		t.Fatalf("Unexpected generating RSA key: %+v", err)
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)

	blockType := "RSA PRIVATE KEY"
	password := []byte("password")
	cipherType := x509.PEMCipherAES256

	// Encrypt priv with password
	encryptedPEMBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		blockType,
		privDer,
		password,
		cipherType)
	if err != nil {
		t.Fatalf("Unexpected error encrypting PEM block: %+v", err)
	}

	// Parse private key (with wrong password)
	_, err = ParsePrivateKey(pem.EncodeToMemory(encryptedPEMBlock), []byte("foo"))
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "decryption password incorrect") {
		t.Errorf("Expected error to contain 'decryption password incorrect', got %+v", err)
	}
}

func TestParseEncryptedPrivateKeyInvalidNoPassword(t *testing.T) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		t.Fatalf("Unexpected generating RSA key: %+v", err)
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)

	blockType := "RSA PRIVATE KEY"
	password := []byte("password")
	cipherType := x509.PEMCipherAES256

	// Encrypt priv with password
	encryptedPEMBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		blockType,
		privDer,
		password,
		cipherType)
	if err != nil {
		t.Fatalf("Unexpected error encrypting PEM block: %+v", err)
	}

	// Parse private key (with wrong password)
	_, err = ParsePrivateKey(pem.EncodeToMemory(encryptedPEMBlock), []byte{})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "no pass phrase provided") {
		t.Errorf("Expected error to contain 'no pass phrase provided', got %+v", err)
	}
}
