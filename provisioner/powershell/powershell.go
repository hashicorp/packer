package powershell

import (
	"encoding/base64"
)

func powershellEncode(buffer []byte) string {
	// 2 byte chars to make PowerShell happy
	wideCmd := ""
	for _, b := range buffer {
		wideCmd += string(b) + "\x00"
	}

	// Base64 encode the command
	input := []uint8(wideCmd)
	return base64.StdEncoding.EncodeToString(input)
}

func powershellDecode(message string) (retour string) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	base64.StdEncoding.Decode(base64Text, []byte(message))
	return string(base64Text)
}
