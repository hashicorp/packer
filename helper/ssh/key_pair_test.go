package ssh

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/hashicorp/packer/common/uuid"
	gossh "golang.org/x/crypto/ssh"
	"strconv"
	"testing"
)

// expected contains the data that the key pair should contain.
type expected struct {
	kind KeyPairType
	bits int
	desc string
	name string
	data []byte
}

const (
	PemRsa1024 = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDJEMFPpTBiWNDb3qEIPTSeEnIP8FZdBpG8njOrclcMoQQNhzZ+
4uz37tqtHMp36Z7LB4/+85NN6epNXO+ekyZIHswiyBcJC2sT3KuH7nG1BESOooPY
DfeCSM+CJT9GDIhy9nUXSsJjrceEyh/B5DjEtIbS0XfcRelrNTJodCmPJwIDAQAB
AoGAK66GMOV0c4lUJtBhL8cMTWM4gJn4SVGKC+5az16R5t58YOwFPN/UF7E+tOlS
W2bX5sgH0p3cXMr66j/Mlyjk4deLg7trDavulIP93MyVO2SUJ0cstQ0ZmRz2oGwx
Gow+hD75Cet7uvepdmG4DKHJe8D/I72rtP1WKuZyd0vP6WECQQDua6wWlyEdIimx
XoGWUvmywACWPnQmBnyHG7x5hxMjijQoQZu60zRxSU9I5q08BerTsvbTc+xLnDVv
mFzlcjT/AkEA1+P7lcvViZeNKoDB1Qt+VV+pkcqL5aoRwdnLA51SyFJ9tXkxeZwA
LOof3xtoRGhCld7ixi3kF5aZsafAJOZd2QJAH8rFyMFgTgU3MAqdFxF7cGV/7ojn
bgahZlbBfCcR20Rbjh6piHEPZifTZbI02XMkjBQqK6oikTaEPZxAjuv6uwJANczu
yWm+kUdfOpRTuY/fr87jJx3etyEmw7ROz1vJYXqNMUg+eBvUP10pDCR8W2/QCCE/
Sjvtd6NkMc2oKInwIQJAFZ1xJte0EaQsXaCIoZwHrQJbK1dd5l1xTAzz51voAcKH
2K23xgx4I+/eam2enjFa7wXLZFoW0xg/51xsaIjnrA==
-----END RSA PRIVATE KEY-----
`
	PemRsa2048 = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA/ZPazeRmBapF01gzHXtJGpu0S936xHY+pOrIyIk6lEE06paf
q5gh6BCuiN/60Keed5Nz+Es4dPGc73mql9pd7N0HOoEc1IQjZzJVqWOy3E55oWbz
rXr1qbmMjw8bGHalZsVBov1UhyB6f2bKi88fGkThJi9HZ+Dc3Jr87eW+whS4D0bI
JJe5dkY0VhDqB0YVEk299TxlAiDkeXD1EcMZrD/yHsusapwlXL2WHWmCgbPpbeYW
YJhD1bScChYmf41iiInBwFymG7kz4bPsup7wCBXpcLJplY1iuXdtVVujNLDbJwlb
Xi2oBm3WizPjYcUthvMlqOieuy6Z4KzyJd7EnQIDAQABAoIBAByZ8LQIbvl0myub
ZyiMH1LA/TURdJd0PtybHsp/r/vI3w8WrivMnQZv2z/VA5VFUrpqB0qaMWP/XJQm
RPebybxNseMHbRkLTnL1WnQgqqvurglmc1W96LecFh6MtaGswDs3RI/9wur63tY/
4dijI/7yhfKoooU097RqRt0ObNW3BxGwNKUraMLKEZjtohv1cZBeRqzGZuui351E
YsG1jt23/3OP3Acfd1xpzoi+daadxl9JTr02kE7lMjfq32quhTdzuNZP84sQsaV+
RXLNEoiSufjzy3nHTEpG6QaEWQc4gszCIBVRabxr7LtIOqJn2KmXxtOyFE52AJJj
ls3ifAECgYEA/9K+5oHdZBizWWUvNzQWXrdHUtVavCwjj+0v+yRFZSoAhKtVmLYl
8R4NeG6NCIOoJsqmGVpgtCyPndR4PQ6yr4Jt1FJorjsNw21eYrjOVG+y9Z0DkCwJ
uCRVUeqB42jLu7v9r1V3OBQdKLN6VxO4np05KEZyv1LOGGt0XC8NCykCgYEA/cC2
NR7Y4Z5OjCc2GHMQNrVZ2HTDDir71RjcIIEmsIQ5/AMAELFLCSqzD73tJ87T5jPi
aWeOpIcnK78jMvIIsbV0BXaDsjtlvCdQui2AoX63FuK4q4E+vwe5Q/TqY2nDh2io
mGHfeXECyUx4gxIede2XEO9zYQ0lP8gxnjmLkFUCgYBO8LolqQcm/xRAzp9eOn14
prekkN+Z10j1/avjpFKhn+9fAPu9zt8wYySm9/4fFXlK1xegFSpoDqQWgNzFgoaS
7/1yGifhM6nQlywb7IkGtx0S+2uBDoXFQ7jsOR/xi4HqoVzrwMS0EkjZKWDkA9rh
XwSnL+3yqduc33OdiotM2QKBgCgNCrVHsSOrQOqOJdOmFaEM7qljhIXv8t+nlNbs
i5bAyAYm0xPPZ/CCdNC/QXdPBdMHzWylk7YUPvKAsKWR3h1ubmmOUysGhQA1lGBO
XkcfIPbTwiIPvD+akHtRZM1cHCh7NGEY0ZTxaWcsUrkdWwFyBq39nVBsKrzudCZt
HsIhAoGBAMv3erZQawzIgX9nOUHB2UJl8pS+LZSNMlYvFoMHKI2hLq0g/VxUlnsh
Jzw9+fTLMVFdY+F3ydO6qQFd8wlfov7deyscdoSj8R5gjGKJsarBs+YVdFde2oLG
gkFsXmbmc2boyqGg51CbAX34VJOhGQKhWgKCWqDGmoYXafmyiZc+
-----END RSA PRIVATE KEY-----
`
	PemOpenSshRsa1024 = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAlwAAAAdzc2gtcn
NhAAAAAwEAAQAAAIEAzzknaHV741775aJOPacDpd2SiDpIDYmm7/w2sgY8lrinSakfLIVk
1qn0IBRLNOzMxoF/pvIgGQXS51xvE1vB3QK8L+8vJwH06DuOXPP1WgVoDTU03gGvBJ7MNF
5HcQYvBiIaU5XxG8l0OZO88B9RFhPP9r0XrYxAlSjuk9KKlEcAAAIYLQ46zy0OOs8AAAAH
c3NoLXJzYQAAAIEAzzknaHV741775aJOPacDpd2SiDpIDYmm7/w2sgY8lrinSakfLIVk1q
n0IBRLNOzMxoF/pvIgGQXS51xvE1vB3QK8L+8vJwH06DuOXPP1WgVoDTU03gGvBJ7MNF5H
cQYvBiIaU5XxG8l0OZO88B9RFhPP9r0XrYxAlSjuk9KKlEcAAAADAQABAAAAgQDJ9Jq6jF
08P/LhXug/38iHW0UW7S4Ru4jttHGd2MQt5DJtcJzIKA0ZxLL+nKibIPmFsOm2y5yKpolg
IE7EoBVzTeg0LedbRayc0Kc5tY7PEz0Shi9ABIMYbNo2L2pNmsq9ns0xA8ur3OugfKHsH8
XjJ1rdHsyLjoMx2ADfLY0xkQAAAEAvyrgW4jswENdErbF0rOdP+Y73B/8rxBaY/QBE2qtG
oUp7bpOtUAH2Ip7RjXOX4xTAt4n2QeHBSfX7gfXRjmY6AAAAQQDmYlgSWtTYLV9VZSScLU
OG+GkhQxYqkKN/N9LSpTP4Pwh81KpMp40yvIlufmKLgGihWVxUDzRap3aoR7PqIvHPAAAA
QQDmQ47VwclxiVn5tVAht/Lk2ZVa7rSjeFlXAkAWZkUAiHboaH8IfW9W4gYV7o2BqJO11L
0vi+vCq+le45F416wJAAAAImNocmlzQHBvZXRhc3Rlci5jb3JwLm11dHVhbGluay5uZXQ=

-----END OPENSSH PRIVATE KEY-----
`
	PemOpenSshRsa2048 = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAxWfWNu0i8sbmwPqTUfKSeXOSt/fLMuqucn9KYU7rJ+83trznRhAn
AHQzKgcSU8PBgkax+PDEUexYUB9gZApNI6K/2twVDYh3Hgwx7EjXf05rji7bQk6TFyKEp4
n348CWAdK8iFmNUutSpJLy7GciyMPLu3BK+EsXpsnuPpIm184hEFOiADZyHTGeUgvsKOAc
G7u5hBS3kty8LRZmL+pihbktFwGC4D5bapCcTaF2++zkUy4JKcVE5/2JfK1Ya6D0ATczjz
1b6+r7j2RUg1mXfK6AwMHEcamzhgeuM9RdrPtMdhZI09LCJzjmXc9pzlGu1HCZzh3rJ3hd
8PVmlAd3VQAAA+A9hesQPYXrEAAAAAdzc2gtcnNhAAABAQDFZ9Y27SLyxubA+pNR8pJ5c5
K398sy6q5yf0phTusn7ze2vOdGECcAdDMqBxJTw8GCRrH48MRR7FhQH2BkCk0jor/a3BUN
iHceDDHsSNd/TmuOLttCTpMXIoSniffjwJYB0ryIWY1S61KkkvLsZyLIw8u7cEr4Sxemye
4+kibXziEQU6IANnIdMZ5SC+wo4Bwbu7mEFLeS3LwtFmYv6mKFuS0XAYLgPltqkJxNoXb7
7ORTLgkpxUTn/Yl8rVhroPQBNzOPPVvr6vuPZFSDWZd8roDAwcRxqbOGB64z1F2s+0x2Fk
jT0sInOOZdz2nOUa7UcJnOHesneF3w9WaUB3dVAAAAAwEAAQAAAQEAvA8Z8iWjX6nA9yM/
6ZevluhVY9E60XzlR8qgL2ehet/YMcxwfzywCyyn+WfXO9mHpfZ3YfLs9Ca2U04w4900c7
h+EaAMpmHVKNjxTmpucadhq4hT9S0pz6ZgvcMgVuaHgaEjXroBencYuhQMPM5cQurUUfK+
WSAgnhJNV2qgeoEGgfDZoL1HkItckEZwIzmx4lfMVAuaeqVq5tJNcdv5ukNHpnIYl6fgDp
WGUn/9F8sSHO7P7kGl67IZIsAz+1wW+6pFaVgxbZJ3baPiURtRp+nRSaKLYZSMph6MAiTu
YC8EEVqi3X4m/ZHy+BkphfzR24ouwpt1Vv9QOAPzXXsPwQAAAIEAvmA+yiBdzsJplCifTA
KljE+KpSuvLPRPTb7MGsBO0LIvxrkXOMCXZF4I2VP1zSUH+SDPPc6JeR1Q8liMqPC3Md6c
CIkHfVFBAZL709d0ZtTiir1BipG/l5vIpBnepNX/bWIszIOMzPF2at0WF1lFe6THWujuE8
Xjp2AJSFZlUjAAAACBAOMxr6FN38VwRC1nrDcZyo4TjWVhAdk4p3AkdNJhFSiKS/x5/yo2
K1majzcKbrR8+fEPTVWGszAg+AXQdsOq17q+DMenfrBckQ9ZHr3upSZAaGN+keNwge/Kaj
yOvYiKdYFXmAulQZCPQsDNp7e7Z1dTqxi5IlhUgDPzzO0vRGjNAAAAgQDeb0Ulv7fkYAav
tZ+D0LohGjlGFwTeLdwErcVnq9wGyupdeNhlTXZXxRvF+yEt4FCV9UEUnRX75CAnpk2hT2
D5uYMyixAEfSeIo59Ln27MmAy0alR3UnT7JnLEZRh4dnvFbSSMJ1rHxf8Eg6YFJmpH65fX
exrJE+p69wgRVndoqQAAACJjaHJpc0Bwb2V0YXN0ZXIuY29ycC5tdXR1YWxpbmsubmV0AQ
IDBAUGBw==
-----END OPENSSH PRIVATE KEY-----
`
	PemDsa = `-----BEGIN DSA PRIVATE KEY-----
MIIBuwIBAAKBgQDH/T+IkpbdA9nUM7O4MMRoeS0bn7iXWs63Amo2fsIyJPxDvjjF
5HZBH5Rq045TFCCWHjymwiYof+wvwUMZIUH++ABTrKzes/r5qG5jXp42pFWf6nTI
zHwttdjvNiXr+AgreXOrJKhjv6Ga3hq8MNcXMa9xFsIB83EZNMBPxbj0nwIVAJQW
1eR4Uf8/8haQb4HkTsoH+R5/AoGBAK9FV5LIZxY1TeNsD5eXoqpTqCy1WROMggSG
VZ4yN0rrKCtLd8am61m/L8VCMUWiO3IJQdq3yWBTEBbsShL/toau9beUdTl6rdB8
wcEcNgtZnhypQR58HlmgUFWC45rW37hW4AUJuMDgLxgqSVuoF1pDcHrHSi/fZwgp
7t0MKH2SAoGAJfUcLrXg5ZR8jbpZs4L/ubibUn+y35y+33aos07auMW1MesuNTcZ
Ch42nbH2wKnbjk8eDxHdHLHzzOLGgYVMpUuBeuc7G5Q94rM/Z0I8HGQ6mvIkuFyp
58Unu5yu33GsNUgGEHmriiMGezXNXGNH/72PmTXuyxEMSrad23c6NZoCFAtIqbal
4tGCfnnmWU514A7ZzEKj
-----END DSA PRIVATE KEY-----
`
	PemEcdsa384 = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDAjuEIlmFyhGjFtJoAwD420FuPAjIknN3YwDZL4cfMFpB4YAK+7QVLs
coAJ/ADuT7OgBwYFK4EEACKhZANiAASeXKyBr2prr4f4aOsM4dtVikYOUIL3yYnb
GFOy7yHmauCnkIB48paXpvRE5m53Q8zgu7vkz/z9tcMBcC0GzpY3Sef37fmgTUuZ
AJuJp36DMBdQel+j51TcQ79sizxCayg=
-----END EC PRIVATE KEY-----
`
	PemEcdsa521 = `-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIBVCiwcf/did2vCIu3aMe7OeTD35PULm0hqmfkAK9OKIosi/DjOFfA
8h99rVNPaf+Cx/JNmEzR4bZNnYDyilSRCr+gBwYFK4EEACOhgYkDgYYABABHBMLP
XbQoRF31ZGIeUj9jt9GqKES1dLBtGDEQSiiZFouL4tEIW7NfIZDpOIkA0khNcO8N
xH6eylg0XOgcr01GRwCjY5VOapOahtn63SpajPGeKk+46F2dULIwrov9tWQuYNa3
P50N8j3rx6fAdgyDENOcCJlfNdNcySvkH4bgL1xcsw==
-----END EC PRIVATE KEY-----
`
	PemOpenSshEd25519 = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACAUftPhZQN17kAlThiiWJEgJvddm/pUhHvgrHUtpuYFOQAAAKjN+UhDzflI
QwAAAAtzc2gtZWQyNTUxOQAAACAUftPhZQN17kAlThiiWJEgJvddm/pUhHvgrHUtpuYFOQ
AAAEANXlEZdNU03RMmj77O2ojWh06Hbj8/qQ++H5wkt688NBR+0+FlA3XuQCVOGKJYkSAm
912b+lSEe+CsdS2m5gU5AAAAImNocmlzQHBvZXRhc3Rlci5jb3JwLm11dHVhbGluay5uZX
QBAgM=
-----END OPENSSH PRIVATE KEY-----
`
)

