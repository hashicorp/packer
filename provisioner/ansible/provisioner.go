//go:generate mapstructure-to-hcl2 -type Config

package ansible

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/crypto/ssh"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/adapter"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// The command to run ansible
	Command string

	// Extra options to pass to the ansible command
	ExtraArguments []string `mapstructure:"extra_arguments"`

	AnsibleEnvVars []string `mapstructure:"ansible_env_vars"`

	// The main playbook file to execute.
	PlaybookFile         string   `mapstructure:"playbook_file"`
	Groups               []string `mapstructure:"groups"`
	EmptyGroups          []string `mapstructure:"empty_groups"`
	HostAlias            string   `mapstructure:"host_alias"`
	User                 string   `mapstructure:"user"`
	LocalPort            int      `mapstructure:"local_port"`
	SSHHostKeyFile       string   `mapstructure:"ssh_host_key_file"`
	SSHAuthorizedKeyFile string   `mapstructure:"ssh_authorized_key_file"`
	SFTPCmd              string   `mapstructure:"sftp_command"`
	SkipVersionCheck     bool     `mapstructure:"skip_version_check"`
	UseSFTP              bool     `mapstructure:"use_sftp"`
	InventoryDirectory   string   `mapstructure:"inventory_directory"`
	InventoryFile        string   `mapstructure:"inventory_file"`
	KeepInventoryFile    bool     `mapstructure:"keep_inventory_file"`
	GalaxyFile           string   `mapstructure:"galaxy_file"`
	GalaxyCommand        string   `mapstructure:"galaxy_command"`
	GalaxyForceInstall   bool     `mapstructure:"galaxy_force_install"`
	RolesPath            string   `mapstructure:"roles_path"`
	//TODO: change default to false in v1.6.0.
	UseProxy     config.Trilean `mapstructure:"use_proxy"`
	userWasEmpty bool
}

