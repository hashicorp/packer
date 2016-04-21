package pkcs12

import (
	"bytes"
	"testing"
)

func TestThatPBKDFWorksCorrectlyForLongKeys(t *testing.T) {
	pbkdf := deriveKeyByAlg[pbeWithSHAAnd3KeyTripleDESCBC]

	salt := []byte("\xff\xff\xff\xff\xff\xff\xff\xff")
	password, _ := bmpString("sesame")
	key := pbkdf(salt, password, 2048)

	if expected := []byte("\x7c\xd9\xfd\x3e\x2b\x3b\xe7\x69\x1a\x44\xe3\xbe\xf0\xf9\xea\x0f\xb9\xb8\x97\xd4\xe3\x25\xd9\xd1"); bytes.Compare(key, expected) != 0 {
		t.Fatalf("expected key '% x', but found '% x'", key, expected)
	}
}