func (o expected) matches(kp KeyPair) error {
	if o.kind.String() == "" {
		return errors.New("expected kind's value cannot be empty")
	}

	if o.bits <= 0 {
		return errors.New("expected bits' value cannot be less than or equal to 0")
	}

	if o.desc == "" {
		return errors.New("expected description's value cannot be empty")
	}

	if len(o.data) == 0 {
		return errors.New("expected random data value cannot be nothing")
	}

	if kp.Type() != o.kind {
		return errors.New("key pair type should be " + o.kind.String() +
			" - got '" + kp.Type().String() + "'")
	}

	if kp.Bits() != o.bits {
		return errors.New("key pair bits should be " + strconv.Itoa(o.bits) +
			" - got " + strconv.Itoa(kp.Bits()))
	}

	if len(o.name) > 0 && kp.Name() != o.name {
		return errors.New("key pair name should be '" + o.name +
			"' - got '" + kp.Name() + "'")
	}

	if kp.Description() != o.desc {
		return errors.New("key pair description should be '" +
			o.desc + "' - got '" + kp.Description() + "'")
	}

	err := o.verifyPublicKeyAuthorizedKeysFormat(kp)
	if err != nil {
		return err
	}

	err = o.verifyKeyPair(kp)
	if err != nil {
		return err
	}

	return nil
}

