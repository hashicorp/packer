package vnc

import (
	"net"

	"crypto/des"
	"encoding/binary"
)

// A ClientAuth implements a method of authenticating with a remote server.
type ClientAuth interface {
	// SecurityType returns the byte identifier sent by the server to
	// identify this authentication scheme.
	SecurityType() uint8

	// Handshake is called when the authentication handshake should be
	// performed, as part of the general RFB handshake. (see 7.2.1)
	Handshake(net.Conn) error
}

// ClientAuthNone is the "none" authentication. See 7.2.1
type ClientAuthNone byte

func (*ClientAuthNone) SecurityType() uint8 {
	return 1
}

func (*ClientAuthNone) Handshake(net.Conn) error {
	return nil
}

// PasswordAuth is VNC authentication, 7.2.2
type PasswordAuth struct {
	Password string
}

func (p *PasswordAuth) SecurityType() uint8 {
	return 2
}

func (p *PasswordAuth) Handshake(c net.Conn) error {
	randomValue := make([]uint8, 16)
	if err := binary.Read(c, binary.BigEndian, &randomValue); err != nil {
		return err
	}

	crypted, err := p.encrypt(p.Password, randomValue)

	if (err != nil) {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, &crypted); err != nil {
		return err
	}

	return nil
}

func (p *PasswordAuth) reverseBits(b byte) byte {
	var reverse = [256]int{
		0, 128, 64, 192, 32, 160, 96, 224,
		16, 144, 80, 208, 48, 176, 112, 240,
		8, 136, 72, 200, 40, 168, 104, 232,
		24, 152, 88, 216, 56, 184, 120, 248,
		4, 132, 68, 196, 36, 164, 100, 228,
		20, 148, 84, 212, 52, 180, 116, 244,
		12, 140, 76, 204, 44, 172, 108, 236,
		28, 156, 92, 220, 60, 188, 124, 252,
		2, 130, 66, 194, 34, 162, 98, 226,
		18, 146, 82, 210, 50, 178, 114, 242,
		10, 138, 74, 202, 42, 170, 106, 234,
		26, 154, 90, 218, 58, 186, 122, 250,
		6, 134, 70, 198, 38, 166, 102, 230,
		22, 150, 86, 214, 54, 182, 118, 246,
		14, 142, 78, 206, 46, 174, 110, 238,
		30, 158, 94, 222, 62, 190, 126, 254,
		1, 129, 65, 193, 33, 161, 97, 225,
		17, 145, 81, 209, 49, 177, 113, 241,
		9, 137, 73, 201, 41, 169, 105, 233,
		25, 153, 89, 217, 57, 185, 121, 249,
		5, 133, 69, 197, 37, 165, 101, 229,
		21, 149, 85, 213, 53, 181, 117, 245,
		13, 141, 77, 205, 45, 173, 109, 237,
		29, 157, 93, 221, 61, 189, 125, 253,
		3, 131, 67, 195, 35, 163, 99, 227,
		19, 147, 83, 211, 51, 179, 115, 243,
		11, 139, 75, 203, 43, 171, 107, 235,
		27, 155, 91, 219, 59, 187, 123, 251,
		7, 135, 71, 199, 39, 167, 103, 231,
		23, 151, 87, 215, 55, 183, 119, 247,
		15, 143, 79, 207, 47, 175, 111, 239,
		31, 159, 95, 223, 63, 191, 127, 255,
	}

	return byte(reverse[int(b)])
}

func (p *PasswordAuth) encrypt(key string, bytes []byte) ([]byte, error) {
	keyBytes := []byte{0,0,0,0,0,0,0,0}

	if len(key) > 8 {
		key = key[:8]
	}

	for i := 0; i < len(key); i++ {
		keyBytes[i] = p.reverseBits(key[i])
	}

	block, err := des.NewCipher(keyBytes)

	if err != nil {
		return nil, err
	}

	result1 := make([]byte, 8)
	block.Encrypt(result1, bytes)
	result2 := make([]byte, 8)
	block.Encrypt(result2, bytes[8:])

	crypted := append(result1, result2...)

	return crypted, nil
}
