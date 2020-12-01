package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/go-version"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Communicator struct {
	ContainerID   string
	HostDir       string
	ContainerDir  string
	Version       *version.Version
	Config        *Config
	ContainerUser string
	lock          sync.Mutex
	EntryPoint    []string
}

var _ packersdk.Communicator = new(Communicator)

func (c *Communicator) Start(ctx context.Context, remote *packersdk.RemoteCmd) error {
	dockerArgs := []string{
		"exec",
		"-i",
		c.ContainerID,
	}
	dockerArgs = append(dockerArgs, c.EntryPoint...)
	dockerArgs = append(dockerArgs, fmt.Sprintf("(%s)", remote.Command))

	if c.Config.Pty {
		dockerArgs = append(dockerArgs[:2], append([]string{"-t"}, dockerArgs[2:]...)...)
	}

	if c.Config.ExecUser != "" {
		dockerArgs = append(dockerArgs[:2],
			append([]string{"-u", c.Config.ExecUser}, dockerArgs[2:]...)...)
	}

	cmd := exec.Command("docker", dockerArgs...)

	var (
		stdin_w io.WriteCloser
		err     error
	)

	stdin_w, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	stderr_r, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout_r, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Run the actual command in a goroutine so that Start doesn't block
	go c.run(cmd, remote, stdin_w, stdout_r, stderr_r)

	return nil
}

// Upload uploads a file to the docker container
func (c *Communicator) Upload(dst string, src io.Reader, fi *os.FileInfo) error {
	if fi == nil {
		return c.uploadReader(dst, src)
	}
	return c.uploadFile(dst, src, fi)
}

// uploadReader writes an io.Reader to a temporary file before uploading
func (c *Communicator) uploadReader(dst string, src io.Reader) error {
	// Create a temporary file to store the upload
	tempfile, err := ioutil.TempFile(c.HostDir, "upload")
	if err != nil {
		return fmt.Errorf("Failed to open temp file for writing: %s", err)
	}
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	if _, err := io.Copy(tempfile, src); err != nil {
		return fmt.Errorf("Failed to copy upload file to tempfile: %s", err)
	}
	tempfile.Seek(0, 0)
	fi, err := tempfile.Stat()
	if err != nil {
		return fmt.Errorf("Error getting tempfile info: %s", err)
	}
	return c.uploadFile(dst, tempfile, &fi)
}

// uploadFile uses docker cp to copy the file from the host to the container
func (c *Communicator) uploadFile(dst string, src io.Reader, fi *os.FileInfo) error {
	// command format: docker cp /path/to/infile containerid:/path/to/outfile
	log.Printf("Copying to %s on container %s.", dst, c.ContainerID)

	localCmd := exec.Command("docker", "cp", "-",
		fmt.Sprintf("%s:%s", c.ContainerID, filepath.Dir(dst)))

	stderrP, err := localCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to open pipe: %s", err)
	}

	stdin, err := localCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to open pipe: %s", err)
	}

	if err := localCmd.Start(); err != nil {
		return err
	}

	archive := tar.NewWriter(stdin)
	header, err := tar.FileInfoHeader(*fi, "")
	if err != nil {
		return err
	}
	header.Name = filepath.Base(dst)
	archive.WriteHeader(header)
	numBytes, err := io.Copy(archive, src)
	if err != nil {
		return fmt.Errorf("Failed to pipe upload: %s", err)
	}
	log.Printf("Copied %d bytes for %s", numBytes, dst)

	if err := archive.Close(); err != nil {
		return fmt.Errorf("Failed to close archive: %s", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("Failed to close stdin: %s", err)
	}

	stderrOut, err := ioutil.ReadAll(stderrP)
	if err != nil {
		return err
	}

	if err := localCmd.Wait(); err != nil {
		return fmt.Errorf("Failed to upload to '%s' in container: %s. %s.", dst, stderrOut, err)
	}

	if err := c.fixDestinationOwner(dst); err != nil {
		return err
	}

	return nil
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	/*
		from https://docs.docker.com/engine/reference/commandline/cp/#extended-description
		SRC_PATH specifies a directory
			DEST_PATH does not exist
				DEST_PATH is created as a directory and the contents of the source directory are copied into this directory
			DEST_PATH exists and is a file
				Error condition: cannot copy a directory to a file
			DEST_PATH exists and is a directory
				SRC_PATH does not end with /. (that is: slash followed by dot)
					the source directory is copied into this directory
				SRC_PATH does end with /. (that is: slash followed by dot)
					the content of the source directory is copied into this directory

		translating that in to our semantics:

		if source ends in /
			docker cp src. dest
		otherwise, cp source dest

	*/

	dockerSource := src
	if src[len(src)-1] == '/' {
		dockerSource = fmt.Sprintf("%s.", src)
	}

	// Make the directory, then copy into it
	localCmd := exec.Command("docker", "cp", dockerSource, fmt.Sprintf("%s:%s", c.ContainerID, dst))

	stderrP, err := localCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to open pipe: %s", err)
	}
	if err := localCmd.Start(); err != nil {
		return fmt.Errorf("Failed to copy: %s", err)
	}
	stderrOut, err := ioutil.ReadAll(stderrP)
	if err != nil {
		return err
	}

	// Wait for the copy to complete
	if err := localCmd.Wait(); err != nil {
		return fmt.Errorf("Failed to upload to '%s' in container: %s. %s.", dst, stderrOut, err)
	}

	if err := c.fixDestinationOwner(dst); err != nil {
		return err
	}

	return nil
}