func (o expected) verifyPublicKeyAuthorizedKeysFormat(kp KeyPair) error {
	newLines := []NewLineOption{
		UnixNewLine,
		NoNewLine,
		WindowsNewLine,
	}

	for _, nl := range newLines {
		publicKeyAk := kp.PublicKeyAuthorizedKeysLine(nl)

		if len(publicKeyAk) < 2 {
			return errors.New("expected public key in authorized keys format to be at least 2 bytes")
		}

		switch nl {
		case NoNewLine:
			if publicKeyAk[len(publicKeyAk)-1] == '\n' {
				return errors.New("public key in authorized keys format has trailing new line when none was specified")
			}
		case UnixNewLine:
			if publicKeyAk[len(publicKeyAk)-1] != '\n' {
				return errors.New("public key in authorized keys format does not have unix new line when unix was specified")
			}
			if string(publicKeyAk[len(publicKeyAk)-2:]) == WindowsNewLine.String() {
				return errors.New("public key in authorized keys format has windows new line when unix was specified")
			}
		case WindowsNewLine:
			if string(publicKeyAk[len(publicKeyAk)-2:]) != WindowsNewLine.String() {
				return errors.New("public key in authorized keys format does not have windows new line when windows was specified")
			}
		}

		if len(o.name) > 0 {
			if len(publicKeyAk) < len(o.name) {
				return errors.New("public key in authorized keys format is shorter than the key pair's name")
			}

			suffix := []byte{' '}
			suffix = append(suffix, o.name...)
			suffix = append(suffix, nl.Bytes()...)
			if !bytes.HasSuffix(publicKeyAk, suffix) {
				return errors.New("public key in authorized keys format with name does not have name in suffix - got '" +
					string(publicKeyAk) + "'")
			}
		}
	}

	return nil
}

