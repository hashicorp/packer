package powershell

import (
	"encoding/base64"

	"golang.org/x/text/encoding/unicode"
)

func powershellUtf8(message string) (string, error) {
	utf8 := unicode.UTF8
	utfEncoder := utf8.NewEncoder()
	utf8EncodedMessage, err := utfEncoder.String(message)

	return utf8EncodedMessage, err
}

func powershellEncode(message string) (string, error) {
	utf8EncodedMessage, err := powershellUtf8(message)
	if err != nil {
		return "", err
	}

	// Base64 encode the command
	input := []uint8(utf8EncodedMessage)
	return base64.StdEncoding.EncodeToString(input), nil
}

func powershellDecode(message string) (retour string, err error) {
	data, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
