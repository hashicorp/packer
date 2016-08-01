// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package pkcs12

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"testing"
)

func TestRoundTripPkcs8Rsa(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	bytes, err := marshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %s", err)
	}

	key, err := x509.ParsePKCS8PrivateKey(bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %s", err)
	}

	actualPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("expected key to be of type *rsa.PrivateKey, but actual was %T", key)
	}

	if actualPrivateKey.Validate() != nil {
		t.Fatal("private key did not validate")
	}

	if actualPrivateKey.N.Cmp(privateKey.N) != 0 {
		t.Errorf("private key's N did not round trip")
	}
	if actualPrivateKey.D.Cmp(privateKey.D) != 0 {
		t.Errorf("private key's D did not round trip")
	}
	if actualPrivateKey.E != privateKey.E {
		t.Errorf("private key's E did not round trip")
	}
	if actualPrivateKey.Primes[0].Cmp(privateKey.Primes[0]) != 0 {
		t.Errorf("private key's P did not round trip")
	}
	if actualPrivateKey.Primes[1].Cmp(privateKey.Primes[1]) != 0 {
		t.Errorf("private key's Q did not round trip")
	}
}

func TestRoundTripPkcs8Ecdsa(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	bytes, err := marshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %s", err)
	}

	key, err := x509.ParsePKCS8PrivateKey(bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %s", err)
	}

	actualPrivateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("expected key to be of type *ecdsa.PrivateKey, but actual was %T", key)
	}

	// sanity check, not exhaustive
	if actualPrivateKey.D.Cmp(privateKey.D) != 0 {
		t.Errorf("private key's D did not round trip")
	}
	if actualPrivateKey.X.Cmp(privateKey.X) != 0 {
		t.Errorf("private key's X did not round trip")
	}
	if actualPrivateKey.Y.Cmp(privateKey.Y) != 0 {
		t.Errorf("private key's Y did not round trip")
	}
	if actualPrivateKey.Curve.Params().B.Cmp(privateKey.Curve.Params().B) != 0 {
		t.Errorf("private key's Curve.B did not round trip")
	}
}

func TestNullParametersPkcs8Rsa(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	checkNullParameter(t, privateKey)
}

func TestNullParametersPkcs8Ecdsa(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate a private key: %s", err)
	}

	checkNullParameter(t, privateKey)
}

func checkNullParameter(t *testing.T, privateKey interface{}) {
	bytes, err := marshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %s", err)
	}

	var pkcs pkcs8
	rest, err := asn1.Unmarshal(bytes, &pkcs)
	if err != nil {
		t.Fatalf("failed to unmarshal PKCS#8: %s", err)
	}

	if len(rest) != 0 {
		t.Fatalf("unexpected trailing bytes of len=%d, bytes=%x", len(rest), rest)
	}

	// Only version == 0 is known and valid
	if pkcs.Version != 0 {
		t.Errorf("expected version=0, but actual=%d", pkcs.Version)
	}

	// ensure a NULL parameter is inserted
	if pkcs.Algo.Parameters.Tag != 5 {
		t.Errorf("expected parameters to be NULL, but actual tag=%d, class=%d, isCompound=%t, bytes=%x",
			pkcs.Algo.Parameters.Tag,
			pkcs.Algo.Parameters.Class,
			pkcs.Algo.Parameters.IsCompound,
			pkcs.Algo.Parameters.Bytes)
	}
}
