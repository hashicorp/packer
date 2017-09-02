// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkcs12

import (
	"bytes"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"io"
	"testing"
)

var sha1WithTripleDES = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 3})

func TestPbDecrypterFor(t *testing.T) {
	params, _ := asn1.Marshal(pbeParams{
		Salt:       []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Iterations: 2048,
	})
	alg := pkix.AlgorithmIdentifier{
		Algorithm: asn1.ObjectIdentifier([]int{1, 2, 3}),
		Parameters: asn1.RawValue{
			FullBytes: params,
		},
	}

	pass, _ := bmpString("Sesame open")

	_, _, err := pbDecrypterFor(alg, pass)
	if _, ok := err.(NotImplementedError); !ok {
		t.Errorf("expected not implemented error, got: %T %s", err, err)
	}

	alg.Algorithm = sha1WithTripleDES
	cbc, blockSize, err := pbDecrypterFor(alg, pass)
	if err != nil {
		t.Errorf("unexpected error from pbDecrypterFor %v", err)
	}
	if blockSize != 8 {
		t.Errorf("unexpected block size %d, wanted 8", blockSize)
	}

	plaintext := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	expectedCiphertext := []byte{185, 73, 135, 249, 137, 1, 122, 247}
	ciphertext := make([]byte, len(plaintext))
	cbc.CryptBlocks(ciphertext, plaintext)

	if bytes.Compare(ciphertext, expectedCiphertext) != 0 {
		t.Errorf("bad ciphertext, got %x but wanted %x", ciphertext, expectedCiphertext)
	}
}

var pbDecryptTests = []struct {
	in            []byte
	expected      []byte
	expectedError error
}{
	{
		[]byte("\x33\x73\xf3\x9f\xda\x49\xae\xfc\xa0\x9a\xdf\x5a\x58\xa0\xea\x46"), // 7 padding bytes
		[]byte("A secret!"),
		nil,
	},
	{
		[]byte("\x33\x73\xf3\x9f\xda\x49\xae\xfc\x96\x24\x2f\x71\x7e\x32\x3f\xe7"), // 8 padding bytes
		[]byte("A secret"),
		nil,
	},
	{
		[]byte("\x35\x0c\xc0\x8d\xab\xa9\x5d\x30\x7f\x9a\xec\x6a\xd8\x9b\x9c\xd9"), // 9 padding bytes, incorrect
		nil,
		ErrDecryption,
	},
	{
		[]byte("\xb2\xf9\x6e\x06\x60\xae\x20\xcf\x08\xa0\x7b\xd9\x6b\x20\xef\x41"), // incorrect padding bytes: [ ... 0x04 0x02 ]
		nil,
		ErrDecryption,
	},
}

func TestPbDecrypt(t *testing.T) {
	salt := []byte("\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8")

	for i, test := range pbDecryptTests {
		decryptable := makeTestDecryptable(test.in, salt)
		password, _ := bmpString("sesame")

		plaintext, err := pbDecrypt(decryptable, password)
		if err != test.expectedError {
			t.Errorf("#%d: got error %q, but wanted %q", i, err, test.expectedError)
			continue
		}

		if !bytes.Equal(plaintext, test.expected) {
			t.Errorf("#%d: got %x, but wanted %x", i, plaintext, test.expected)
		}
	}
}

func TestRoundTripPkc12EncryptDecrypt(t *testing.T) {
	salt := []byte{0xfe, 0xee, 0xfa, 0xce}
	password := salt

	// Sweep the possible padding lengths
	for i := 0; i < 9; i++ {
		bs := make([]byte, i)
		_, err := io.ReadFull(rand.Reader, bs)
		if err != nil {
			t.Fatalf("failed to read: %s", err)
		}

		cipherText, err := pbEncrypt(bs, salt, password, 4096)
		if err != nil {
			t.Fatalf("failed to encrypt: %s\n", err)
		}

		if len(cipherText)%8 != 0 {
			t.Fatalf("plain text was not padded as expected")
		}

		decryptable := makeTestDecryptable(cipherText, salt)
		plainText, err := pbDecrypt(decryptable, password)
		if err != nil {
			t.Fatalf("failed to decrypt: %s\n", err)
		}

		if !bytes.Equal(bs, plainText) {
			t.Fatalf("got %x, but wanted %x", bs, plainText)
		}
	}
}

func makeTestDecryptable(bytes, salt []byte) testDecryptable {
	decryptable := testDecryptable{
		data: bytes,
		algorithm: pkix.AlgorithmIdentifier{
			Algorithm: sha1WithTripleDES,
			Parameters: pbeParams{
				Salt:       salt,
				Iterations: 4096,
			}.RawASN1(),
		},
	}

	return decryptable
}

type testDecryptable struct {
	data      []byte
	algorithm pkix.AlgorithmIdentifier
}

func (d testDecryptable) Algorithm() pkix.AlgorithmIdentifier { return d.algorithm }
func (d testDecryptable) Data() []byte                        { return d.data }

func (params pbeParams) RawASN1() (raw asn1.RawValue) {
	asn1Bytes, err := asn1.Marshal(params)
	if err != nil {
		panic(err)
	}
	_, err = asn1.Unmarshal(asn1Bytes, &raw)
	if err != nil {
		panic(err)
	}
	return
}
