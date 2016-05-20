package qemu

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/mitchellh/multistep"
)

type DriverCancelCallback func(state multistep.StateBag) bool

// A driver is able to talk to qemu-system-x86_64 and perform certain
// operations with it.
type Driver interface {
	// Stop stops a running machine, forcefully.
	Stop() error

	// Qemu executes the given command via qemu-system-x86_64
	Qemu(qemuArgs ...string) error

	// wait on shutdown of the VM with option to cancel
	WaitForShutdown(<-chan struct{}) bool

	// Qemu executes the given command via qemu-img
	QemuImg(...string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of Qemu that is installed.
	Version() (string, error)
}

type QemuDriver struct {
	QemuPath    string
	QemuImgPath string

	vmCmd   *exec.Cmd
	vmEndCh <-chan int
	lock    sync.Mutex
}

func (d *QemuDriver) Stop() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.vmCmd != nil {
		if err := d.vmCmd.Process.Kill(); err != nil {
			return err
		}
	}

	return nil
}

func (d *QemuDriver) Qemu(qemuArgs ...string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.vmCmd != nil {
		panic("Existing VM state found")
	}

	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	log.Printf("Executing %s: %#v", d.QemuPath, qemuArgs)
	cmd := exec.Command(d.QemuPath, qemuArgs...)
	cmd.Stdout = stdout_w
	cmd.Stderr = stderr_w

	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("Error starting VM: %s", err)
		return err
	}

	go logReader("Qemu stdout", stdout_r)
	go logReader("Qemu stderr", stderr_r)

	log.Printf("Started Qemu. Pid: %d", cmd.Process.Pid)

	// Wait for Qemu to complete in the background, and mark when its done
	endCh := make(chan int, 1)
	go func() {
		defer stderr_w.Close()
		defer stdout_w.Close()

		var exitCode int = 0
		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				} else {
					exitCode = 254
				}
			}
		}

		endCh <- exitCode

		d.lock.Lock()
		defer d.lock.Unlock()
		d.vmCmd = nil
		d.vmEndCh = nil
	}()

	// Wait at least a couple seconds for an early fail from Qemu so
	// we can report that.
	select {
	case exit := <-endCh:
		if exit != 0 {
			return fmt.Errorf("Qemu failed to start. Please run with PACKER_LOG=1 to get more info.")
		}
	case <-time.After(2 * time.Second):
	}

	// Setup our state so we know we are running
	d.vmCmd = cmd
	d.vmEndCh = endCh

	return nil
}

func (d *QemuDriver) WaitForShutdown(cancelCh <-chan struct{}) bool {
	d.lock.Lock()
	endCh := d.vmEndCh
	d.lock.Unlock()

	if endCh == nil {
		return true
	}

	select {
	case <-endCh:
		return true
	case <-cancelCh:
		return false
	}
}

func (d *QemuDriver) QemuImg(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing qemu-img: %#v", args)
	cmd := exec.Command(d.QemuImgPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("QemuImg error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}

func (d *QemuDriver) Verify() error {
	return nil
}

func (d *QemuDriver) Version() (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.QemuPath, "-version")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(stdout.String())
	log.Printf("Qemu --version output: %s", versionOutput)
	versionRe := regexp.MustCompile("[\\.[0-9]+]*")
	matches := versionRe.FindStringSubmatch(versionOutput)
	if len(matches) == 0 {
		return "", fmt.Errorf("No version found: %s", versionOutput)
	}

	log.Printf("Qemu version: %s", matches[0])
	return matches[0], nil
}

func logReader(name string, r io.Reader) {
	bufR := bufio.NewReader(r)
	for {
		line, err := bufR.ReadString('\n')
		if line != "" {
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			log.Printf("%s: %s", name, line)
		}

		if err == io.EOF {
			break
		}
	}
}
