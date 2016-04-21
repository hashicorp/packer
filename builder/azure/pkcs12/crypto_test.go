package pkcs12

import (
	"bytes"
	"crypto/cipher"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
)

func pbDecrypterFor(algorithm pkix.AlgorithmIdentifier, password []byte) (cipher.BlockMode, error) {
	algorithmName, supported := algByOID[algorithm.Algorithm.String()]
	if !supported {
		return nil, NotImplementedError("algorithm " + algorithm.Algorithm.String() + " is not supported")
	}

	var params pbeParams
	if _, err := asn1.Unmarshal(algorithm.Parameters.FullBytes, &params); err != nil {
		return nil, err
	}

	k := deriveKeyByAlg[algorithmName](params.Salt, password, params.Iterations)
	iv := deriveIVByAlg[algorithmName](params.Salt, password, params.Iterations)
	password = nil

	code, err := blockcodeByAlg[algorithmName](k)
	if err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(code, iv)
	return cbc, nil
}

func pbDecrypt(info decryptable, password []byte) (decrypted []byte, err error) {
	cbc, err := pbDecrypterFor(info.GetAlgorithm(), password)
	password = nil
	if err != nil {
		return nil, err
	}

	encrypted := info.GetData()

	decrypted = make([]byte, len(encrypted))
	cbc.CryptBlocks(decrypted, encrypted)

	if psLen := int(decrypted[len(decrypted)-1]); psLen > 0 && psLen <= cbc.BlockSize() {
		m := decrypted[:len(decrypted)-psLen]
		ps := decrypted[len(decrypted)-psLen:]
		if bytes.Compare(ps, bytes.Repeat([]byte{byte(psLen)}, psLen)) != 0 {
			return nil, ErrDecryption
		}
		decrypted = m
	} else {
		return nil, ErrDecryption
	}

	return
}

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

	_, err := pbDecrypterFor(alg, pass)
	if _, ok := err.(NotImplementedError); !ok {
		t.Errorf("expected not implemented error, got: %T %s", err, err)
	}

	alg.Algorithm = asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 3})
	cbc, err := pbDecrypterFor(alg, pass)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	M := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	expectedM := []byte{185, 73, 135, 249, 137, 1, 122, 247}
	cbc.CryptBlocks(M, M)

	if bytes.Compare(M, expectedM) != 0 {
		t.Errorf("expected M to be '%d', but found '%d", expectedM, M)
	}
}

func TestPbDecrypt(t *testing.T) {

	tests := [][]byte{
		[]byte("\x33\x73\xf3\x9f\xda\x49\xae\xfc\xa0\x9a\xdf\x5a\x58\xa0\xea\x46"), // 7 padding bytes
		[]byte("\x33\x73\xf3\x9f\xda\x49\xae\xfc\x96\x24\x2f\x71\x7e\x32\x3f\xe7"), // 8 padding bytes
		[]byte("\x35\x0c\xc0\x8d\xab\xa9\x5d\x30\x7f\x9a\xec\x6a\xd8\x9b\x9c\xd9"), // 9 padding bytes, incorrect
		[]byte("\xb2\xf9\x6e\x06\x60\xae\x20\xcf\x08\xa0\x7b\xd9\x6b\x20\xef\x41"), // incorrect padding bytes: [ ... 0x04 0x02 ]
	}
	expected := []interface{}{
		[]byte("A secret!"),
		[]byte("A secret"),
		ErrDecryption,
		ErrDecryption,
	}

	for i, c := range tests {
		td := testDecryptable{
			data: c,
			algorithm: pkix.AlgorithmIdentifier{
				Algorithm: asn1.ObjectIdentifier([]int{1, 2, 840, 113549, 1, 12, 1, 3}), // SHA1/3TDES
				Parameters: pbeParams{
					Salt:       []byte("\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8"),
					Iterations: 4096,
				}.RawASN1(),
			},
		}
		p, _ := bmpString("sesame")

		m, err := pbDecrypt(td, p)

		switch e := expected[i].(type) {
		case []byte:
			if err != nil {
				t.Errorf("error decrypting C=%x: %v", c, err)
			}
			if bytes.Compare(m, e) != 0 {
				t.Errorf("expected C=%x to be decoded to M=%x, but found %x", c, e, m)
			}
		case error:
			if err == nil || err.Error() != e.Error() {
				t.Errorf("expecting error '%v' during decryption of c=%x, but found err='%v'", e, c, err)
			}
		}
	}
}

type testDecryptable struct {
	data      []byte
	algorithm pkix.AlgorithmIdentifier
}

func (d testDecryptable) GetAlgorithm() pkix.AlgorithmIdentifier { return d.algorithm }
func (d testDecryptable) GetData() []byte                        { return d.data }

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
