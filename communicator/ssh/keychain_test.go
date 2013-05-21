package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"testing"
)

const testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBALdGZxkXDAjsYk10ihwU6Id2KeILz1TAJuoq4tOgDWxEEGeTrcld
r/ZwVaFzjWzxaf6zQIJbfaSEAhqD5yo72+sCAwEAAQJBAK8PEVU23Wj8mV0QjwcJ
tZ4GcTUYQL7cF4+ezTCE9a1NrGnCP2RuQkHEKxuTVrxXt+6OF15/1/fuXnxKjmJC
nxkCIQDaXvPPBi0c7vAxGwNY9726x01/dNbHCE0CBtcotobxpwIhANbbQbh3JHVW
2haQh4fAG5mhesZKAGcxTyv4mQ7uMSQdAiAj+4dzMpJWdSzQ+qGHlHMIBvVHLkqB
y2VdEyF7DPCZewIhAI7GOI/6LDIFOvtPo6Bj2nNmyQ1HU6k/LRtNIXi4c9NJAiAr
rrxx26itVhJmcvoUhOjwuzSlP2bE5VHAvkGB352YBg==
-----END RSA PRIVATE KEY-----`

func TestAddPEMKey(t *testing.T) {
	k := &SimpleKeychain{}
	err := k.AddPEMKey(testPrivateKey)
	if err != nil {
		t.Fatalf("error while adding key: %s", err)
	}
}

func TestSimpleKeyChain_ImplementsClientkeyring(t *testing.T) {
	var raw interface{}
	raw = &SimpleKeychain{}
	if _, ok := raw.(ssh.ClientKeyring); !ok {
		t.Fatal("SimpleKeychain is not a valid ssh.ClientKeyring")
	}
}