func (o expected) verifyKeyPair(kp KeyPair) error {
	signer, err := gossh.ParsePrivateKey(kp.PrivateKeyPemBlock())
	if err != nil {
		return errors.New("failed to parse private key during verification - " + err.Error())
	}

	signature, err := signer.Sign(rand.Reader, o.data)
	if err != nil {
		return errors.New("failed to sign test data during verification - " + err.Error())
	}

	err = signer.PublicKey().Verify(o.data, signature)
	if err != nil {
		return errors.New("failed to verify test data - " + err.Error())
	}

	return nil
}

func TestDefaultKeyPairBuilder_Build_Default(t *testing.T) {
	kp, err := NewKeyPairBuilder().Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_EcdsaDefault(t *testing.T) {
	kp, err := NewKeyPairBuilder().
		SetType(Ecdsa).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_EcdsaSupportedCurves(t *testing.T) {
	supportedBits := []int{
		521,
		384,
		256,
	}

	for _, bits := range supportedBits {
		kp, err := NewKeyPairBuilder().
			SetType(Ecdsa).
			SetBits(bits).
			Build()
		if err != nil {
			t.Fatal(err.Error())
		}

		err = expected{
			kind: Ecdsa,
			bits: bits,
			desc: strconv.Itoa(bits) + " bit ECDSA",
			data: []byte(uuid.TimeOrderedUUID()),
		}.matches(kp)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestDefaultKeyPairBuilder_Build_RsaDefault(t *testing.T) {
	kp, err := NewKeyPairBuilder().
		SetType(Rsa).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Rsa,
		bits: 4096,
		desc: "4096 bit RSA",
		data: []byte(uuid.TimeOrderedUUID()),
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_NamedEcdsa(t *testing.T) {
	name := uuid.TimeOrderedUUID()

	kp, err := NewKeyPairBuilder().
		SetType(Ecdsa).
		SetName(name).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Ecdsa,
		bits: 521,
		desc: "521 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
		name: name,
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_Build_NamedRsa(t *testing.T) {
	name := uuid.TimeOrderedUUID()

	kp, err := NewKeyPairBuilder().
		SetType(Rsa).
		SetName(name).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	err = expected{
		kind: Rsa,
		bits: 4096,
		desc: "4096 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
		name: name,
	}.matches(kp)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDefaultKeyPairBuilder_SetPrivateKey(t *testing.T) {
	name := uuid.TimeOrderedUUID()
	pemData := make(map[string]expected)
	pemData[PemRsa1024] = expected{
		bits: 1024,
		kind: Rsa,
		name: name,
		desc: "1024 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemRsa2048] = expected{
		bits: 2048,
		kind: Rsa,
		name: name,
		desc: "2048 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshRsa1024] = expected{
		bits: 1024,
		kind: Rsa,
		name: name,
		desc: "1024 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshRsa2048] = expected{
		bits: 2048,
		kind: Rsa,
		name: name,
		desc: "2048 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemDsa] = expected{
		bits: 1024,
		kind: Dsa,
		name: name,
		desc: "1024 bit DSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemEcdsa384] = expected{
		bits: 384,
		kind: Ecdsa,
		name: name,
		desc: "384 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemEcdsa521] = expected{
		bits: 521,
		kind: Ecdsa,
		name: name,
		desc: "521 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshEd25519] = expected{
		bits: 256,
		kind: Ed25519,
		name: name,
		desc: "256 bit ED25519 named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}

	for s, l := range pemData {
		kp, err := NewKeyPairBuilder().SetPrivateKey([]byte(s)).SetName(name).Build()
		if err != nil {
			t.Fatal(err)
		}
		err = l.matches(kp)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestDefaultKeyPairBuilder_SetPrivateKey_Override(t *testing.T) {
	name := uuid.TimeOrderedUUID()
	pemData := make(map[string]expected)
	pemData[PemRsa1024] = expected{
		bits: 1024,
		kind: Rsa,
		name: name,
		desc: "1024 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemRsa2048] = expected{
		bits: 2048,
		kind: Rsa,
		name: name,
		desc: "2048 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshRsa1024] = expected{
		bits: 1024,
		kind: Rsa,
		name: name,
		desc: "1024 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshRsa2048] = expected{
		bits: 2048,
		kind: Rsa,
		name: name,
		desc: "2048 bit RSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemDsa] = expected{
		bits: 1024,
		kind: Dsa,
		name: name,
		desc: "1024 bit DSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemEcdsa384] = expected{
		bits: 384,
		kind: Ecdsa,
		name: name,
		desc: "384 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemEcdsa521] = expected{
		bits: 521,
		kind: Ecdsa,
		name: name,
		desc: "521 bit ECDSA named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}
	pemData[PemOpenSshEd25519] = expected{
		bits: 256,
		kind: Ed25519,
		name: name,
		desc: "256 bit ED25519 named " + name,
		data: []byte(uuid.TimeOrderedUUID()),
	}

	supportedKeyTypes := []KeyPairType{Rsa, Dsa}
	for _, kt := range supportedKeyTypes {
		for s, l := range pemData {
			kp, err := NewKeyPairBuilder().
				SetPrivateKey([]byte(s)).
				SetName(name).
				SetType(kt).
				Build()
			if err != nil {
				t.Fatal(err)
			}
			err = l.matches(kp)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
