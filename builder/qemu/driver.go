package qemu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type DriverCancelCallback func(state multistep.StateBag) bool

// A driver is able to talk to VirtualBox and perform certain
// operations with it.
type Driver interface {
	// Initializes the driver with the given values:
	// Arguments: qemuPath - string value for the qemu-system-x86_64 executable
	//            qemuImgPath - string value for the qemu-img executable
	Initialize(string, string)

	// Checks if the VM with the given name is running.
	IsRunning(string) (bool, error)

	// Stop stops a running machine, forcefully.
	Stop(string) error

	// SuppressMessages should do what needs to be done in order to
	// suppress any annoying popups from VirtualBox.
	SuppressMessages() error

	// Qemu executes the given command via qemu-system-x86_64
	Qemu(vmName string, qemuArgs ...string) error

	// wait on shutdown of the VM with option to cancel
	WaitForShutdown(
		vmName string,
		block bool,
		state multistep.StateBag,
		cancellCallback DriverCancelCallback) error

	// Qemu executes the given command via qemu-img
	QemuImg(...string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Version reads the version of VirtualBox that is installed.
	Version() (string, error)
}

type driverState struct {
	cmd        *exec.Cmd
	cancelChan chan struct{}
	waitDone   chan error
}

type QemuDriver struct {
	qemuPath    string
	qemuImgPath string
	state       map[string]*driverState
}

func (d *QemuDriver) getDriverState(name string) *driverState {
	if _, ok := d.state[name]; !ok {
		d.state[name] = &driverState{}
	}
	return d.state[name]
}

func (d *QemuDriver) Initialize(qemuPath string, qemuImgPath string) {
	d.qemuPath = qemuPath
	d.qemuImgPath = qemuImgPath
	d.state = make(map[string]*driverState)
}

func (d *QemuDriver) IsRunning(name string) (bool, error) {
	ds := d.getDriverState(name)
	return ds.cancelChan != nil, nil
}

func (d *QemuDriver) Stop(name string) error {
	ds := d.getDriverState(name)

	// signal to the command 'wait' to kill the process
	if ds.cancelChan != nil {
		close(ds.cancelChan)
		ds.cancelChan = nil
	}
	return nil
}

func (d *QemuDriver) SuppressMessages() error {
	return nil
}

func (d *QemuDriver) Qemu(vmName string, qemuArgs ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing %s: %#v", d.qemuPath, qemuArgs)
	ds := d.getDriverState(vmName)
	ds.cmd = exec.Command(d.qemuPath, qemuArgs...)
	ds.cmd.Stdout = &stdout
	ds.cmd.Stderr = &stderr

	err := ds.cmd.Start()

	if err != nil {
		err = fmt.Errorf("Error starting VM: %s", err)
	} else {
		log.Printf("---- Started Qemu ------- PID = ", ds.cmd.Process.Pid)

		ds.cancelChan = make(chan struct{})

		// make the channel to watch the process
		ds.waitDone = make(chan error)

		// start the virtual machine in the background
		go func() {
			ds.waitDone <- ds.cmd.Wait()
		}()
	}

	return err
}

func (d *QemuDriver) WaitForShutdown(vmName string,
	block bool,
	state multistep.StateBag,
	cancelCallback DriverCancelCallback) error {
	var err error

	ds := d.getDriverState(vmName)

	if block {
		// wait in the background for completion or caller cancel
		for {
			select {
			case <-ds.cancelChan:
				log.Println("Qemu process request to cancel -- killing Qemu process.")
				if err = ds.cmd.Process.Kill(); err != nil {
					log.Printf("Failed to kill qemu: %v", err)
				}

				// clear out the error channel since it's just a cancel
				// and therefore the reason for failure is clear
				log.Println("Empytying waitDone channel.")
				<-ds.waitDone

				// this gig is over -- assure calls to IsRunning see the nil
				log.Println("'Nil'ing out cancelChan.")
				ds.cancelChan = nil
				return errors.New("WaitForShutdown cancelled")
			case err = <-ds.waitDone:
				log.Printf("Qemu Process done with output = %v", err)
				// assure calls to IsRunning see the nil
				log.Println("'Nil'ing out cancelChan.")
				ds.cancelChan = nil
				return nil
			case <-time.After(1 * time.Second):
				cancel := cancelCallback(state)
				if cancel {
					log.Println("Qemu process request to cancel -- killing Qemu process.")

					// The step sequence was cancelled, so cancel waiting for SSH
					// and just start the halting process.
					close(ds.cancelChan)

					log.Println("Cancel request made, quitting waiting for Qemu.")
					return errors.New("WaitForShutdown cancelled by interrupt.")
				}
			}
		}
	} else {
		go func() {
			select {
			case <-ds.cancelChan:
				log.Println("Qemu process request to cancel -- killing Qemu process.")
				if err = ds.cmd.Process.Kill(); err != nil {
					log.Printf("Failed to kill qemu: %v", err)
				}

				// clear out the error channel since it's just a cancel
				// and therefore the reason for failure is clear
				log.Println("Empytying waitDone channel.")
				<-ds.waitDone
				log.Println("'Nil'ing out cancelChan.")
				ds.cancelChan = nil

			case err = <-ds.waitDone:
				log.Printf("Qemu Process done with output = %v", err)
				log.Println("'Nil'ing out cancelChan.")
				ds.cancelChan = nil
			}
		}()
	}

	ds.cancelChan = nil
	return err
}

func (d *QemuDriver) QemuImg(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing qemu-img: %#v", args)
	cmd := exec.Command(d.qemuImgPath, args...)
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

	cmd := exec.Command(d.qemuPath, "-version")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(stdout.String())
	log.Printf("Qemu --version output: %s", versionOutput)
	versionRe := regexp.MustCompile("qemu-kvm-[0-9]\\.[0-9]")
	matches := versionRe.Split(versionOutput, 2)
	if len(matches) == 0 {
		return "", fmt.Errorf("No version found: %s", versionOutput)
	}

	log.Printf("Qemu version: %s", matches[0])
	return matches[0], nil
}
