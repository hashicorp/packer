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
	commonhelper "github.com/hashicorp/packer/helper/common"
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
	GalaxyFile           string   `mapstructure:"galaxy_file"`
	GalaxyCommand        string   `mapstructure:"galaxy_command"`
	GalaxyForceInstall   bool     `mapstructure:"galaxy_force_install"`
	RolesPath            string   `mapstructure:"roles_path"`
}

type Provisioner struct {
	config            Config
	adapter           *adapter.Adapter
	done              chan struct{}
	ansibleVersion    string
	ansibleMajVersion uint
}

type PassthroughTemplate struct {
	WinRMPassword string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	p.done = make(chan struct{})

	// Create passthrough for winrm password so we can fill it in once we know
	// it
	p.config.ctx.Data = &PassthroughTemplate{
		WinRMPassword: `{{.WinRMPassword}}`,
	}

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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Ansible...")
	// Interpolate env vars to check for .WinRMPassword
	p.config.ctx.Data = &PassthroughTemplate{
		WinRMPassword: getWinRMPassword(p.config.PackerBuildName),
	}
	for i, envVar := range p.config.AnsibleEnvVars {
		envVar, err := interpolate.Render(envVar, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.AnsibleEnvVars[i] = envVar
	}
	// Interpolate extra vars to check for .WinRMPassword
	for i, arg := range p.config.ExtraArguments {
		arg, err := interpolate.Render(arg, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.ExtraArguments[i] = arg
	}

	k, err := newUserKey(p.config.SSHAuthorizedKeyFile)
	if err != nil {
		return err
	}

	hostSigner, err := newSigner(p.config.SSHHostKeyFile)
	if err != nil {
		return fmt.Errorf("error creating host signer: %s", err)
	}

	// Remove the private key file
	if len(k.privKeyFile) > 0 {
		defer os.Remove(k.privKeyFile)
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
		return err
	}

	ui = &packer.SafeUi{
		Sem: make(chan int, 1),
		Ui:  ui,
	}
	p.adapter = adapter.NewAdapter(p.done, localListener, config, p.config.SFTPCmd, ui, comm)

	defer func() {
		log.Print("shutting down the SSH proxy")
		close(p.done)
		p.adapter.Shutdown()
	}()

	go p.adapter.Serve()

	if len(p.config.InventoryFile) == 0 {
		tf, err := ioutil.TempFile(p.config.InventoryDirectory, "packer-provisioner-ansible")
		if err != nil {
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		defer os.Remove(tf.Name())

		host := fmt.Sprintf("%s ansible_host=127.0.0.1 ansible_user=%s ansible_port=%d\n",
			p.config.HostAlias, p.config.User, p.config.LocalPort)
		if p.ansibleMajVersion < 2 {
			host = fmt.Sprintf("%s ansible_ssh_host=127.0.0.1 ansible_ssh_user=%s ansible_ssh_port=%d\n",
				p.config.HostAlias, p.config.User, p.config.LocalPort)
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
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		tf.Close()
		p.config.InventoryFile = tf.Name()
		defer func() {
			p.config.InventoryFile = ""
		}()
	}

	if err := p.executeAnsible(ui, comm, k.privKeyFile); err != nil {
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

func (p *Provisioner) executeAnsible(ui packer.Ui, comm packer.Communicator, privKeyFile string) error {
	playbook, _ := filepath.Abs(p.config.PlaybookFile)
	inventory := p.config.InventoryFile

	var envvars []string

	// Fetch external dependencies
	if len(p.config.GalaxyFile) > 0 {
		if err := p.executeGalaxy(ui, comm); err != nil {
			return fmt.Errorf("Error executing Ansible Galaxy: %s", err)
		}
	}
	args := []string{"--extra-vars", fmt.Sprintf("packer_build_name=%s packer_builder_type=%s -o IdentitiesOnly=yes",
		p.config.PackerBuildName, p.config.PackerBuilderType),
		"-i", inventory, playbook}
	if len(privKeyFile) > 0 {
		// Changed this from using --private-key to supplying -e ansible_ssh_private_key_file as the latter
		// is treated as a highest priority variable, and thus prevents overriding by dynamic variables
		// as seen in #5852
		// args = append(args, "--private-key", privKeyFile)
		args = append(args, "-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", privKeyFile))
	}

	// expose packer_http_addr extra variable
	httpAddr := common.GetHTTPAddr()
	if httpAddr != "" {
		args = append(args, "--extra-vars", fmt.Sprintf("packer_http_addr=%s", httpAddr))
	}

	args = append(args, p.config.ExtraArguments...)
	if len(p.config.AnsibleEnvVars) > 0 {
		envvars = append(envvars, p.config.AnsibleEnvVars...)
	}

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
	if len(getWinRMPassword(p.config.PackerBuildName)) > 0 {
		sanitized = strings.Replace(sanitized,
			getWinRMPassword(p.config.PackerBuildName), "*****", -1)
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

func getWinRMPassword(buildName string) string {
	winRMPass, _ := commonhelper.RetrieveSharedState("winrm_password", buildName)
	packer.LogSecretFilter.Set(winRMPass)
	return winRMPass
}
