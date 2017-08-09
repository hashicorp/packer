package docker

import (
	"archive/tar"
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
	"github.com/hashicorp/packer/packer"
)

type Communicator struct {
	ContainerId  string
	HostDir      string
	ContainerDir string
	Version      *version.Version
	Config       *Config
	lock         sync.Mutex
}

func (c *Communicator) Start(remote *packer.RemoteCmd) error {
	var cmd *exec.Cmd
	if c.Config.Pty {
		cmd = exec.Command("docker", "exec", "-i", "-t", c.ContainerId, "/bin/sh", "-c", fmt.Sprintf("(%s)", remote.Command))
	} else {
		cmd = exec.Command("docker", "exec", "-i", c.ContainerId, "/bin/sh", "-c", fmt.Sprintf("(%s)", remote.Command))
	}

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

func (c *Communicator) Upload(dst string, src io.Reader, fi *os.FileInfo) error {
	// Create a temporary file to store the upload
	tempfile, err := ioutil.TempFile(c.HostDir, "upload")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	// Copy the contents to the temporary file
	_, err = io.Copy(tempfile, src)
	if err != nil {
		return err
	}

	if fi != nil {
		tempfile.Chmod((*fi).Mode())
	}
	tempfile.Close()

	// Copy the file into place by copying the temporary file we put
	// into the shared folder into the proper location in the container
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("command cp %s/%s %s", c.ContainerDir,
			filepath.Base(tempfile.Name()), dst),
	}

	if err := c.Start(cmd); err != nil {
		return err
	}

	// Wait for the copy to complete
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Upload failed with non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	// Create the temporary directory that will store the contents of "src"
	// for copying into the container.
	td, err := ioutil.TempDir(c.HostDir, "dirupload")
	if err != nil {
		return err
	}
	defer os.RemoveAll(td)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		hostpath := filepath.Join(td, relpath)

		// If it is a directory, just create it
		if info.IsDir() {
			return os.MkdirAll(hostpath, info.Mode())
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			dest, err := os.Readlink(path)

			if err != nil {
				return err
			}

			return os.Symlink(dest, hostpath)
		}

		// It is a file, copy it over, including mode.
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(hostpath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		si, err := src.Stat()
		if err != nil {
			return err
		}

		return dst.Chmod(si.Mode())
	}

	// Copy the entire directory tree to the temporary directory
	if err := filepath.Walk(src, walkFn); err != nil {
		return err
	}

	// Determine the destination directory
	containerSrc := filepath.Join(c.ContainerDir, filepath.Base(td))
	containerDst := dst
	if src[len(src)-1] != '/' {
		containerDst = filepath.Join(dst, filepath.Base(src))
	}

	// Make the directory, then copy into it
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("set -e; mkdir -p %s; command cp -R %s/ %s",
			containerDst, containerSrc, containerDst),
	}
	if err := c.Start(cmd); err != nil {
		return err
	}

	// Wait for the copy to complete
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Upload failed with non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

// Download pulls a file out of a container using `docker cp`. We have a source
// path and want to write to an io.Writer, not a file. We use - to make docker
// cp to write to stdout, and then copy the stream to our destination io.Writer.
func (c *Communicator) Download(src string, dst io.Writer) error {
	log.Printf("Downloading file from container: %s:%s", c.ContainerId, src)
	localCmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", c.ContainerId, src), "-")

	pipe, err := localCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to open pipe: %s", err)
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
func (c *Communicator) run(cmd *exec.Cmd, remote *packer.RemoteCmd, stdin io.WriteCloser, stdout, stderr io.ReadCloser) {
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