// Download pulls a file out of a container using `docker cp`. We have a source
// path and want to write to an io.Writer, not a file. We use - to make docker
// cp to write to stdout, and then copy the stream to our destination io.Writer.
func (c *Communicator) Download(src string, dst io.Writer) error {
	log.Printf("Downloading file from container: %s:%s", c.ContainerID, src)
	localCmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", c.ContainerID, src), "-")

	pipe, err := localCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to open pipe: %s", err)
	}

	stderrP, err := localCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to open stderr pipe: %s", err)
	}

	if err = localCmd.Start(); err != nil {
		return fmt.Errorf("Failed to start download: %s", err)
	}

	// When you use - to send docker cp to stdout it is streamed as a tar; this
	// enables it to work with directories. We don't actually support
	// directories in Download() but we still need to handle the tar format.

	archive := tar.NewReader(pipe)
	_, err = archive.Next()
	if err != nil {
		// see if we can get a useful error message from stderr, since stdout
		// is messed up.
		if stderrOut, err := ioutil.ReadAll(stderrP); err == nil {
			if string(stderrOut) != "" {
				return fmt.Errorf("Error downloading file: %s", string(stderrOut))
			}
		}
		return fmt.Errorf("Failed to read header from tar stream: %s", err)
	}

	numBytes, err := io.Copy(dst, archive)
	if err != nil {
		return fmt.Errorf("Failed to pipe download: %s", err)
	}
	log.Printf("Copied %d bytes for %s", numBytes, src)

	if err = localCmd.Wait(); err != nil {
		return fmt.Errorf("Failed to download '%s' from container: %s", src, err)
	}

	return nil
}

func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("DownloadDir is not implemented for docker")
}

// Runs the given command and blocks until completion
func (c *Communicator) run(cmd *exec.Cmd, remote *packersdk.RemoteCmd, stdin io.WriteCloser, stdout, stderr io.ReadCloser) {
	// For Docker, remote communication must be serialized since it
	// only supports single execution.
	c.lock.Lock()
	defer c.lock.Unlock()

	wg := sync.WaitGroup{}
	repeat := func(w io.Writer, r io.ReadCloser) {
		io.Copy(w, r)
		r.Close()
		wg.Done()
	}

	if remote.Stdout != nil {
		wg.Add(1)
		go repeat(remote.Stdout, stdout)
	}

	if remote.Stderr != nil {
		wg.Add(1)
		go repeat(remote.Stderr, stderr)
	}

	// Start the command
	log.Printf("Executing %s:", strings.Join(cmd.Args, " "))
	if err := cmd.Start(); err != nil {
		log.Printf("Error executing: %s", err)
		remote.SetExited(254)
		return
	}

	var exitStatus int

	if remote.Stdin != nil {
		go func() {
			io.Copy(stdin, remote.Stdin)
			// close stdin to support commands that wait for stdin to be closed before exiting.
			stdin.Close()
		}()
	}

	wg.Wait()
	err := cmd.Wait()

	if exitErr, ok := err.(*exec.ExitError); ok {
		exitStatus = 1

		// There is no process-independent way to get the REAL
		// exit status so we just try to go deeper.
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
		}
	}

	// Set the exit status which triggers waiters
	remote.SetExited(exitStatus)
}

// TODO Workaround for #5307. Remove once #5409 is fixed.
func (c *Communicator) fixDestinationOwner(destination string) error {
	if !c.Config.FixUploadOwner {
		return nil
	}

	owner := c.ContainerUser
	if owner == "" {
		owner = "root"
	}

	chownArgs := []string{
		"docker", "exec", "--user", "root", c.ContainerID, "/bin/sh", "-c",
		fmt.Sprintf("chown -R %s %s", owner, destination),
	}
	if output, err := exec.Command(chownArgs[0], chownArgs[1:]...).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to set owner of the uploaded file: %s, %s", err, output)
	}

	return nil
}
