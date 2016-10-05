package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/packer/packer"
)

// Communicator is the Docker implementation for packer.Communicator.
type Communicator struct {
	ContainerID string
	Version     *version.Version
	Config      *Config
	lock        sync.Mutex
	Client      *docker.Client
}

// newDockerClient gets a new docker.Client for use with the communicator.
func newDockerClient() (*docker.Client, error) {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		log.Printf("[DEBUG] Setting up Docker connection to %s", host)
		client, err := docker.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	var socket string
	// Handle windows
	if runtime.GOOS == "windows" {
		socket = "npipe:////./pipe/docker_engine"
	} else {
		socket = "unix:///var/run/docker.sock"
	}
	log.Printf("[DEBUG] Setting up Docker connection to: %s", socket)
	client, err := docker.NewClient(socket)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Start takes a remote command and starts it, using the functionality in the API.
func (c *Communicator) Start(remote *packer.RemoteCmd) error {
	log.Printf("[DEBUG] Executing \"%s\" in container ID: %s", remote.Command, c.ContainerID)

	// We need to determine how to run the command, either via sh -c or via cmd.exe /c.
	var cmd []string
	info, err := c.Client.Info()
	if err != nil {
		return err
	}
	if info.OSType == "windows" {
		cmd = []string{"cmd.exe", "/c", remote.Command}
	} else {
		cmd = []string{"sh", "-c", remote.Command}
	}

	createOpts := docker.CreateExecOptions{
		Cmd:       cmd,
		Container: c.ContainerID,
	}

	startOpts := docker.StartExecOptions{
		Detach: false,
	}

	if remote.Stdout != nil {
		createOpts.AttachStdout = true
		startOpts.OutputStream = remote.Stdout
	}
	if remote.Stdin != nil {
		createOpts.AttachStdin = true
		startOpts.InputStream = remote.Stdin
	}
	if remote.Stderr != nil {
		createOpts.AttachStderr = true
		startOpts.ErrorStream = remote.Stderr
	}
	if c.Config.Pty {
		createOpts.Tty = true
		startOpts.Tty = true
	}

	log.Printf("[DEBUG] Creating exec instance for command \"%s\" in container ID: %s", remote.Command, c.ContainerID)
	exec, err := c.Client.CreateExec(createOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Starting exec instance %s for container ID: %s", exec.ID, c.ContainerID)
	waiter, err := c.Client.StartExecNonBlocking(exec.ID, startOpts)
	if err != nil {
		return err
	}
	// Wait for the command to complete in a goroutine so that we don't block.
	go func() {
		log.Printf("[DEBUG] Waiting for exec instance ID %s to finish", exec.ID)
		waiter.Wait()
		log.Printf("[DEBUG] Exec complete for exec instance ID %s, getting status", exec.ID)
		inspect, err := c.Client.InspectExec(exec.ID)
		if err != nil {
			log.Printf("Error inspecting exec instance after exit: %s - setting status 254", err)
			remote.SetExited(254)
			return
		}
		log.Printf("[DEBUG] Exit code %d for exec instance ID %s", inspect.ExitCode, exec.ID)
		remote.SetExited(inspect.ExitCode)
	}()
	return nil
}

// uploadDockerTar takes a tar file (implementing io.Reader) and uploads it
// to a remote container using docker.Client.UploadToContainer().
func (c *Communicator) uploadDockerTar(buf *bytes.Buffer, dst string) error {
	log.Printf("[DEBUG] Uploading tar stream to remote container/path: %s:%s", c.ContainerID, dst)

	opts := docker.UploadToContainerOptions{
		InputStream: buf,
		Path:        dst,
	}

	if err := c.Client.UploadToContainer(c.ContainerID, opts); err != nil {
		return err
	}
	log.Printf("[DEBUG] Successful upload of tar stream to remote container/path: %s:%s", c.ContainerID, dst)
	return nil
}

// downloadDockerTar reads a tar file from a remote container,
// created by archiving a supplied remote directory.
func (c *Communicator) downloadDockerTar(src string) (*bytes.Buffer, error) {
	log.Printf("[DEBUG] Reading tar stream from remote container/path: %s:%s", c.ContainerID, src)

	buf := new(bytes.Buffer)
	opts := docker.DownloadFromContainerOptions{
		OutputStream:      buf,
		Path:              src,
		InactivityTimeout: time.Duration(time.Second * 30),
	}

	if err := c.Client.DownloadFromContainer(c.ContainerID, opts); err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Successful download of tar stream from remote container/path: %s:%s", c.ContainerID, src)
	return buf, nil
}

// tarFromFileStream makes a tar archive stream of a single file, which can
// then be passed to uploadDockerTar. The name of the file in the tar is
// given by the name parameter.
func tarFromFileStream(data *bytes.Buffer, fi *os.FileInfo, name string) (*bytes.Buffer, error) {
	log.Printf("[DEBUG] Creating upload tar of stream with internal file name: %s", name)
	buf := new(bytes.Buffer)
	hdr := new(tar.Header)
	var err error
	tw := tar.NewWriter(buf)
	if fi != nil {
		hdr, err = tar.FileInfoHeader(*fi, (*fi).Name())
		if err != nil {
			return nil, err
		}
		// strip owner and group from header. since this is a file upload
		// preservation of this data is probably not intended.
		hdr.Uid = 0
		hdr.Gid = 0
		hdr.Uname = ""
		hdr.Gname = ""
	} else {
		hdr.Size = int64(data.Len())
		hdr.Mode = 0644
		hdr.Linkname = name
	}
	hdr.Name = name

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	n, err := io.Copy(tw, data)
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Successful creation of upload tar (%d bytes) with internal file name: %s", n, name)
	return buf, nil
}

// tarFromDirectory makes a tar archive stream of a directory, which can
// then be passed to uploadDockerTar.
func tarFromDirectory(src string, exclude []string) (*bytes.Buffer, error) {
	log.Printf("[DEBUG] Creating upload tar for source directory %s", src)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	if _, err := os.Stat(src); err != nil {
		return nil, err
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		log.Printf("[DEBUG] walkFn: raw path: %s", path)
		relative, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if len(relative) < 1 || relative == "." {
			log.Printf("[DEBUG] skipping root path %s for upload tar", relative)
			return nil
		}

		for _, v := range exclude {
			if relative == v {
				log.Printf("[DEBUG] skipping excluded file path %s for upload tar", relative)
				return nil
			}
		}

		log.Printf("[DEBUG] Adding file or directory %s to upload tar", relative)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		hdr.Name = relative

		// strip owner and group from header. since this is a file upload
		// preservation of this data is probably not intended.
		hdr.Uid = 0
		hdr.Gid = 0
		hdr.Uname = ""
		hdr.Gname = ""

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.Mode().IsRegular() == false {
			log.Printf("[DEBUG] %s is not a regular file, no IO write necessary", relative)
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err := io.Copy(tw, src); err != nil {
			return err
		}

		log.Printf("[DEBUG] Successful write for file %s in upload tar", relative)
		return nil
	}

	// Copy the entire directory tree to the tar
	if err := filepath.Walk(src, walkFn); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Success creating upload tar for source directory %s", src)
	return buf, nil
}

// tarToStream reads the first file from a tar buffer indiscriminately, and
// returns the data as an io.Writer.
func tarToStream(data *bytes.Buffer) (*bytes.Buffer, error) {
	log.Println("[DEBUG] Fetching data for first file in tar only")

	tr := tar.NewReader(data)
	h, err := tr.Next()
	if err != nil {
		return nil, fmt.Errorf("Failed to read header from tar stream: %s", err)
	}

	buf := new(bytes.Buffer)

	numBytes, err := io.Copy(buf, tr)
	if err != nil {
		return nil, fmt.Errorf("Failed to pipe download: %s", err)
	}
	log.Printf("[DEBUG] Successfully copied %d bytes for %s", numBytes, h.Name)
	return buf, nil
}

// tarToDirectory extracts an entire tar to a destination directory.
//
// strip will strip n levels of path from the extracted file, by splitting
// first with filepath.Split(), splitting that path off of os.PathSeparator,
// and then re-assembling the path starting from [strip:].
func tarToDirectory(data *bytes.Buffer, dst string, exclude []string, strip int) error {
	log.Printf("[DEBUG] Extracting tar to destination: %s", dst)

	tr := tar.NewReader(data)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Failed to read header from tar stream: %s", err)
		}

		// compute actual path from within tar, based off strip parameter.
		dir, file := filepath.Split(h.Name)
		dirs := strings.Split(dir, string(os.PathSeparator))
		name := filepath.Join(strings.Join(dirs[strip:], string(os.PathSeparator)), file)

		for _, v := range exclude {
			if name == v {
				log.Printf("[DEBUG] skipping excluded file path %s in download tar", name)
				return nil
			}
		}

		path := filepath.Join(dst, name)

		// strip owner and group from header. since this is a file upload
		// preservation of this data is probably not intended.
		h.Uid = 0
		h.Gid = 0
		h.Uname = ""
		h.Gname = ""

		info := h.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}
		fh, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer fh.Close()

		n, err := io.Copy(fh, tr)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Wrote %d bytes for %s", n, name)
	}

	log.Println("[DEBUG] Extracting tar completed")
	return nil
}

