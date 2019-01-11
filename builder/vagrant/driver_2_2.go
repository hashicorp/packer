package vagrant

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Vagrant_2_2_Driver struct {
	vagrantBinary string
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
func (d *Vagrant_2_2_Driver) Halt() error {
	_, _, err := d.vagrantCmd([]string{"halt"}...)
	return err
}

// Calls "vagrant suspend"
func (d *Vagrant_2_2_Driver) Suspend() error {
	_, _, err := d.vagrantCmd([]string{"suspend"}...)
	return err
}

// Calls "vagrant destroy"
func (d *Vagrant_2_2_Driver) Destroy() error {
	_, _, err := d.vagrantCmd([]string{"destroy", "-f"}...)
	return err
}

// Calls "vagrant package"
func (d *Vagrant_2_2_Driver) Package(args []string) error {
	_, _, err := d.vagrantCmd([]string{"package"}...)
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
	return out
}

func yesno(yn string) bool {
	if yn == "no" {
		return false
	}
	return true
}

func (d *Vagrant_2_2_Driver) SSHConfig() (*VagrantSSHConfig, error) {
	// vagrant ssh-config --host 8df7860
	stdout, _, err := d.vagrantCmd([]string{"ssh-config"}...)
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
	stdoutString, _, err := d.vagrantCmd([]string{"version"}...)
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
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Vagrant error: %s", stderrString)
	}

	log.Printf("[vagrant driver] stdout: %s", stdoutString)
	log.Printf("[vagrant driver] stderr: %s", stderrString)

	return stdoutString, stderrString, err
}
