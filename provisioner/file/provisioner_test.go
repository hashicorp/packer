// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package file

import (
	"bytes"
	"context"
	"os"
	"path"
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

func TestPassesFilters(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		exclude []string
		rel     string
		want    bool
	}{
		{"no filters passes", nil, nil, "a/b.txt", true},
		{"include matches rel path", []string{"config/*.yaml"}, nil, "config/app.yaml", true},
		{"include matches basename", []string{"*.log"}, nil, "var/log/app.log", true},
		{"include misses", []string{"*.log"}, nil, "app.txt", false},
		{"exclude wins over include", []string{"*.log"}, []string{"app.log"}, "app.log", false},
		{"exclude matches rel path", nil, []string{"tmp/*"}, "tmp/x", false},
		{"exclude basename", nil, []string{"*.tmp"}, "a/b.tmp", false},
		{"exclude misses", nil, []string{"*.tmp"}, "a/b.txt", true},
		{"include and exclude different files", []string{"src/**"}, []string{"*.bak"}, "src/x.bak", false},
		{"include subdir via basename", []string{"*.yaml"}, nil, "deep/nested/conf.yaml", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := passesFilters(tt.include, tt.exclude, tt.rel); got != tt.want {
				t.Errorf("passesFilters(%v, %v, %q) = %v, want %v", tt.include, tt.exclude, tt.rel, got, tt.want)
			}
		})
	}
}

func TestProvisionerPrepare_IncludeExcludeWithContentFails(t *testing.T) {
	var p Provisioner
	config := map[string]interface{}{
		"content":     "hello",
		"destination": "something",
		"include":     []string{"*.txt"},
	}
	if err := p.Prepare(config); err == nil {
		t.Fatal("expected error when include is used with content")
	}

	var p2 Provisioner
	config2 := map[string]interface{}{
		"content":     "hello",
		"destination": "something",
		"exclude":     []string{"*.log"},
	}
	if err := p2.Prepare(config2); err == nil {
		t.Fatal("expected error when exclude is used with content")
	}
}

