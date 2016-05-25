package iso

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	gossh "golang.org/x/crypto/ssh"
)

// ESX5 driver talks to an ESXi5 hypervisor remotely over SSH to build
// virtual machines. This driver can only manage one machine at a time.
type ESX5Driver struct {
	Host           string
	Port           uint
	Username       string
	Password       string
	PrivateKey     string
	Datastore      string
	CacheDatastore string
	CacheDirectory string

	comm      packer.Communicator
	outputDir string
	vmId      string
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

func (d *ESX5Driver) IsRunning(string) (bool, error) {
	state, err := d.run(nil, "vim-cmd", "vmsvc/power.getstate", d.vmId)
	if err != nil {
		return false, err
	}
	return strings.Contains(state, "Powered on"), nil
}

func (d *ESX5Driver) ReloadVM() error {
	return d.sh("vim-cmd", "vmsvc/reload", d.vmId)
}

func (d *ESX5Driver) Start(vmxPathLocal string, headless bool) error {
	for i := 0; i < 20; i++ {
		err := d.sh("vim-cmd", "vmsvc/power.on", d.vmId)
		if err != nil {
			return err
		}
		time.Sleep((time.Duration(i) * time.Second) + 1)
		running, err := d.IsRunning(vmxPathLocal)
		if err != nil {
			return err
		}
		if running {
			return nil
		}
	}
	return errors.New("Retry limit exceeded")
}

func (d *ESX5Driver) Stop(vmxPathLocal string) error {
	return d.sh("vim-cmd", "vmsvc/power.off", d.vmId)
}

func (d *ESX5Driver) Register(vmxPathLocal string) error {
	vmxPath := filepath.ToSlash(filepath.Join(d.outputDir, filepath.Base(vmxPathLocal)))
	if err := d.upload(vmxPath, vmxPathLocal); err != nil {
		return err
	}
	r, err := d.run(nil, "vim-cmd", "solo/registervm", vmxPath)
	if err != nil {
		return err
	}
	d.vmId = strings.TrimRight(r, "\n")
	return nil
}

func (d *ESX5Driver) SuppressMessages(vmxPath string) error {
	return nil
}

func (d *ESX5Driver) Unregister(vmxPathLocal string) error {
	return d.sh("vim-cmd", "vmsvc/unregister", d.vmId)
}

func (d *ESX5Driver) Destroy() error {
	return d.sh("vim-cmd", "vmsvc/destroy", d.vmId)
}

func (d *ESX5Driver) IsDestroyed() (bool, error) {
	err := d.sh("test", "!", "-e", d.outputDir)
	if err != nil {
		return false, err
	}
	return true, err
}

func (d *ESX5Driver) UploadISO(localPath string, checksum string, checksumType string) (string, error) {
	finalPath := d.cachePath(localPath)
	if err := d.mkdir(filepath.ToSlash(filepath.Dir(finalPath))); err != nil {
		return "", err
	}

	log.Printf("Verifying checksum of %s", finalPath)
	if d.verifyChecksum(checksumType, checksum, finalPath) {
		log.Println("Initial checksum matched, no upload needed.")
		return finalPath, nil
	}

	if err := d.upload(finalPath, localPath); err != nil {
		return "", err
	}

	return finalPath, nil
}

func (d *ESX5Driver) ToolsIsoPath(string) string {
	return ""
}

func (d *ESX5Driver) ToolsInstall() error {
	return d.sh("vim-cmd", "vmsvc/tools.install", d.vmId)
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
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", d.Host, d.Port))
	defer conn.Close()
	if err != nil {
		return "", err
	}

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	return host, err
}

func (d *ESX5Driver) VNCAddress(vncBindIP string, portMin, portMax uint) (string, uint, error) {
	var vncPort uint

	//Process ports ESXi is listening on to determine which are available
	//This process does best effort to detect ports that are unavailable,
	//it will ignore any ports listened to by only localhost
	r, err := d.esxcli("network", "ip", "connection", "list")
	if err != nil {
		err = fmt.Errorf("Could not retrieve network information for ESXi: %v", err)
		return "", 0, err
	}

	listenPorts := make(map[string]bool)
	for record, err := r.read(); record != nil && err == nil; record, err = r.read() {
		if record["State"] == "LISTEN" {
			splitAddress := strings.Split(record["LocalAddress"], ":")
			if splitAddress[0] != "127.0.0.1" {
				port := splitAddress[len(splitAddress)-1]
				log.Printf("ESXi listening on address %s, port %s unavailable for VNC", record["LocalAddress"], port)
				listenPorts[port] = true
			}
		}
	}

	for port := portMin; port <= portMax; port++ {
		if _, ok := listenPorts[fmt.Sprintf("%d", port)]; ok {
			log.Printf("Port %d in use", port)
			continue
		}
		address := fmt.Sprintf("%s:%d", d.Host, port)
		log.Printf("Trying address: %s...", address)
		l, err := net.DialTimeout("tcp", address, 1*time.Second)

		if err != nil {
			if e, ok := err.(*net.OpError); ok {
				if e.Timeout() {
					log.Printf("Timeout connecting to: %s (check firewall rules)", address)
				} else {
					vncPort = port
					break
				}
			}
		} else {
			defer l.Close()
		}
	}

	if vncPort == 0 {
		err := fmt.Errorf("Unable to find available VNC port between %d and %d",
			portMin, portMax)
		return d.Host, vncPort, err
	}

	return d.Host, vncPort, nil
}

func (d *ESX5Driver) CommHost(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*Config)

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

	address := record["IPAddress"]
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

		files = append(files, filepath.ToSlash(filepath.Join(d.outputDir, string(line))))
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
	dirPath := filepath.Dir(path)
	return filepath.ToSlash(filepath.Join("/vmfs/volumes", d.Datastore, dirPath, filepath.Base(path)))
}

func (d *ESX5Driver) cachePath(path string) string {
	return filepath.ToSlash(filepath.Join("/vmfs/volumes", d.CacheDatastore, d.CacheDirectory, filepath.Base(path)))
}

func (d *ESX5Driver) connect() error {
	address := fmt.Sprintf("%s:%d", d.Host, d.Port)

	auth := []gossh.AuthMethod{
		gossh.Password(d.Password),
		gossh.KeyboardInteractive(
			ssh.PasswordKeyboardInteractive(d.Password)),
	}

	if d.PrivateKey != "" {
		signer, err := commonssh.FileSigner(d.PrivateKey)
		if err != nil {
			return err
		}

		auth = append(auth, gossh.PublicKeys(signer))
	}

	sshConfig := &ssh.Config{
		Connection: ssh.ConnectFunc("tcp", address),
		SSHConfig: &gossh.ClientConfig{
			User: d.Username,
			Auth: auth,
		},
	}

	comm, err := ssh.New(address, sshConfig)
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
	return d.comm.Upload(dst, f, nil)
}

func (d *ESX5Driver) verifyChecksum(ctype string, hash string, file string) bool {
	if ctype == "none" {
		if err := d.sh("stat", file); err != nil {
			return false
		}
	} else {
		stdin := bytes.NewBufferString(fmt.Sprintf("%s  %s", hash, file))
		_, err := d.run(stdin, fmt.Sprintf("%ssum", ctype), "-c")
		if err != nil {
			return false
		}
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
