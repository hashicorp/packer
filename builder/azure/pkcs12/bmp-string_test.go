package pkcs12

import (
	"bytes"
	"errors"
	"testing"
	"unicode/utf16"
)

func decodeBMPString(bmpString []byte) (string, error) {
	if len(bmpString)%2 != 0 {
		return "", errors.New("expected BMP byte string to be an even length")
	}

	// strip terminator if present
	if terminator := bmpString[len(bmpString)-2:]; terminator[0] == terminator[1] && terminator[1] == 0 {
		bmpString = bmpString[:len(bmpString)-2]
	}

	s := make([]uint16, 0, len(bmpString)/2)
	for len(bmpString) > 0 {
		s = append(s, uint16(bmpString[0])*265+uint16(bmpString[1]))
		bmpString = bmpString[2:]
	}

	return string(utf16.Decode(s)), nil
}

func TestBMPStringDecode(t *testing.T) {
	_, err := decodeBMPString([]byte("a"))
	if err == nil {
		t.Fatal("expected decode to fail, but it succeeded")
	}
}

func TestBMPString(t *testing.T) {
	str, err := bmpString("")
	if !bytes.Equal(str, []byte{0, 0}) {
		t.Errorf("expected empty string to return double 0, but found: % x", str)
	}
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// Example from https://tools.ietf.org/html/rfc7292#appendix-B
	str, err = bmpString("Beavis")
	if !bytes.Equal(str, []byte{0x00, 0x42, 0x00, 0x65, 0x00, 0x61, 0x00, 0x0076, 0x00, 0x69, 0x00, 0x73, 0x00, 0x00}) {
		t.Errorf("expected 'Beavis' to return 0x00 0x42 0x00 0x65 0x00 0x61 0x00 0x76 0x00 0x69 0x00 0x73 0x00 0x00, but found: % x", str)
	}
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// some characters from the "Letterlike Symbols Unicode block"
	tst := "\u2115 - Double-struck N"
	str, err = bmpString(tst)
	if !bytes.Equal(str, []byte{0x21, 0x15, 0x00, 0x20, 0x00, 0x2d, 0x00, 0x20, 0x00, 0x44, 0x00, 0x6f, 0x00, 0x75, 0x00, 0x62, 0x00, 0x6c, 0x00, 0x65, 0x00, 0x2d, 0x00, 0x73, 0x00, 0x74, 0x00, 0x72, 0x00, 0x75, 0x00, 0x63, 0x00, 0x6b, 0x00, 0x20, 0x00, 0x4e, 0x00, 0x00}) {
		t.Errorf("expected '%s' to return 0x21 0x15 0x00 0x20 0x00 0x2d 0x00 0x20 0x00 0x44 0x00 0x6f 0x00 0x75 0x00 0x62 0x00 0x6c 0x00 0x65 0x00 0x2d 0x00 0x73 0x00 0x74 0x00 0x72 0x00 0x75 0x00 0x63 0x00 0x6b 0x00 0x20 0x00 0x4e 0x00 0x00, but found: % x", tst, str)
	}
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// some character outside the BMP should error
	tst = "\U0001f000 East wind (Mahjong)"
	_, err = bmpString(tst)
	if err == nil {
		t.Errorf("expected '%s' to throw error because the first character is not in the BMP", tst)
	}
}
