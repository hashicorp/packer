package pkcs12

import (
	"crypto/hmac"
	"encoding/asn1"
	"testing"
)

func verifyMac(macData *macData, message, password []byte) error {
	if !macData.Mac.Algorithm.Algorithm.Equal(oidSha1Algorithm) {
		return NotImplementedError("unknown digest algorithm: " + macData.Mac.Algorithm.Algorithm.String())
	}

	expectedMAC := computeMac(message, macData.Iterations, macData.MacSalt, password)

	if !hmac.Equal(macData.Mac.Digest, expectedMAC) {
		return ErrIncorrectPassword
	}
	return nil
}

func TestVerifyMac(t *testing.T) {
	td := macData{
		Mac: digestInfo{
			Digest: []byte{0x18, 0x20, 0x3d, 0xff, 0x1e, 0x16, 0xf4, 0x92, 0xf2, 0xaf, 0xc8, 0x91, 0xa9, 0xba, 0xd6, 0xca, 0x9d, 0xee, 0x51, 0x93},
		},
		MacSalt:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Iterations: 2048,
	}

	message := []byte{11, 12, 13, 14, 15}
	password, _ := bmpString("")

	td.Mac.Algorithm.Algorithm = asn1.ObjectIdentifier([]int{1, 2, 3})
	err := verifyMac(&td, message, password)
	if _, ok := err.(NotImplementedError); !ok {
		t.Errorf("err: %v", err)
	}

	td.Mac.Algorithm.Algorithm = asn1.ObjectIdentifier([]int{1, 3, 14, 3, 2, 26})
	err = verifyMac(&td, message, password)
	if err != ErrIncorrectPassword {
		t.Errorf("Expected incorrect password, got err: %v", err)
	}

	password, _ = bmpString("Sesame open")
	err = verifyMac(&td, message, password)
	if err != nil {
		t.Errorf("err: %v", err)
	}

}
