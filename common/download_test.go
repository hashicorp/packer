package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	getter "github.com/hashicorp/go-getter"
)

func TestDownloadClientVerifyChecksum(t *testing.T) {
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
		HashType: "md5",
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

func TestDownloadClient_basic(t *testing.T) {
	tf, _ := ioutil.TempFile("", "packer")
	tf.Close()
	os.Remove(tf.Name())

	ts := httptest.NewServer(http.FileServer(http.Dir("./test-fixtures/root")))
	defer ts.Close()

	client := NewDownloadClient(&DownloadConfig{
		Url:        ts.URL + "/basic.txt",
		TargetPath: tf.Name(),
	})

	path, err := client.Get()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if string(raw) != "hello\n" {
		t.Fatalf("bad: %s", string(raw))
	}
}

func TestDownloadClient_checksumBad(t *testing.T) {
	checksum, err := hex.DecodeString("b2946ac92492d2347c6235b4d2611184")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tf, _ := ioutil.TempFile("", "packer")
	tf.Close()
	os.Remove(tf.Name())

	ts := httptest.NewServer(http.FileServer(http.Dir("./test-fixtures/root")))
	defer ts.Close()

	client := NewDownloadClient(&DownloadConfig{
		Url:        ts.URL + "/basic.txt",
		TargetPath: tf.Name(),
		Hash:       getter.HashForType("md5"),
		HashType:   "md5",
		Checksum:   checksum,
	})
	if _, err := client.Get(); err == nil {
		t.Fatal("should error")
	}
}

func TestDownloadClient_checksumGood(t *testing.T) {
	checksum, err := hex.DecodeString("b1946ac92492d2347c6235b4d2611184")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tf, _ := ioutil.TempFile("", "packer")
	tf.Close()
	os.Remove(tf.Name())

	ts := httptest.NewServer(http.FileServer(http.Dir("./test-fixtures/root")))
	defer ts.Close()

	client := NewDownloadClient(&DownloadConfig{
		Url:        ts.URL + "/basic.txt",
		TargetPath: tf.Name(),
		Hash:       getter.HashForType("md5"),
		HashType:   "md5",
		Checksum:   checksum,
	})
	path, err := client.Get()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if string(raw) != "hello\n" {
		t.Fatalf("bad: %s", string(raw))
	}
}

func TestDownloadClient_checksumNoDownload(t *testing.T) {
	checksum, err := hex.DecodeString("3740570a423feec44c2a759225a9fcf9")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	ts := httptest.NewServer(http.FileServer(http.Dir("./test-fixtures/root")))
	defer ts.Close()

	client := NewDownloadClient(&DownloadConfig{
		Url:        ts.URL + "/basic.txt",
		TargetPath: "./test-fixtures/root/another.txt",
		Hash:       getter.HashForType("md5"),
		HashType:   "md5",
		Checksum:   checksum,
	})
	path, err := client.Get()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// If this says "hello" it means we downloaded it. We faked out
	// the downloader above by giving it the checksum for "another", but
	// requested the download of "hello"
	if string(raw) != "another\n" {
		t.Fatalf("bad: %s", string(raw))
	}
}

func TestDownloadClient_resume(t *testing.T) {
	tf, _ := ioutil.TempFile("", "packer")
	tf.Write([]byte("w"))
	tf.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			rw.Header().Set("Accept-Ranges", "bytes")
			rw.WriteHeader(204)
			return
		}

		http.ServeFile(rw, r, "./test-fixtures/root/basic.txt")
	}))
	defer ts.Close()

	client := NewDownloadClient(&DownloadConfig{
		Url:        ts.URL,
		TargetPath: tf.Name(),
	})
	path, err := client.Get()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if string(raw) != "wello\n" {
		t.Fatalf("bad: %s", string(raw))
	}
}

func TestDownloadClient_usesDefaultUserAgent(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("tempfile error: %s", err)
	}
	defer os.Remove(tf.Name())

	defaultUserAgent := ""
	asserted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if defaultUserAgent == "" {
			defaultUserAgent = r.UserAgent()
		} else {
			incomingUserAgent := r.UserAgent()
			if incomingUserAgent != defaultUserAgent {
				t.Fatalf("Expected user agent %s, got: %s", defaultUserAgent, incomingUserAgent)
			}

			asserted = true
		}
	}))

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	_, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	config := &DownloadConfig{
		Url:        server.URL,
		TargetPath: tf.Name(),
	}

	client := NewDownloadClient(config)
	_, err = client.Get()
	if err != nil {
		t.Fatal(err)
	}

	if !asserted {
		t.Fatal("User-Agent never observed")
	}
}

func TestDownloadClient_setsUserAgent(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("tempfile error: %s", err)
	}
	defer os.Remove(tf.Name())

	asserted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asserted = true
		if r.UserAgent() != "fancy user agent" {
			t.Fatalf("Expected useragent fancy user agent, got: %s", r.UserAgent())
		}
	}))
	config := &DownloadConfig{
		Url:        server.URL,
		TargetPath: tf.Name(),
		UserAgent:  "fancy user agent",
	}

	client := NewDownloadClient(config)
	_, err = client.Get()
	if err != nil {
		t.Fatal(err)
	}

	if !asserted {
		t.Fatal("HTTP request never made")
	}
}

// TestDownloadFileUrl tests a special case where we use a local file for
// iso_url. In this case we can still verify the checksum but we should not
// delete the file if the checksum fails. Instead we'll just error and let the
// user fix the checksum.
func TestDownloadFileUrl(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to detect working directory: %s", err)
	}

	// source_path is a file path and source is a network path
	sourcePath := fmt.Sprintf("%s/test-fixtures/fileurl/%s", cwd, "cake")

	filePrefix := "file://"
	if runtime.GOOS == "windows" {
		filePrefix += "/"
	}

	source := fmt.Sprintf(filePrefix + sourcePath)
	t.Logf("Trying to download %s", source)

	config := &DownloadConfig{
		Url: source,
		// This should be wrong. We want to make sure we don't delete
		Checksum: []byte("nope"),
		Hash:     getter.HashForType("sha256"),
		HashType: "sha256",
		CopyFile: false,
	}

	client := NewDownloadClient(config)

	// Verify that we fail to match the checksum
	_, err = client.Get()
	if err.Error() != "checksums didn't match expected: 6e6f7065" {
		t.Fatalf("Unexpected failure; expected checksum not to match")
	}

	if _, err = os.Stat(sourcePath); err != nil {
		t.Errorf("Could not stat source file: %s", sourcePath)
	}

}
