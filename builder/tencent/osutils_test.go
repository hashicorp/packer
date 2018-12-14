package tencent

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_pathExists(t *testing.T) {
	dir := os.TempDir()
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"path exists", args{dir}, true},
		{"path doesn't exist", args{dir + "1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathExists(tt.args.path); got != tt.want {
				t.Errorf("pathExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryExists(t *testing.T) {
	dir := os.TempDir()
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty directory exists", args{""}, true},
		{"temp dir exists", args{dir}, true},
		{"dir doesn't exist", args{dir + "1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirectoryExists(tt.args.path); got != tt.want {
				t.Errorf("DirectoryExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	var tempfilename string
	defer func() {
		os.Remove(tempfilename)
	}()
	tempfilename = TempFileName()
	file, err := os.Create(tempfilename)
	defer func() {
		file.Close()
	}()
	if err != nil {
		t.Errorf("Unable to create file, %s", tempfilename)
	}
	if !FileExists(tempfilename) {
		t.Fatalf("Unable to test file: %s exists!", tempfilename)
	}
}

func TestTempFileName(t *testing.T) {
	tempfilename := TempFileName()
	dir, _ := filepath.Split(tempfilename)
	if !DirectoryExists(dir) {
		t.Errorf("The specified directory doesn't exist, directory: %s!", dir)
	}
	if FileExists(tempfilename) {
		t.Fatalf("TempFileName file already exists, filename: %s", tempfilename)
	}
}
