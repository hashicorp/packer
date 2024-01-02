// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package file

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"destination": "something",
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a provisioner")
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_InvalidSource(t *testing.T) {
	var p Provisioner
	config := testConfig()
	config["source"] = "/this/should/not/exist"

	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should require existing file")
	}

	config["generated"] = false
	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should required existing file")
	}
}

func TestProvisionerPrepare_ValidSource(t *testing.T) {
	var p Provisioner

	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config := testConfig()
	config["source"] = tf.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should allow valid file: %s", err)
	}

	config["generated"] = false
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should allow valid file: %s", err)
	}
}

func TestProvisionerPrepare_GeneratedSource(t *testing.T) {
	var p Provisioner

	config := testConfig()
	config["source"] = "/this/should/not/exist"
	config["generated"] = true
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should allow non-existing file: %s", err)
	}
}

func TestProvisionerPrepare_EmptyDestination(t *testing.T) {
	var p Provisioner

	config := testConfig()
	delete(config, "destination")
	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should require destination path")
	}
}

func TestProvisionerProvision_SendsFile(t *testing.T) {
	var p Provisioner
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	if _, err = tf.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"source":      tf.Name(),
		"destination": "something",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), tf.Name()) {
		t.Fatalf("should print source filename")
	}

	if !strings.Contains(b.String(), "something") {
		t.Fatalf("should print destination filename")
	}

	if comm.UploadPath != "something" {
		t.Fatalf("should upload to configured destination")
	}

	if comm.UploadData != "hello" {
		t.Fatalf("should upload with source file's data")
	}
}

func TestProvisionerProvision_SendsContent(t *testing.T) {
	var p Provisioner

	dst := "something.txt"
	content := "hello"
	config := map[string]interface{}{
		"content":     content,
		"destination": dst,
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), "something") {
		t.Fatalf("should print destination filename")
	}

	if comm.UploadPath != dst {
		t.Fatalf("should upload to configured destination")
	}

	if comm.UploadData != content {
		t.Fatalf("should upload with source file's data")
	}

}

func TestProvisionerProvision_SendsFileMultipleFiles(t *testing.T) {
	var p Provisioner
	tf1, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf1.Name())

	if _, err = tf1.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	tf2, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf2.Name())

	if _, err = tf2.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"sources":     []string{tf1.Name(), tf2.Name()},
		"destination": "something",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), tf1.Name()) {
		t.Fatalf("should print first source filename")
	}

	if !strings.Contains(b.String(), tf2.Name()) {
		t.Fatalf("should print second source filename")
	}
}

func TestProvisionerProvision_SendsFileMultipleDirs(t *testing.T) {
	var p Provisioner

	// Prepare the first directory
	td1, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatalf("error temp folder 1: %s", err)
	}
	defer os.Remove(td1)

	tf1, err := os.CreateTemp(td1, "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}

	if _, err = tf1.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	// Prepare the second directory
	td2, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatalf("error temp folder 1: %s", err)
	}
	defer os.Remove(td2)

	tf2, err := os.CreateTemp(td2, "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}

	if _, err = tf2.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	if _, err = tf1.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	// Run Provision

	config := map[string]interface{}{
		"sources":     []string{td1, td2},
		"destination": "something",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), td1) {
		t.Fatalf("should print first directory")
	}

	if !strings.Contains(b.String(), td2) {
		t.Fatalf("should print second directory")
	}
}

func TestProvisionerProvision_DownloadsMultipleFilesToFolder(t *testing.T) {
	var p Provisioner

	tf1, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf1.Name())

	if _, err = tf1.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	tf2, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf2.Name())

	if _, err = tf2.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"sources":     []string{tf1.Name(), tf2.Name()},
		"destination": "something/",
		"direction":   "download",
	}

	// Cleaning up destination directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed getting current working directory")
	}
	destinationDir := filepath.Join(cwd, "something")
	defer os.RemoveAll(destinationDir)

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), tf1.Name()) {
		t.Errorf("should print source filenam '%s'e; output: \n%s", tf1.Name(), b.String())
	}

	if !strings.Contains(b.String(), tf2.Name()) {
		t.Errorf("should second source filename '%s'; output: \n%s", tf2.Name(), b.String())
	}

	dst1 := filepath.Join("something", filepath.Base(tf1.Name()))
	if !strings.Contains(b.String(), dst1) {
		t.Errorf("should print destination filename '%s'; output: \n%s", dst1, b.String())
	}

	dst2 := filepath.Join("something", filepath.Base(tf2.Name()))
	if !strings.Contains(b.String(), dst2) {
		t.Errorf("should print destination filename '%s'; output: \n%s", dst2, b.String())
	}
}

func TestProvisionerProvision_SendsFileMultipleFilesToFolder(t *testing.T) {
	var p Provisioner

	tf1, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf1.Name())

	if _, err = tf1.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	tf2, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf2.Name())

	if _, err = tf2.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"sources":     []string{tf1.Name(), tf2.Name()},
		"destination": "something/",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{
		Writer: b,
		PB:     &packersdk.NoopProgressTracker{},
	}
	comm := &packersdk.MockCommunicator{}
	err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(b.String(), tf1.Name()) {
		t.Fatalf("should print first source filename")
	}

	if !strings.Contains(b.String(), tf2.Name()) {
		t.Fatalf("should print second source filename")
	}

	dstRegex := regexp.MustCompile("something/\n")
	allDst := dstRegex.FindAllString(b.String(), -1)
	if len(allDst) != 2 {
		t.Fatalf("some destinations are broken; output: \n%s", b.String())
	}
}

func TestProvisionDownloadMkdirAll(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"dir"},
		{"dir/"},
		{"dir/subdir"},
		{"dir/subdir/"},
		{"path/to/dir"},
		{"path/to/dir/"},
	}
	tmpDir, err := os.MkdirTemp("", "packer-file")
	if err != nil {
		t.Fatalf("error tempdir: %s", err)
	}
	defer os.RemoveAll(tmpDir)
	tf, err := os.CreateTemp(tmpDir, "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config := map[string]interface{}{
		"source": tf.Name(),
	}
	var p Provisioner
	for _, test := range tests {
		path := filepath.Join(tmpDir, test.path)
		config["destination"] = filepath.Join(path, "something")
		if err := p.Prepare(config); err != nil {
			t.Fatalf("err: %s", err)
		}
		b := bytes.NewBuffer(nil)
		ui := &packersdk.BasicUi{
			Writer: b,
			PB:     &packersdk.NoopProgressTracker{},
		}
		comm := &packersdk.MockCommunicator{}
		err = p.ProvisionDownload(ui, comm)
		if err != nil {
			t.Fatalf("should successfully provision: %s", err)
		}

		if !strings.Contains(b.String(), tf.Name()) {
			t.Fatalf("should print source filename")
		}

		if !strings.Contains(b.String(), "something") {
			t.Fatalf("should print destination filename")
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat of download dir should not error: %s", err)
		}

		if _, err := os.Stat(config["destination"].(string)); err != nil {
			t.Fatalf("stat of destination file should not error: %s", err)
		}
	}
}
