package iso

import (
	"bufio"
	"bytes"
	gossh "code.google.com/p/go.crypto/ssh"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
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

// ESX5 driver talks to an ESXi5 hypervisor remotely over SSH to build
// virtual machines. This driver can only manage one machine at a time.
type ESX5Driver struct {
	Host      string
	Port      uint
	Username  string
	Password  string
	Datastore string

	comm      packer.Communicator
	outputDir string
}

func (d *ESX5Driver) Clone(dst, src string) error {
	return errors.New("Cloning is not supported with the ESX driver.")
}

func (d *ESX5Driver) CompactDisk(diskPathLocal string) error {
	return nil
}

func (d *ESX5Driver) CreateDisk(diskPathLocal string, size string, typeId string) error {
	diskPath := d.datastorePath(diskPathLocal)
	return d.sh("vmkfstools", "-c", size, "-d", typeId, "-a", "lsilogic", diskPath)
}

func (d *ESX5Driver) IsRunning(vmxPathLocal string) (bool, error) {
	vmxPath := filepath.Join(d.outputDir, filepath.Base(vmxPathLocal))
	state, err := d.run(nil, "vim-cmd", "vmsvc/power.getstate", vmxPath)
	if err != nil {
		return false, err
	}
	return strings.Contains(state, "Powered on"), nil
}

func (d *ESX5Driver) Start(vmxPathLocal string, headless bool) error {
	vmxPath := filepath.Join(d.outputDir, filepath.Base(vmxPathLocal))
	return d.sh("vim-cmd", "vmsvc/power.on", vmxPath)
}

func (d *ESX5Driver) Stop(vmxPathLocal string) error {
	vmxPath := filepath.Join(d.outputDir, filepath.Base(vmxPathLocal))
	return d.sh("vim-cmd", "vmsvc/power.off", vmxPath)
}

func (d *ESX5Driver) Register(vmxPathLocal string) error {
	vmxPath := filepath.Join(d.outputDir, filepath.Base(vmxPathLocal))
	if err := d.upload(vmxPath, vmxPathLocal); err != nil {
		return err
	}
	return d.sh("vim-cmd", "solo/registervm", vmxPath)
}

func (d *ESX5Driver) SuppressMessages(vmxPath string) error {
	return nil
}

func (d *ESX5Driver) Unregister(vmxPathLocal string) error {
	vmxPath := filepath.Join(d.outputDir, filepath.Base(vmxPathLocal))
	return d.sh("vim-cmd", "vmsvc/unregister", vmxPath)
}

func (d *ESX5Driver) UploadISO(localPath string) (string, error) {
	cacheRoot, _ := filepath.Abs(".")
	targetFile, err := filepath.Rel(cacheRoot, localPath)
	if err != nil {
		return "", err
	}

	finalPath := d.datastorePath(targetFile)
	if err := d.mkdir(filepath.Dir(finalPath)); err != nil {
		return "", err
	}

	if err := d.upload(finalPath, localPath); err != nil {
		return "", err
	}

	return finalPath, nil
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
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (d *ESX5Driver) HostIP() (string, error) {
	ip := net.ParseIP(d.Host)
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
		address := fmt.Sprintf("%s:%d", d.Host, port)
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

	return d.Host, vncPort
}

func (d *ESX5Driver) SSHAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*config)

	if address, ok := state.GetOk("vm_address"); ok {
		return address.(string), nil
	}

	r, err := d.esxcli("network", "vm", "list")
	if err != nil {
		return "", err
	}

	record, err := r.find("Name", config.VMName)
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

	address := fmt.Sprintf("%s:%d", record["IPAddress"], config.SSHPort)
	state.Put("vm_address", address)
	return address, nil
}

//-------------------------------------------------------------------
// OutputDir implementation
//-------------------------------------------------------------------

func (d *ESX5Driver) DirExists() (bool, error) {
	err := d.sh("test", "-e", d.outputDir)
	return err == nil, nil
}

func (d *ESX5Driver) ListFiles() ([]string, error) {
	stdout, err := d.ssh("ls -1p "+d.outputDir, nil)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, 10)
	reader := bufio.NewReader(stdout)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if line[len(line)-1] == '/' {
			continue
		}

		files = append(files, filepath.Join(d.outputDir, string(line)))
	}

	return files, nil
}

func (d *ESX5Driver) MkdirAll() error {
	return d.mkdir(d.outputDir)
}

func (d *ESX5Driver) Remove(path string) error {
	return d.sh("rm", path)
}

func (d *ESX5Driver) RemoveAll() error {
	return d.sh("rm", "-rf", d.outputDir)
}

func (d *ESX5Driver) SetOutputDir(path string) {
	d.outputDir = d.datastorePath(path)
}

func (d *ESX5Driver) String() string {
	return d.outputDir
}

func (d *ESX5Driver) datastorePath(path string) string {
	return filepath.Join("/vmfs/volumes", d.Datastore, path)
}

func (d *ESX5Driver) connect() error {
	address := fmt.Sprintf("%s:%d", d.Host, d.Port)

	auth := []gossh.ClientAuth{
		gossh.ClientAuthPassword(ssh.Password(d.Password)),
		gossh.ClientAuthKeyboardInteractive(
			ssh.PasswordKeyboardInteractive(d.Password)),
	}

	// TODO(dougm) KeyPath support
	sshConfig := &ssh.Config{
		Connection: ssh.ConnectFunc("tcp", address),
		SSHConfig: &gossh.ClientConfig{
			User: d.Username,
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

	log.Printf("Connected to %s %s %s", record["Product"],
		record["Version"], record["Build"])
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
		return errors.New(
			"GuestIPHack is required, enable by running this on the ESX machine:\n" +
				"esxcli system settings advanced set -o /Net/GuestIPHack -i 1")
	}

	return nil
}

func (d *ESX5Driver) mkdir(path string) error {
	return d.sh("mkdir", "-p", path)
}

func (d *ESX5Driver) upload(dst, src string) error {
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
