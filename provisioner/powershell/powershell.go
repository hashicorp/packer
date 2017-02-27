package powershell

import (
	"encoding/base64"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
)

func convertUtf8ToUtf16LE(message string) (string, error) {
	utf16le := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	utfEncoder := utf16le.NewEncoder()
	ut16LeEncodedMessage, err := utfEncoder.String(message)

	return ut16LeEncodedMessage, err
}

// UTF16BytesToString converts UTF-16 encoded bytes, in big or little endian byte order,
// to a UTF-8 encoded string.
func UTF16BytesToString(b []byte, o binary.ByteOrder) string {
	utf := make([]uint16, (len(b)+(2-1))/2)
	for i := 0; i+(2-1) < len(b); i += 2 {
		utf[i/2] = o.Uint16(b[i:])
	}
	if len(b)/2 < len(utf) {
		utf[len(utf)-1] = utf8.RuneError
	}
	return string(utf16.Decode(utf))
}

func powershellEncode(message string) (string, error) {
	utf16LEEncodedMessage, err := convertUtf8ToUtf16LE(message)
	if err != nil {
		return "", err
	}

	// Base64 encode the command
	input := []uint8(utf16LEEncodedMessage)
	return base64.StdEncoding.EncodeToString(input), nil
}

func powershellDecode(messageBase64 string) (retour string, err error) {
	messageUtf16LeByteArray, err := base64.StdEncoding.DecodeString(messageBase64)

	if err != nil {
		return "", err
	}

	message := UTF16BytesToString(messageUtf16LeByteArray, binary.LittleEndian)

	return message, nil
}
