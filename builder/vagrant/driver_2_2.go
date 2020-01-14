package vagrant

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

const VAGRANT_MIN_VERSION = ">= 2.0.2"

type Vagrant_2_2_Driver struct {
	vagrantBinary string
	VagrantCWD    string
}

// Calls "vagrant init"
func (d *Vagrant_2_2_Driver) Init(args []string) error {
	_, _, err := d.vagrantCmd(append([]string{"init"}, args...)...)
	return err
}

// Calls "vagrant add"
func (d *Vagrant_2_2_Driver) Add(args []string) error {
	// vagrant box add partyvm ubuntu-14.04.vmware.box
	_, _, err := d.vagrantCmd(append([]string{"box", "add"}, args...)...)
	return err
}

// Calls "vagrant up"
func (d *Vagrant_2_2_Driver) Up(args []string) (string, string, error) {
	stdout, stderr, err := d.vagrantCmd(append([]string{"up"}, args...)...)
	return stdout, stderr, err
}

// Calls "vagrant halt"
func (d *Vagrant_2_2_Driver) Halt(id string) error {
	args := []string{"halt"}
	if id != "" {
		args = append(args, id)
	}
	_, _, err := d.vagrantCmd(args...)
	return err
}

// Calls "vagrant suspend"
func (d *Vagrant_2_2_Driver) Suspend(id string) error {
	args := []string{"suspend"}
	if id != "" {
		args = append(args, id)
	}
	_, _, err := d.vagrantCmd(args...)
	return err
}

// Calls "vagrant destroy"
func (d *Vagrant_2_2_Driver) Destroy(id string) error {
	args := []string{"destroy", "-f"}
	if id != "" {
		args = append(args, id)
	}
	_, _, err := d.vagrantCmd(args...)
	return err
}

// Calls "vagrant package"
func (d *Vagrant_2_2_Driver) Package(args []string) error {
	// Ideally we'd pass vagrantCWD into the package command but
	// we have to change directory into the vagrant cwd instead in order to
	// work around an upstream bug with the vagrant-libvirt plugin.
	// We can stop doing this when
	// https://github.com/vagrant-libvirt/vagrant-libvirt/issues/765
	// is fixed.
	oldDir, _ := os.Getwd()
	os.Chdir(d.VagrantCWD)
	defer os.Chdir(oldDir)
	args = append(args, "--output", "package.box")
	_, _, err := d.vagrantCmd(append([]string{"package"}, args...)...)
	return err
}

// Verify makes sure that Vagrant exists at the given path
func (d *Vagrant_2_2_Driver) Verify() error {
	vagrantPath, err := exec.LookPath(d.vagrantBinary)
	if err != nil {
		return fmt.Errorf("Can't find Vagrant binary!")
	}
	_, err = os.Stat(vagrantPath)
	if err != nil {
		return fmt.Errorf("Can't find Vagrant binary.")
	}

	constraints, err := version.NewConstraint(VAGRANT_MIN_VERSION)
	if err != nil {
		return fmt.Errorf("error parsing vagrant minimum version: %v", err)
	}
	vers, err := d.Version()
	if err != nil {
		return fmt.Errorf("error getting virtualbox version: %v", err)
	}
	v, err := version.NewVersion(vers)
	if err != nil {
		return fmt.Errorf("Error figuring out Vagrant version.")
	}

	if !constraints.Check(v) {
		return fmt.Errorf("installed Vagrant version must be >=2.0.2")
	}

	return nil
}

type VagrantSSHConfig struct {
	Hostname               string
	User                   string
	Port                   string
	UserKnownHostsFile     string
	StrictHostKeyChecking  bool
	PasswordAuthentication bool
	IdentityFile           string
	IdentitiesOnly         bool
	LogLevel               string
}

func parseSSHConfig(lines []string, value string) string {
	out := ""
	for _, line := range lines {
		if index := strings.Index(line, value); index != -1 {
			out = line[index+len(value):]
		}
	}
	return strings.Trim(out, "\r\n")
}

func yesno(yn string) bool {
	if yn == "no" {
		return false
	}
	return true
}

func (d *Vagrant_2_2_Driver) SSHConfig(id string) (*VagrantSSHConfig, error) {
	// vagrant ssh-config --host 8df7860
	args := []string{"ssh-config"}
	if id != "" {
		args = append(args, id)
	}
	stdout, _, err := d.vagrantCmd(args...)
	sshConf := &VagrantSSHConfig{}

	lines := strings.Split(stdout, "\n")
	sshConf.Hostname = parseSSHConfig(lines, "HostName ")
	sshConf.User = parseSSHConfig(lines, "User ")
	sshConf.Port = parseSSHConfig(lines, "Port ")
	sshConf.UserKnownHostsFile = parseSSHConfig(lines, "UserKnownHostsFile ")
	sshConf.IdentityFile = parseSSHConfig(lines, "IdentityFile ")
	sshConf.LogLevel = parseSSHConfig(lines, "LogLevel ")

	// handle the booleans
	sshConf.StrictHostKeyChecking = yesno(parseSSHConfig(lines, "StrictHostKeyChecking "))
	sshConf.PasswordAuthentication = yesno(parseSSHConfig(lines, "PasswordAuthentication "))
	sshConf.IdentitiesOnly = yesno((parseSSHConfig(lines, "IdentitiesOnly ")))

	return sshConf, err
}

// Version reads the version of VirtualBox that is installed.
func (d *Vagrant_2_2_Driver) Version() (string, error) {
	stdoutString, _, err := d.vagrantCmd([]string{"--version"}...)
	// Example stdout:

	// 	Installed Version: 2.2.3
	//
	// Vagrant was unable to check for the latest version of Vagrant.
	// Please check manually at https://www.vagrantup.com

	// Use regex to find version
	reg := regexp.MustCompile(`(\d+\.)?(\d+\.)?(\*|\d+)`)
	version := reg.FindString(stdoutString)
	if version == "" {
		return "", err
	}

	return version, nil
}

func (d *Vagrant_2_2_Driver) vagrantCmd(args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Calling Vagrant CLI: %#v", args)
	cmd := exec.Command(d.vagrantBinary, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAGRANT_CWD=%s", d.VagrantCWD))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return "", "", fmt.Errorf("Error starting vagrant command with args: %q",
			strings.Join(args, " "))
	}

	stdoutString := ""
	stderrString := ""

	donech := make(chan bool)

	go func() {
		for {
			select {
			case <-donech:
				line := stdout.String()
				if line != "" {
					log.Printf("[vagrant driver] stdout: %s", line)
				}
				stdoutString += line
				line = stderr.String()
				if line != "" {
					log.Printf("[vagrant driver] stderr: %s", line)
				}
				stderrString += line
				break
			default:
				line, _ := stdout.ReadString('\n')
				if line != "" {
					log.Printf("[vagrant driver] stdout: %s", line)
					stdoutString += line

				}
				line, _ = stderr.ReadString('\n')
				if line != "" {
					log.Printf("[vagrant driver] stderr: %s", line)
					stderrString += line

				}
				time.Sleep(500)
			}
		}
	}()

	err = cmd.Wait()

	donech <- true

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Vagrant error: %s", stderrString)
	}

	return stdoutString, stderrString, err
}