// Upload uses docker.Client.UploadToContainer upload a file to the container.
// The single file is streamed to a tar archive first as it is needed by the
// Docker API.
func (c *Communicator) Upload(dst string, src io.Reader, fi *os.FileInfo) error {
	var fileName string
	if fi == nil {
		fileName = "<streamed data>"
	} else {
		fileName = (*fi).Name()
	}
	log.Printf("[DEBUG] Uploading %s to remote container/path: %s:%s", fileName, c.ContainerID, dst)

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, src); err != nil {
		return fmt.Errorf("Error copying source buffer: %s", err.Error())
	}

	buf, err := tarFromFileStream(data, fi, filepath.Base(dst))
	if err != nil {
		return err
	}
	if err := c.uploadDockerTar(buf, filepath.Dir(dst)); err != nil {
		return err
	}

	log.Printf("[DEBUG] Upload successful: %s to remote container/path: %s:%s", fileName, c.ContainerID, dst)
	return nil
}

// UploadDir uses docker.Client.UploadToContainer upload an entire directory
// to the container. The directory is streamed into a tar archive first.
func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	log.Printf("[DEBUG] Uploading directory %s to remote container/path: %s:%s", src, c.ContainerID, dst)
	buf, err := tarFromDirectory(src, exclude)
	if err != nil {
		return err
	}

	if err := c.uploadDockerTar(buf, dst); err != nil {
		return err
	}

	log.Printf("[DEBUG] Directory upload successful: %s to remote container/path: %s:%s", src, c.ContainerID, dst)
	return nil
}

