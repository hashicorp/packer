package getter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// SmbClientGetter is a Getter implementation that will download a module from
// a shared folder using smbclient cli.
type SmbClientGetter struct{}

func (g *SmbClientGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, new(smbPathError)
	}

	// Use smbclient cli to verify mode
	mode, err := g.smbClientMode(u)
	if err == nil {
		return mode, nil
	}
	return 0, &smbGeneralError{err}
}

func (g *SmbClientGetter) smbClientMode(u *url.URL) (Mode, error) {
	hostPath, filePath, err := g.findHostAndFilePath(u)
	if err != nil {
		return 0, err
	}
	file := ""
	// Get file and subdirectory name when existent
	if strings.Contains(filePath, "/") {
		i := strings.LastIndex(filePath, "/")
		file = filePath[i+1:]
		filePath = filePath[:i]
	} else {
		file = filePath
		filePath = "."
	}

	cmdArgs := g.smbclientCmdArgs(u.User, hostPath, filePath)
	// check if file exists in the smb shared folder and check the mode
	isDir, err := g.isDirectory(cmdArgs, file)
	if err != nil {
		return 0, err
	}
	if isDir {
		return ModeDir, nil
	}
	return ModeFile, nil
}

func (g *SmbClientGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	// If dst folder doesn't exists, we need to remove the created on later in case of failures
	dstExisted := false
	if req.Dst != "" {
		if _, err := os.Lstat(req.Dst); err == nil {
			dstExisted = true
		}
	}

	// Download the directory content using smbclient cli
	err := g.smbclientGet(req)
	if err == nil {
		return nil
	}

	if !dstExisted {
		// Remove the destination created for smbclient
		os.Remove(req.Dst)
	}

	return &smbGeneralError{err}
}

func (g *SmbClientGetter) smbclientGet(req *Request) error {
	hostPath, directory, err := g.findHostAndFilePath(req.u)
	if err != nil {
		return err
	}

	cmdArgs := g.smbclientCmdArgs(req.u.User, hostPath, ".")
	// check directory exists in the smb shared folder and is a directory
	isDir, err := g.isDirectory(cmdArgs, directory)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("%s source path must be a directory", directory)
	}

	// download everything that's inside the directory (files and subdirectories)
	cmdArgs = append(cmdArgs, "-c")
	cmdArgs = append(cmdArgs, "prompt OFF;recurse ON; mget *")

	if req.Dst != "" {
		_, err := os.Lstat(req.Dst)
		if err != nil {
			if os.IsNotExist(err) {
				// Create destination folder if it doesn't exist
				if err := os.MkdirAll(req.Dst, 0755); err != nil {
					return fmt.Errorf("failed to create destination path: %s", err.Error())
				}
			} else {
				return err
			}
		}
	}

	_, err = g.runSmbClientCommand(req.Dst, cmdArgs)
	return err
}

func (g *SmbClientGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	// If dst folder doesn't exist, we need to remove the created one later in case of failures
	dstExisted := false
	if req.Dst != "" {
		if _, err := os.Lstat(req.Dst); err == nil {
			dstExisted = true
		}
	}

	// If not mounted, try downloading the file using smbclient cli
	err := g.smbclientGetFile(req)
	if err == nil {
		return nil
	}

	if !dstExisted {
		// Remove the destination created for smbclient
		os.Remove(req.Dst)
	}

	return &smbGeneralError{err}
}

