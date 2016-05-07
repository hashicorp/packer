package pkcs12

import (
	"errors"
	"unicode/utf16"
)

func bmpString(s string) ([]byte, error) {
	// References:
	// https://tools.ietf.org/html/rfc7292#appendix-B.1
	// http://en.wikipedia.org/wiki/Plane_(Unicode)#Basic_Multilingual_Plane
	//  - non-BMP characters are encoded in UTF 16 by using a surrogate pair of 16-bit codes
	//	  EncodeRune returns 0xfffd if the rune does not need special encoding
	//  - the above RFC provides the info that BMPStrings are NULL terminated.

	rv := make([]byte, 0, 2*len(s)+2)

	for _, r := range s {
		if t, _ := utf16.EncodeRune(r); t != 0xfffd {
			return nil, errors.New("string contains characters that cannot be encoded in UCS-2")
		}
		rv = append(rv, byte(r/256), byte(r%256))
	}
	rv = append(rv, 0, 0)
	return rv, nil
}
