package vmware

import (
	"bytes"
	gossh "code.google.com/p/go.crypto/ssh"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type ESX5Driver struct {
	comm   packer.Communicator
	config *config
}

func (d *ESX5Driver) CompactDisk(diskPathLocal string) error {
	return nil
}

func (d *ESX5Driver) CreateDisk(diskPathLocal string, size string, typeId string) error {
	diskPath := d.datastorePath(diskPathLocal)
	return d.sh("vmkfstools", "-c", size, "-d", typeId, "-a", "lsilogic", diskPath)
}

func (d *ESX5Driver) IsRunning(vmxPathLocal string) (bool, error) {
	vmxPath := d.datastorePath(vmxPathLocal)
	state, err := d.run(nil, "vim-cmd", "vmsvc/power.getstate", vmxPath)
	if err != nil {
		return false, err
	}
	return strings.Contains(state, "Powered on"), nil
}

func (d *ESX5Driver) Start(vmxPathLocal string, headless bool) error {
	return d.sh("vim-cmd", "vmsvc/power.on", d.datastorePath(vmxPathLocal))
}

func (d *ESX5Driver) Stop(vmxPathLocal string) error {
	return d.sh("vim-cmd", "vmsvc/power.off", d.datastorePath(vmxPathLocal))
}

func (d *ESX5Driver) Register(vmxPathLocal string) error {
	vmxPath := d.datastorePath(vmxPathLocal)
	if err := d.upload(vmxPathLocal, vmxPath); err != nil {
		return err
	}
	return d.sh("vim-cmd", "solo/registervm", vmxPath)
}

func (d *ESX5Driver) Unregister(vmxPathLocal string) error {
	return d.sh("vim-cmd", "vmsvc/unregister", d.datastorePath(vmxPathLocal))
}

func (d *ESX5Driver) ToolsIsoPath(string) string {
	return ""
}

func (d *ESX5Driver) DhcpLeasesPath(string) string {
	return ""
}

func (d *ESX5Driver) Verify() error {
	checks := []func() error{
		d.connect,
		d.checkSystemVersion,
		d.checkGuestIPHackEnabled,
		d.checkOutputFolder,
	}

	for _, check := range checks {
		err := check()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *ESX5Driver) HostIP() (string, error) {
	ip := net.ParseIP(d.config.RemoteHost)
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, dev := range interfaces {
		addrs, err := dev.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.Contains(ip) {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", errors.New("Unable to determine Host IP")
}

func (d *ESX5Driver) VNCAddress(portMin, portMax uint) (string, uint) {
	var vncPort uint
	// TODO(dougm) use esxcli network ip connection list
	for port := portMin; port <= portMax; port++ {
		address := fmt.Sprintf("%s:%d", d.config.RemoteHost, port)
		log.Printf("Trying address: %s...", address)
		l, err := net.DialTimeout("tcp", address, 1*time.Second)

		if err == nil {
			log.Printf("%s in use", address)
			l.Close()
		} else if e, ok := err.(*net.OpError); ok {
			if e.Err == syscall.ECONNREFUSED {
				// then port should be available for listening
				vncPort = port
				break
			} else if e.Timeout() {
				log.Printf("Timeout connecting to: %s (check firewall rules)", address)
			}
		}
	}

	return d.config.RemoteHost, vncPort
}

func (d *ESX5Driver) SSHAddress() func(multistep.StateBag) (string, error) {
	return d.sshAddress
}

func (d *ESX5Driver) Download() func(*common.DownloadConfig, multistep.StateBag) (string, error, bool) {
	return d.download
}

func (d *ESX5Driver) FileExists(path string) bool {
	err := d.sh("test", "-e", d.datastorePath(path))
	if err != nil {
		return false
	}
	return true
}

func (d *ESX5Driver) MkdirAll(path string) error {
	return d.sh("mkdir", "-p", d.datastorePath(path))
}

func (d *ESX5Driver) RemoveAll(path string) error {
	return d.sh("rm", "-rf", d.datastorePath(path))
}

func (d *ESX5Driver) DirType() string {
	return "datastore"
}

func (d *ESX5Driver) datastorePath(path string) string {
	return filepath.Join("/vmfs/volumes", d.config.RemoteDatastore, path)
}

func (d *ESX5Driver) connect() error {
	if d.config.RemoteHost == "" {
		return errors.New("A remote_host must be specified.")
	}
	if d.config.RemotePort == 0 {
		d.config.RemotePort = 22
	}
	address := fmt.Sprintf("%s:%d", d.config.RemoteHost, d.config.RemotePort)
	auth := []gossh.ClientAuth{
		gossh.ClientAuthPassword(ssh.Password(d.config.RemotePassword)),
		gossh.ClientAuthKeyboardInteractive(
			ssh.PasswordKeyboardInteractive(d.config.RemotePassword)),
	}
	// TODO(dougm) KeyPath support
	sshConfig := &ssh.Config{
		Connection: ssh.ConnectFunc("tcp", address),
		SSHConfig: &gossh.ClientConfig{
			User: d.config.RemoteUser,
			Auth: auth,
		},
		NoPty: true,
	}

	comm, err := ssh.New(sshConfig)
	if err != nil {
		return err
	}

	d.comm = comm
	return nil
}

func (d *ESX5Driver) checkSystemVersion() error {
	r, err := d.esxcli("system", "version", "get")

	if err != nil {
		return err
	}

	record, err := r.read()
	if err != nil {
		return err
	}

	log.Printf("Connected to %s %s %s", record["Product"], record["Version"], record["Build"])

	return nil
}

func (d *ESX5Driver) checkGuestIPHackEnabled() error {
	r, err := d.esxcli("system", "settings", "advanced", "list", "-o", "/Net/GuestIPHack")

	if err != nil {
		return err
	}

	record, err := r.read()

	if err != nil {
		return err
	}

	if record["IntValue"] != "1" {
		return errors.New("GuestIPHack is required, enable with:\n" +
			"esxcli system settings advanced set -o /Net/GuestIPHack -i 1")
	}

	return nil
}

func (d *ESX5Driver) checkOutputFolder() error {
	if d.config.RemoteDatastore == "" {
		d.config.RemoteDatastore = "datastore1"
	}
	if !d.config.PackerForce && d.FileExists(d.config.OutputDir) {
		return fmt.Errorf("Output folder '%s' already exists. It must not exist.",
			d.config.OutputDir)
	}
	return nil
}

func (d *ESX5Driver) download(config *common.DownloadConfig, state multistep.StateBag) (string, error, bool) {
	cacheRoot, _ := filepath.Abs(".")
	targetFile, err := filepath.Rel(cacheRoot, config.TargetPath)

	if err != nil {
		return "", err, false
	}

	path := d.datastorePath(targetFile)

	err = d.MkdirAll(filepath.Dir(targetFile))
	if err != nil {
		return "", err, false
	}

	if d.verifyChecksum(d.config.ISOChecksumType, d.config.ISOChecksum, path) {
		log.Println("Initial checksum matched, no download needed.")
		return path, nil, true
	}
	// TODO(dougm) progress and handle interrupt
	err = d.sh("wget", config.Url, "-O", path)

	return path, err, true
}

func (d *ESX5Driver) upload(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.comm.Upload(dst, f)
}

func (d *ESX5Driver) verifyChecksum(ctype string, hash string, file string) bool {
	stdin := bytes.NewBufferString(fmt.Sprintf("%s  %s", hash, file))
	_, err := d.run(stdin, fmt.Sprintf("%ssum", ctype), "-c")
	if err != nil {
		return false
	}
	return true
}

func (d *ESX5Driver) ssh(command string, stdin io.Reader) (*bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer

	cmd := &packer.RemoteCmd{
		Command: command,
		Stdout:  &stdout,
		Stderr:  &stderr,
		Stdin:   stdin,
	}

	err := d.comm.Start(cmd)
	if err != nil {
		return nil, err
	}

	cmd.Wait()

	if cmd.ExitStatus != 0 {
		err = fmt.Errorf("'%s'\n\nStdout: %s\n\nStderr: %s",
			cmd.Command, stdout.String(), stderr.String())
		return nil, err
	}

	return &stdout, nil
}

func (d *ESX5Driver) run(stdin io.Reader, args ...string) (string, error) {
	stdout, err := d.ssh(strings.Join(args, " "), stdin)
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func (d *ESX5Driver) sh(args ...string) error {
	_, err := d.run(nil, args...)
	return err
}

func (d *ESX5Driver) esxcli(args ...string) (*esxcliReader, error) {
	stdout, err := d.ssh("esxcli --formatter csv "+strings.Join(args, " "), nil)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(stdout.Bytes()))
	r.TrailingComma = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	return &esxcliReader{r, header}, nil
}

type esxcliReader struct {
	cr     *csv.Reader
	header []string
}

func (r *esxcliReader) read() (map[string]string, error) {
	fields, err := r.cr.Read()

	if err != nil {
		return nil, err
	}

	record := map[string]string{}
	for i, v := range fields {
		record[r.header[i]] = v
	}

	return record, nil
}

func (r *esxcliReader) find(key, val string) (map[string]string, error) {
	for {
		record, err := r.read()
		if err != nil {
			return nil, err
		}
		if record[key] == val {
			return record, nil
		}
	}
}

func (d *ESX5Driver) sshAddress(state multistep.StateBag) (string, error) {
	if address, ok := state.GetOk("vm_address"); ok {
		return address.(string), nil
	}

	r, err := d.esxcli("network", "vm", "list")
	if err != nil {
		return "", err
	}

	record, err := r.find("Name", d.config.VMName)
	if err != nil {
		return "", err
	}
	wid := record["WorldID"]
	if wid == "" {
		return "", errors.New("VM WorldID not found")
	}

	r, err = d.esxcli("network", "vm", "port", "list", "-w", wid)
	if err != nil {
		return "", err
	}

	record, err = r.read()
	if err != nil {
		return "", err
	}

	if record["IPAddress"] == "0.0.0.0" {
		return "", errors.New("VM network port found, but no IP address")
	}

	address := fmt.Sprintf("%s:%d", record["IPAddress"], d.config.SSHPort)
	state.Put("vm_address", address)
	return address, nil
}
