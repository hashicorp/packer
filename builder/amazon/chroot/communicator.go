package chroot

//  pf := func () { somefunc("a str", 1) } 

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Command func(string) *exec.Cmd

// Communicator is a special communicator that works by executing
// commands locally but within a chroot.
type Communicator struct {
	Chroot         string
	ChrootCmd      Command
	wrappedCommand Command
}

func (c *Communicator) Start(cmd *packer.RemoteCmd) error {
	localCmd := c.ChrootCmd(cmd.Command)
	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr
	log.Printf("Executing: %s %#v", localCmd.Path, localCmd.Args)
	if err := localCmd.Start(); err != nil {
		return err
	}

	go func() {
		exitStatus := 0
		if err := localCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitStatus = 1

				// There is no process-independent way to get the REAL
				// exit status so we just try to go deeper.
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				}
			}
		}

		log.Printf(
			"Chroot executation ended with '%d': '%s'",
			exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *Communicator) Upload(dst string, r io.Reader) error {
	dst = filepath.Join(c.Chroot, dst)
	log.Printf("Uploading to chroot dir: %s", dst)
	tf, err := ioutil.TempFile("", "packer-amazon-chroot")
	if err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err)
	}
	defer os.Remove(tf.Name())
	io.Copy(tf, r)
	cpCmd := fmt.Sprintf("cp %s %s", dst, tf.Name())
	return (*c.ChrootCmd(cpCmd)).Run()
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	walkFn := func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		path, err := filepath.Rel(src, fullPath)
		if err != nil {
			return err
		}

		for _, e := range exclude {
			if e == path {
				log.Printf("Skipping excluded file: %s", path)
				return nil
			}
		}

		chrootDest := filepath.Join(c.Chroot, dst, path)
		log.Printf("Uploading to chroot dir: %s", dst)
		cpCmd := fmt.Sprintf("cp %s %s", chrootDest, fullPath)
		return c.ChrootCmd(cpCmd).Run()
	}

	log.Printf("Uploading directory '%s' to '%s'", src, dst)
	return filepath.Walk(src, walkFn)
}

func (c *Communicator) Download(src string, w io.Writer) error {
	src = filepath.Join(c.Chroot, src)
	log.Printf("Downloading from chroot dir: %s", src)
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return err
	}

	return nil
}
