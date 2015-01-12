package winrm

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"github.com/sneal/go-winrm"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type fileManager struct {
	comm           *comm
	guestUploadDir string
	hostUploadDir  string
}

func (f *fileManager) UploadFile(dst string, src string) error {
	winDest := winFriendlyPath(dst)
	log.Printf("Uploading: %s ->%s", src, winDest)

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	return f.Upload(winDest, srcFile)
}

func (f *fileManager) Upload(dst string, input io.Reader) error {
	guestFileName := fmt.Sprintf("winrm-upload-%s.tmp", uuid.TimeOrderedUUID())

	// Paths to the file, one for cmd.exe and the other for PowerShell
	cmdGuestFilePath := fmt.Sprintf("%%TEMP%%\\%s", guestFileName)
	psGuestFilePath := fmt.Sprintf("$env:TEMP\\%s", guestFileName)

	// Upload the file in chunks to get around the Windows command line size limit
	bytes, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}

	for _, chunk := range encodeChunks(bytes, 8000-len(cmdGuestFilePath)) {
		err = f.runCommand(fmt.Sprintf("echo %s >> \"%s\"", chunk, cmdGuestFilePath))
		if err != nil {
			return err
		}
	}

	// Move the file to its permanent location and decode
	err = f.runCommand(winrm.Powershell(fmt.Sprintf(`
    $tmp_file_path = [System.IO.Path]::GetFullPath("%s")
    $dest_file_path = [System.IO.Path]::GetFullPath("%s")

    if (Test-Path $dest_file_path) {
      rm $dest_file_path
    }
    else {
      $dest_dir = ([System.IO.Path]::GetDirectoryName($dest_file_path))
      New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path $dest_dir
    }

    if (Test-Path $tmp_file_path) {
			$base64_lines = Get-Content $tmp_file_path
    	$base64_string = [string]::join("",$base64_lines)
   		$bytes = [System.Convert]::FromBase64String($base64_string) 
    	[System.IO.File]::WriteAllBytes($dest_file_path, $bytes)
    	Remove-Item $tmp_file_path -Force -ErrorAction SilentlyContinue
    } else {
    	echo $null > $dest_file_path
    }
  `, psGuestFilePath, dst)))

	return err
}

func (f *fileManager) UploadDir(dst string, src string) error {
	// We need these dirs later when walking files
	f.guestUploadDir = dst
	f.hostUploadDir = src

	// Walk all files in the src directory on the host system
	return filepath.Walk(src, f.walkFile)
}

func (f *fileManager) walkFile(hostPath string, hostFileInfo os.FileInfo, err error) error {
	if err == nil && shouldUploadFile(hostFileInfo) {
		relPath := filepath.Dir(hostPath[len(f.hostUploadDir):len(hostPath)])
		guestPath := filepath.Join(f.guestUploadDir, relPath, hostFileInfo.Name())
		err = f.UploadFile(guestPath, hostPath)
	}
	return err
}

func (f *fileManager) runCommand(cmd string) error {
	remoteCmd := &packer.RemoteCmd{
		Command: cmd,
	}

	err := f.comm.StartUnelevated(remoteCmd)
	if err != nil {
		return err
	}
	remoteCmd.Wait()

	if remoteCmd.ExitStatus != 0 {
		return errors.New("A file upload command failed with a non-zero exit code")
	}

	return nil
}

func winFriendlyPath(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}

func shouldUploadFile(hostFile os.FileInfo) bool {
	// Ignore dir entries and OS X special hidden file
	return !hostFile.IsDir() && ".DS_Store" != hostFile.Name()
}

func encodeChunks(bytes []byte, chunkSize int) []string {
	text := base64.StdEncoding.EncodeToString(bytes)
	reader := strings.NewReader(text)

	var chunks []string
	chunk := make([]byte, chunkSize)

	for {
		n, _ := reader.Read(chunk)
		if n == 0 {
			break
		}

		chunks = append(chunks, string(chunk[:n]))
	}

	return chunks
}
