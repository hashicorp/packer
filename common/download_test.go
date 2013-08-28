package common

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
)

func TestDownloadClient_VerifyChecksum(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("tempfile error: %s", err)
	}
	defer os.Remove(tf.Name())

	// "foo"
	checksum, err := hex.DecodeString("acbd18db4cc2f85cedef654fccc4a4d8")
	if err != nil {
		t.Fatalf("decode err: %s", err)
	}

	// Write the file
	tf.Write([]byte("foo"))
	tf.Close()

	config := &DownloadConfig{
		Hash:     md5.New(),
		Checksum: checksum,
	}

	d := NewDownloadClient(config)
	result, err := d.VerifyChecksum(tf.Name())
	if err != nil {
		t.Fatalf("Verify err: %s", err)
	}

	if !result {
		t.Fatal("didn't verify")
	}
}

func TestHashForType(t *testing.T) {
	if h := HashForType("md5"); h == nil {
		t.Fatalf("md5 hash is nil")
	} else {
		h.Write([]byte("foo"))
		result := h.Sum(nil)

		expected := "acbd18db4cc2f85cedef654fccc4a4d8"
		actual := hex.EncodeToString(result)
		if actual != expected {
			t.Fatalf("bad hash: %s", actual)
		}
	}

	if h := HashForType("sha1"); h == nil {
		t.Fatalf("sha1 hash is nil")
	} else {
		h.Write([]byte("foo"))
		result := h.Sum(nil)

		expected := "0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"
		actual := hex.EncodeToString(result)
		if actual != expected {
			t.Fatalf("bad hash: %s", actual)
		}
	}

	if h := HashForType("sha256"); h == nil {
		t.Fatalf("sha256 hash is nil")
	} else {
		h.Write([]byte("foo"))
		result := h.Sum(nil)

		expected := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
		actual := hex.EncodeToString(result)
		if actual != expected {
			t.Fatalf("bad hash: %s", actual)
		}
	}

	if h := HashForType("sha512"); h == nil {
		t.Fatalf("sha512 hash is nil")
	} else {
		h.Write([]byte("foo"))
		result := h.Sum(nil)

		expected := "f7fbba6e0636f890e56fbbf3283e524c6fa3204ae298382d624741d0dc6638326e282c41be5e4254d8820772c5518a2c5a8c0c7f7eda19594a7eb539453e1ed7"
		actual := hex.EncodeToString(result)
		if actual != expected {
			t.Fatalf("bad hash: %s", actual)
		}
	}

	if HashForType("fake") != nil {
		t.Fatalf("fake hash is not nil")
	}
}