type Provisioner struct {
	config            Config
	adapter           *adapter.Adapter
	done              chan struct{}
	ansibleVersion    string
	ansibleMajVersion uint
	generatedData     map[string]interface{}

	setupAdapterFunc   func(ui packer.Ui, comm packer.Communicator) (string, error)
	executeAnsibleFunc func(ui packer.Ui, comm packer.Communicator, privKeyFile string) error
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	p.done = make(chan struct{})

	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	if p.config.Command == "" {
		p.config.Command = "ansible-playbook"
	}

	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	if p.config.HostAlias == "" {
		p.config.HostAlias = "default"
	}

	var errs *packer.MultiError
	err = validateFileConfig(p.config.PlaybookFile, "playbook_file", true)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	// Check that the galaxy file exists, if configured
	if len(p.config.GalaxyFile) > 0 {
		err = validateFileConfig(p.config.GalaxyFile, "galaxy_file", true)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	// Check that the authorized key file exists
	if len(p.config.SSHAuthorizedKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHAuthorizedKeyFile, "ssh_authorized_key_file", true)
		if err != nil {
			log.Println(p.config.SSHAuthorizedKeyFile, "does not exist")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}
	if len(p.config.SSHHostKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHHostKeyFile, "ssh_host_key_file", true)
		if err != nil {
			log.Println(p.config.SSHHostKeyFile, "does not exist")
			errs = packer.MultiErrorAppend(errs, err)
		}
	} else {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_HOST_KEY_CHECKING=False")
	}

	if !p.config.UseSFTP {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_SCP_IF_SSH=True")
	}

	if p.config.LocalPort > 65535 {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("local_port: %d must be a valid port", p.config.LocalPort))
	}

	if len(p.config.InventoryDirectory) > 0 {
		err = validateInventoryDirectoryConfig(p.config.InventoryDirectory)
		if err != nil {
			log.Println(p.config.InventoryDirectory, "does not exist")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if !p.config.SkipVersionCheck {
		err = p.getVersion()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if p.config.User == "" {
		p.config.userWasEmpty = true
		usr, err := user.Current()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		} else {
			p.config.User = usr.Username
		}
	}
	if p.config.User == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("user: could not determine current user from environment."))
	}

	// These fields exist so that we can replace the functions for testing
	// logic inside of the Provision func; in actual use, these don't ever
	// need to get set.
	if p.setupAdapterFunc == nil {
		p.setupAdapterFunc = p.setupAdapter
	}
	if p.executeAnsibleFunc == nil {
		p.executeAnsibleFunc = p.executeAnsible
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) getVersion() error {
	out, err := exec.Command(p.config.Command, "--version").Output()
	if err != nil {
		return fmt.Errorf(
			"Error running \"%s --version\": %s", p.config.Command, err.Error())
	}

	versionRe := regexp.MustCompile(`\w (\d+\.\d+[.\d+]*)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return fmt.Errorf(
			"Could not find %s version in output:\n%s", p.config.Command, string(out))
	}

	version := matches[1]
	log.Printf("%s version: %s", p.config.Command, version)
	p.ansibleVersion = version

	majVer, err := strconv.ParseUint(strings.Split(version, ".")[0], 10, 0)
	if err != nil {
		return fmt.Errorf("Could not parse major version from \"%s\".", version)
	}
	p.ansibleMajVersion = uint(majVer)

	return nil
}

func (p *Provisioner) setupAdapter(ui packer.Ui, comm packer.Communicator) (string, error) {
	ui.Message("Setting up proxy adapter for Ansible....")

	k, err := newUserKey(p.config.SSHAuthorizedKeyFile)
	if err != nil {
		return "", err
	}

	hostSigner, err := newSigner(p.config.SSHHostKeyFile)
	if err != nil {
		return "", fmt.Errorf("error creating host signer: %s", err)
	}

	keyChecker := ssh.CertChecker{
		UserKeyFallback: func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if user := conn.User(); user != p.config.User {
				return nil, errors.New(fmt.Sprintf("authentication failed: %s is not a valid user", user))
			}

			if !bytes.Equal(k.Marshal(), pubKey.Marshal()) {
				return nil, errors.New("authentication failed: unauthorized key")
			}

			return nil, nil
		},
		IsUserAuthority: func(k ssh.PublicKey) bool { return true },
	}

	config := &ssh.ServerConfig{
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			log.Printf("authentication attempt from %s to %s as %s using %s", conn.RemoteAddr(), conn.LocalAddr(), conn.User(), method)
		},
		PublicKeyCallback: keyChecker.Authenticate,
		//NoClientAuth:      true,
	}

	config.AddHostKey(hostSigner)

	localListener, err := func() (net.Listener, error) {

		port := p.config.LocalPort
		tries := 1
		if port != 0 {
			tries = 10
		}
		for i := 0; i < tries; i++ {
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			port++
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			_, portStr, err := net.SplitHostPort(l.Addr().String())
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			p.config.LocalPort, err = strconv.Atoi(portStr)
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			return l, nil
		}
		return nil, errors.New("Error setting up SSH proxy connection")
	}()

	if err != nil {
		return "", err
	}

	ui = &packer.SafeUi{
		Sem: make(chan int, 1),
		Ui:  ui,
	}
	p.adapter = adapter.NewAdapter(p.done, localListener, config, p.config.SFTPCmd, ui, comm)

	return k.privKeyFile, nil
}

const DefaultSSHInventoryFilev2 = "{{ .HostAlias }} ansible_host={{ .Host }} ansible_user={{ .User }} ansible_port={{ .Port }}\n"
const DefaultSSHInventoryFilev1 = "{{ .HostAlias }} ansible_ssh_host={{ .Host }} ansible_ssh_user={{ .User }} ansible_ssh_port={{ .Port }}\n"
const DefaultWinRMInventoryFilev2 = "{{ .HostAlias}} ansible_host={{ .Host }} ansible_connection=winrm ansible_winrm_transport=basic ansible_shell_type=powershell ansible_user={{ .User}} ansible_port={{ .Port }}\n"

func (p *Provisioner) createInventoryFile() error {
	log.Printf("Creating inventory file for Ansible run...")
	tf, err := ioutil.TempFile(p.config.InventoryDirectory, "packer-provisioner-ansible")
	if err != nil {
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}

	// figure out which inventory line template to use
	hostTemplate := DefaultSSHInventoryFilev2
	if p.ansibleMajVersion < 2 {
		hostTemplate = DefaultSSHInventoryFilev1
	}
	if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
		hostTemplate = DefaultWinRMInventoryFilev2
	}

	// interpolate template to generate host with necessary vars.
	ctxData := p.generatedData
	ctxData["HostAlias"] = p.config.HostAlias
	ctxData["User"] = p.config.User
	if !p.config.UseProxy.False() {
		ctxData["Host"] = "127.0.0.1"
		ctxData["Port"] = p.config.LocalPort
	}
	p.config.ctx.Data = ctxData

	host, err := interpolate.Render(hostTemplate, &p.config.ctx)

	if err != nil {
		return fmt.Errorf("Error generating inventory file from template: %s", err)
	}

	w := bufio.NewWriter(tf)
	w.WriteString(host)

	for _, group := range p.config.Groups {
		fmt.Fprintf(w, "[%s]\n%s", group, host)
	}

	for _, group := range p.config.EmptyGroups {
		fmt.Fprintf(w, "[%s]\n", group)
	}

	if err := w.Flush(); err != nil {
		tf.Close()
		os.Remove(tf.Name())
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}
	tf.Close()
	p.config.InventoryFile = tf.Name()

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible...")
	// Interpolate env vars to check for generated values like password and port
	p.generatedData = generatedData
	p.config.ctx.Data = generatedData
	for i, envVar := range p.config.AnsibleEnvVars {
		envVar, err := interpolate.Render(envVar, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.AnsibleEnvVars[i] = envVar
	}
	// Interpolate extra vars to check for generated values like password and port
	for i, arg := range p.config.ExtraArguments {
		arg, err := interpolate.Render(arg, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.ExtraArguments[i] = arg
	}

	// Set up proxy if host IP is missing or communicator type is wrong.
	if p.config.UseProxy.False() {
		hostIP := generatedData["Host"].(string)
		if hostIP == "" {
			ui.Error("Warning: use_proxy is false, but instance does" +
				" not have an IP address to give to Ansible. Falling back" +
				" to use localhost proxy.")
			p.config.UseProxy = config.TriTrue
		}
		connType := generatedData["ConnType"]
		if connType != "ssh" && connType != "winrm" {
			ui.Error("Warning: use_proxy is false, but communicator is " +
				"neither ssh nor winrm, so without the proxy ansible will not" +
				" function. Falling back to localhost proxy.")
			p.config.UseProxy = config.TriTrue
		}
	}

	privKeyFile := ""
	if !p.config.UseProxy.False() {
		// We set up the proxy if useProxy is either true or unset.
		pkf, err := p.setupAdapterFunc(ui, comm)
		if err != nil {
			return err
		}
		// This is necessary to avoid accidentally redeclaring
		// privKeyFile in the scope of this if statement.
		privKeyFile = pkf

		defer func() {
			log.Print("shutting down the SSH proxy")
			close(p.done)
			p.adapter.Shutdown()
		}()

		go p.adapter.Serve()

		// Remove the private key file
		if len(privKeyFile) > 0 {
			defer os.Remove(privKeyFile)
		}
	} else {
		connType := generatedData["ConnType"].(string)
		switch connType {
		case "ssh":
			ui.Message("Not using Proxy adapter for Ansible run:\n" +
				"\tUsing ssh keys from Packer communicator...")
			// In this situation, we need to make sure we have the
			// private key we actually use to access the instance.
			SSHPrivateKeyFile := generatedData["SSHPrivateKeyFile"].(string)
			if SSHPrivateKeyFile != "" {
				privKeyFile = SSHPrivateKeyFile
			} else {
				// See if we can get a private key and write that to a tmpfile
				SSHPrivateKey := generatedData["SSHPrivateKey"].(string)
				tmpSSHPrivateKey, err := tmp.File("ansible-key")
				if err != nil {
					return fmt.Errorf("Error writing private key to temp file for"+
						"ansible connection: %v", err)
				}
				_, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
				if err != nil {
					return errors.New("failed to write private key to temp file")
				}
				err = tmpSSHPrivateKey.Close()
				if err != nil {
					return errors.New("failed to close private key temp file")
				}
				privKeyFile = tmpSSHPrivateKey.Name()
			}

			// Also make sure that the username matches the SSH keys given.
			if p.config.userWasEmpty {
				p.config.User = generatedData["User"].(string)
			}
		case "winrm":
			ui.Message("Not using Proxy adapter for Ansible run:\n" +
				"\tUsing WinRM Password from Packer communicator...")
		}
	}

	if len(p.config.InventoryFile) == 0 {
		// Create the inventory file
		err := p.createInventoryFile()
		if err != nil {
			return err
		}
		if !p.config.KeepInventoryFile {
			// Delete the generated inventory file
			defer func() {
				os.Remove(p.config.InventoryFile)
				p.config.InventoryFile = ""
			}()
		}
	}

	if err := p.executeAnsibleFunc(ui, comm, privKeyFile); err != nil {
		return fmt.Errorf("Error executing Ansible: %s", err)
	}

	return nil
}

func (p *Provisioner) executeGalaxy(ui packer.Ui, comm packer.Communicator) error {
	galaxyFile := filepath.ToSlash(p.config.GalaxyFile)

	// ansible-galaxy install -r requirements.yml
	args := []string{"install", "-r", galaxyFile}
	// Add force to arguments
	if p.config.GalaxyForceInstall {
		args = append(args, "-f")
	}
	// Add roles_path argument if specified
	if p.config.RolesPath != "" {
		args = append(args, "-p", filepath.ToSlash(p.config.RolesPath))
	}

	ui.Message(fmt.Sprintf("Executing Ansible Galaxy"))
	cmd := exec.Command(p.config.GalaxyCommand, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				ui.Message(line)
			}
			if err != nil {
				if err == io.EOF {
					break
				} else {
					ui.Error(err.Error())
					break
				}
			}
		}
		wg.Done()
	}
	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Non-zero exit status: %s", err)
	}
	return nil
}

func (p *Provisioner) createCmdArgs(httpAddr, inventory, playbook, privKeyFile string) (args []string, envVars []string) {
	args = []string{}

	if p.config.PackerBuildName != "" {
		// HCL configs don't currently have the PakcerBuildName. Don't
		// cause weirdness with a half-set variable
		args = append(args, "-e", fmt.Sprintf("packer_build_name=%s", p.config.PackerBuildName))
	}

	args = append(args, "-e", fmt.Sprintf("packer_builder_type=%s", p.config.PackerBuilderType))
	if len(privKeyFile) > 0 {
		// "-e ansible_ssh_private_key_file" is preferable to "--private-key"
		// because it is a higher priority variable and therefore won't get
		// overridden by dynamic variables. See #5852 for more details.
		args = append(args, "-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", privKeyFile))
	}

	// expose packer_http_addr extra variable
	if httpAddr != "" {
		args = append(args, "-e", fmt.Sprintf("packer_http_addr=%s", httpAddr))
	}

	// Add password to ansible call.
	if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
		args = append(args, "-e", fmt.Sprintf("ansible_password=%s", p.generatedData["Password"]))
	}

	if p.generatedData["ConnType"] == "ssh" {
		// Add ssh extra args to set IdentitiesOnly
		args = append(args, "--ssh-extra-args", "-o IdentitiesOnly=yes")
	}

	args = append(args, p.config.ExtraArguments...)

	if len(p.config.AnsibleEnvVars) > 0 {
		envVars = append(envVars, p.config.AnsibleEnvVars...)
	}

	// This must be the last arg appended to args
	args = append(args, "-i", inventory, playbook)
	return args, envVars
}

func (p *Provisioner) executeAnsible(ui packer.Ui, comm packer.Communicator, privKeyFile string) error {
	playbook, _ := filepath.Abs(p.config.PlaybookFile)
	inventory := p.config.InventoryFile
	httpAddr := p.generatedData["PackerHTTPAddr"].(string)

	// Fetch external dependencies
	if len(p.config.GalaxyFile) > 0 {
		if err := p.executeGalaxy(ui, comm); err != nil {
			return fmt.Errorf("Error executing Ansible Galaxy: %s", err)
		}
	}

	args, envvars := p.createCmdArgs(httpAddr, inventory, playbook, privKeyFile)

	cmd := exec.Command(p.config.Command, args...)

	cmd.Env = os.Environ()
	if len(envvars) > 0 {
		cmd.Env = append(cmd.Env, envvars...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				ui.Message(line)
			}
			if err != nil {
				if err == io.EOF {
					break
				} else {
					ui.Error(err.Error())
					break
				}
			}
		}
		wg.Done()
	}
	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	// remove winrm password from command, if it's been added
	flattenedCmd := strings.Join(cmd.Args, " ")
	sanitized := flattenedCmd
	winRMPass, ok := p.generatedData["WinRMPassword"]
	if ok && winRMPass != "" {
		sanitized = strings.Replace(sanitized,
			winRMPass.(string), "*****", -1)
	}
	ui.Say(fmt.Sprintf("Executing Ansible: %s", sanitized))

	if err := cmd.Start(); err != nil {
		return err
	}
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Non-zero exit status: %s", err)
	}

	return nil
}

func validateFileConfig(name string, config string, req bool) error {
	if req {
		if name == "" {
			return fmt.Errorf("%s must be specified.", config)
		}
	}
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, name, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, name)
	}
	return nil
}

func validateInventoryDirectoryConfig(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("inventory_directory: %s is invalid: %s", name, err)
	} else if !info.IsDir() {
		return fmt.Errorf("inventory_directory: %s must point to a directory", name)
	}
	return nil
}

type userKey struct {
	ssh.PublicKey
	privKeyFile string
}

func newUserKey(pubKeyFile string) (*userKey, error) {
	userKey := new(userKey)
	if len(pubKeyFile) > 0 {
		pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
		if err != nil {
			return nil, errors.New("Failed to read public key")
		}
		userKey.PublicKey, _, _, _, err = ssh.ParseAuthorizedKey(pubKeyBytes)
		if err != nil {
			return nil, errors.New("Failed to parse authorized key")
		}

		return userKey, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("Failed to generate key pair")
	}
	userKey.PublicKey, err = ssh.NewPublicKey(key.Public())
	if err != nil {
		return nil, errors.New("Failed to extract public key from generated key pair")
	}

	// To support Ansible calling back to us we need to write
	// this file down
	privateKeyDer := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	tf, err := tmp.File("ansible-key")
	if err != nil {
		return nil, errors.New("failed to create temp file for generated key")
	}
	_, err = tf.Write(pem.EncodeToMemory(&privateKeyBlock))
	if err != nil {
		return nil, errors.New("failed to write private key to temp file")
	}

	err = tf.Close()
	if err != nil {
		return nil, errors.New("failed to close private key temp file")
	}
	userKey.privKeyFile = tf.Name()

	return userKey, nil
}

type signer struct {
	ssh.Signer
}

func newSigner(privKeyFile string) (*signer, error) {
	signer := new(signer)

	if len(privKeyFile) > 0 {
		privateBytes, err := ioutil.ReadFile(privKeyFile)
		if err != nil {
			return nil, errors.New("Failed to load private host key")
		}

		signer.Signer, err = ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			return nil, errors.New("Failed to parse private host key")
		}

		return signer, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("Failed to generate server key pair")
	}

	signer.Signer, err = ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, errors.New("Failed to extract private key from generated key pair")
	}

	return signer, nil
}