// Download pulls a file out of a container using `docker cp`. We have a source
// path and want to write to an io.Writer, not a file. We use - to make docker
// cp to write to stdout, and then copy the stream to our destination io.Writer.
func (c *Communicator) Download(src string, dst io.Writer) error {
	log.Printf("[DEBUG] Downloading file from container: %s:%s", c.ContainerID, src)

	buf, err := c.downloadDockerTar(src)
	if err != nil {
		return err
	}

	data, err := tarToStream(buf)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, data); err != nil {
		return err
	}

	log.Printf("[DEBUG] Successful download of file from container: %s:%s", c.ContainerID, src)
	return nil
}

// DownloadDir downloads a directory from a Docker container to a tar file.
// The file is then extracted to the destination.
func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	log.Printf("[DEBUG] Downloading directory from container: %s:%s to %s", c.ContainerID, src, dst)

	buf, err := c.downloadDockerTar(src)
	if err != nil {
		return err
	}

	// The docker API tars up directories using their full relative paths, so if
	// foo/ is downloaded for example, files underneath foo/ will bear names
	// like foo/bar. So a strip level of 1 needs to be passed.
	if err := tarToDirectory(buf, dst, exclude, 1); err != nil {
		return err
	}

	log.Printf("[DEBUG] Successful download of directory from container: %s:%s to %s", c.ContainerID, src, dst)
	return nil
}