func TestProvisionerProvision_UploadDirWithExclude(t *testing.T) {
	var p Provisioner

	td, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatalf("error temp folder: %s", err)
	}
	defer os.Remove(td)

	// tree: keep.txt, skip.log
	files := map[string]string{"keep.txt": "a", "skip.log": "b"}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(td, rel), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := map[string]interface{}{
		"sources":     []string{td},
		"destination": "something/",
		"exclude":     []string{"*.log", "tmp/*"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{Writer: b, PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	if err := p.Provision(context.Background(), ui, comm, make(map[string]interface{})); err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	// Exclude on a directory upload is enforced locally: a filtered tree is
	// staged and handed to UploadDir with a trailing slash and NO
	// communicator-side exclude (WinRM's UploadDir ignores it in SDK v0.6.13).
	if comm.UploadDirSrc == "" || comm.UploadDirSrc == td {
		t.Fatalf("expected UploadDir src to be a staged temp dir, got %q", comm.UploadDirSrc)
	}
	if !strings.HasSuffix(comm.UploadDirSrc, "/") {
		t.Fatalf("expected staged src to carry a trailing slash, got %q", comm.UploadDirSrc)
	}
	if len(comm.UploadDirExclude) != 0 {
		t.Fatalf("expected no exclude forwarded to UploadDir (enforced locally), got %v", comm.UploadDirExclude)
	}
	// Source had no trailing slash, so the basename becomes a subdirectory of
	// the remote destination, joined with slash semantics.
	wantDst := path.Join("something/", filepath.Base(td))
	if comm.UploadDirDst != wantDst {
		t.Fatalf("expected UploadDir dst %q, got %q", wantDst, comm.UploadDirDst)
	}
	out := b.String()
	if !strings.Contains(out, "keep.txt") {
		t.Fatalf("expected output to mention kept file, got:\n%s", out)
	}
	if strings.Contains(out, "skip.log") {
		t.Fatalf("excluded file must not be staged/uploaded, got:\n%s", out)
	}
}

func TestProvisionerProvision_UploadDirWithIncludeStagesFilteredTree(t *testing.T) {
	var p Provisioner

	td, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatalf("error temp folder: %s", err)
	}
	defer os.Remove(td)

	// tree: keep.txt, skip.log, sub/keep2.txt, sub/skip2.log
	files := map[string]string{
		"keep.txt":      "a",
		"skip.log":      "b",
		"sub/keep2.txt": "c",
		"sub/skip2.log": "d",
	}
	for rel, content := range files {
		full := filepath.Join(td, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := map[string]interface{}{
		"sources":     []string{td},
		"destination": "something/",
		"include":     []string{"*.txt"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{Writer: b, PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	if err := p.Provision(context.Background(), ui, comm, make(map[string]interface{})); err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	// include set => filtered tree staged and uploaded via UploadDir (which
	// creates remote parent dirs), not per-file Upload calls.
	if comm.UploadDirSrc == "" || comm.UploadDirSrc == td {
		t.Fatalf("expected UploadDir src to be a staged temp dir, got %q", comm.UploadDirSrc)
	}
	if comm.UploadCalled {
		t.Fatal("expected no individual Upload calls when a directory is filtered")
	}
	wantDst := path.Join("something/", filepath.Base(td))
	if comm.UploadDirDst != wantDst {
		t.Fatalf("expected UploadDir dst %q, got %q", wantDst, comm.UploadDirDst)
	}
	out := b.String()
	if !strings.Contains(out, "keep.txt") || !strings.Contains(out, "keep2.txt") {
		t.Fatalf("expected output to mention included files, got:\n%s", out)
	}
	if strings.Contains(out, "skip.log") || strings.Contains(out, "skip2.log") {
		t.Fatalf("excluded-by-include files must not be staged, got:\n%s", out)
	}
}

func TestProvisionerProvision_UploadDirTrailingSlashKeepsDst(t *testing.T) {
	var p Provisioner

	td, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(td)
	if err := os.WriteFile(filepath.Join(td, "keep.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}

	config := map[string]interface{}{
		"sources":     []string{td + "/"}, // trailing slash: contents go straight into dst
		"destination": "/srv/app/",
		"include":     []string{"*.txt"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := &packersdk.BasicUi{Writer: bytes.NewBuffer(nil), PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	if err := p.Provision(context.Background(), ui, comm, make(map[string]interface{})); err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}
	if comm.UploadDirDst != "/srv/app/" {
		t.Fatalf("expected dst to stay %q with a trailing-slash source, got %q", "/srv/app/", comm.UploadDirDst)
	}
	if !strings.HasSuffix(comm.UploadDirSrc, "/") {
		t.Fatalf("expected staged src to carry a trailing slash, got %q", comm.UploadDirSrc)
	}
}

func TestProvisionerProvision_DownloadDirWithIncludeFails(t *testing.T) {
	var p Provisioner

	config := map[string]interface{}{
		"sources":     []string{"/var/log/"},
		"destination": t.TempDir() + "/",
		"direction":   "download",
		"include":     []string{"*.gz"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := &packersdk.BasicUi{Writer: bytes.NewBuffer(nil), PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err == nil {
		t.Fatal("expected an error: include is not supported for directory downloads")
	}
	if !strings.Contains(err.Error(), "include is not supported for directory or glob downloads") {
		t.Fatalf("unexpected error: %s", err)
	}
	if comm.DownloadDirSrc != "" {
		t.Fatalf("DownloadDir must not be called when include is set, got src %q", comm.DownloadDirSrc)
	}
}

func TestProvisionerProvision_SkipsExcludedFileSource(t *testing.T) {
	var p Provisioner

	td, err := os.MkdirTemp("", "packerdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(td)
	f1 := filepath.Join(td, "a.txt")
	f2 := filepath.Join(td, "b.log")
	for _, f := range []string{f1, f2} {
		if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := map[string]interface{}{
		"sources":     []string{f1, f2},
		"destination": "dest/",
		"exclude":     []string{"*.log"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	b := bytes.NewBuffer(nil)
	ui := &packersdk.BasicUi{Writer: b, PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	if err := p.Provision(context.Background(), ui, comm, make(map[string]interface{})); err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}
	if !comm.UploadCalled {
		t.Fatal("expected the .txt file to be uploaded")
	}
	if strings.Contains(b.String(), "Skipping "+f2) == false {
		t.Fatalf("expected skip message for excluded file, got:\n%s", b.String())
	}
}

func TestProvisionerProvision_DownloadDirWithExclude(t *testing.T) {
	var p Provisioner

	config := map[string]interface{}{
		"sources":     []string{"/var/log/"},
		"destination": t.TempDir() + "/",
		"direction":   "download",
		"exclude":     []string{"*.gz"},
	}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := &packersdk.BasicUi{Writer: bytes.NewBuffer(nil), PB: &packersdk.NoopProgressTracker{}}
	comm := &packersdk.MockCommunicator{}
	if err := p.Provision(context.Background(), ui, comm, make(map[string]interface{})); err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}
	if comm.DownloadDirSrc != "/var/log/" {
		t.Fatalf("expected DownloadDir src /var/log/, got %q", comm.DownloadDirSrc)
	}
	if len(comm.DownloadDirExclude) != 1 || comm.DownloadDirExclude[0] != "*.gz" {
		t.Fatalf("expected exclude forwarded to DownloadDir, got %v", comm.DownloadDirExclude)
	}
}