func (g *SmbClientGetter) smbclientGetFile(req *Request) error {
	hostPath, filePath, err := g.findHostAndFilePath(req.u)
	if err != nil {
		return err
	}

	// Get file and subdirectory name when existent
	file := ""
	if strings.Contains(filePath, "/") {
		i := strings.LastIndex(filePath, "/")
		file = filePath[i+1:]
		filePath = filePath[:i]
	} else {
		file = filePath
		filePath = "."
	}

	cmdArgs := g.smbclientCmdArgs(req.u.User, hostPath, filePath)
	// check file exists in the smb shared folder and is not a directory
	isDir, err := g.isDirectory(cmdArgs, file)
	if err != nil {
		return err
	}
	if isDir {
		return fmt.Errorf("%s source path must be a file", file)
	}

	// download file
	cmdArgs = append(cmdArgs, "-c")
	if req.Dst != "" {
		_, err := os.Lstat(req.Dst)
		if err != nil {
			if os.IsNotExist(err) {
				// Create destination folder if it doesn't exist
				if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
					return fmt.Errorf("failed to creat destination path: %s", err.Error())
				}
			} else {
				return err
			}
		}
		cmdArgs = append(cmdArgs, fmt.Sprintf("get %s %s", file, req.Dst))
	} else {
		cmdArgs = append(cmdArgs, fmt.Sprintf("get %s", file))
	}

	_, err = g.runSmbClientCommand("", cmdArgs)
	return err
}

func (g *SmbClientGetter) smbclientCmdArgs(used *url.Userinfo, hostPath string, fileDir string) (baseCmd []string) {
	baseCmd = append(baseCmd, "-N")

	// Append auth user and password to baseCmd
	auth := used.Username()
	if auth != "" {
		if password, ok := used.Password(); ok {
			auth = auth + "%" + password
		}
		baseCmd = append(baseCmd, "-U")
		baseCmd = append(baseCmd, auth)
	}

	baseCmd = append(baseCmd, hostPath)
	baseCmd = append(baseCmd, "--directory")
	baseCmd = append(baseCmd, fileDir)
	return baseCmd
}

func (g *SmbClientGetter) findHostAndFilePath(u *url.URL) (string, string, error) {
	// Host path
	hostPath := "//" + u.Host

	// Get shared directory
	path := strings.TrimPrefix(u.Path, "/")
	splt := regexp.MustCompile(`/`)
	directories := splt.Split(path, 2)

	if len(directories) > 0 {
		hostPath = hostPath + "/" + directories[0]
	}

	// Check file path
	if len(directories) <= 1 || directories[1] == "" {
		return "", "", fmt.Errorf("can not find file path and/or name in the smb url")
	}

	return hostPath, directories[1], nil
}

func (g *SmbClientGetter) isDirectory(args []string, object string) (bool, error) {
	args = append(args, "-c")
	args = append(args, fmt.Sprintf("allinfo %s", object))
	output, err := g.runSmbClientCommand("", args)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "OBJECT_NAME_NOT_FOUND") {
		return false, fmt.Errorf("source path not found: %s", output)
	}
	return strings.Contains(output, "attributes: D"), nil
}

func (g *SmbClientGetter) runSmbClientCommand(dst string, args []string) (string, error) {
	cmd := exec.Command("smbclient", args...)

	if dst != "" {
		cmd.Dir = dst
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err == nil {
		return buf.String(), nil
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return buf.String(), fmt.Errorf(
				"%s exited with %d: %s",
				cmd.Path,
				status.ExitStatus(),
				buf.String())
		}
	}
	return buf.String(), fmt.Errorf("error running %s: %s", cmd.Path, buf.String())
}

func (g *SmbClientGetter) Detect(req *Request) (bool, error) {
	if len(req.Src) == 0 {
		return false, nil
	}

	if req.Forced != "" {
		// There's a getter being Forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the Forced one
			// Don't use it to try to download the artifact
			return false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(req.Src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the Forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return false, nil
	}

	return false, nil
}

func (g *SmbClientGetter) validScheme(scheme string) bool {
	return scheme == "smb"
}

type smbPathError struct {
	Path string
}

func (e *smbPathError) Error() string {
	if e.Path == "" {
		return "samba path should contain valid host, filepath, and authentication if necessary (smb://<user>:<password>@<host>/<file_path>)"
	}
	return fmt.Sprintf("samba path should contain valid host, filepath, and authentication if necessary (%s)", e.Path)
}

type smbGeneralError struct {
	err error
}

func (e *smbGeneralError) Error() string {
	if e != nil {
		return fmt.Sprintf("smbclient cli needs to be installed and credentials provided when necessary. \n err: %s", e.err.Error())
	}
	return "smbclient cli needs to be installed and credentials provided when necessary."
}
