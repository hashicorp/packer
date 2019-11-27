//go:generate mapstructure-to-hcl2 -type Config

package inspec

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
	"github.com/hashicorp/packer/template/interpolate"
)

var SupportedBackends = map[string]bool{"docker": true, "local": true, "ssh": true, "winrm": true}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// The command to run inspec
	Command    string
	SubCommand string

	// Extra options to pass to the inspec command
	ExtraArguments []string `mapstructure:"extra_arguments"`
	InspecEnvVars  []string `mapstructure:"inspec_env_vars"`

	// The profile to execute.
	Profile              string   `mapstructure:"profile"`
	AttributesDirectory  string   `mapstructure:"attributes_directory"`
	AttributesFiles      []string `mapstructure:"attributes"`
	Backend              string   `mapstructure:"backend"`
	User                 string   `mapstructure:"user"`
	Host                 string   `mapstructure:"host"`
	LocalPort            int      `mapstructure:"local_port"`
	SSHHostKeyFile       string   `mapstructure:"ssh_host_key_file"`
	SSHAuthorizedKeyFile string   `mapstructure:"ssh_authorized_key_file"`
}

type Provisioner struct {
	config           Config
	adapter          *adapter.Adapter
	done             chan struct{}
	inspecVersion    string
	inspecMajVersion uint
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
		p.config.Command = "inspec"
	}

	if p.config.SubCommand == "" {
		p.config.SubCommand = "exec"
	}

	var errs *packer.MultiError
	err = validateProfileConfig(p.config.Profile)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
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
	}

	if p.config.Backend == "" {
		p.config.Backend = "ssh"
	}

	if _, ok := SupportedBackends[p.config.Backend]; !ok {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("backend: %s must be a valid backend", p.config.Backend))
	}

	if p.config.Backend == "docker" && p.config.Host == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("backend: host must be specified for docker backend"))
	}

	if p.config.Host == "" {
		p.config.Host = "127.0.0.1"
	}

	if p.config.LocalPort > 65535 {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("local_port: %d must be a valid port", p.config.LocalPort))
	}

	if len(p.config.AttributesDirectory) > 0 {
		err = validateDirectoryConfig(p.config.AttributesDirectory, "attrs")
		if err != nil {
			log.Println(p.config.AttributesDirectory, "does not exist")
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
	out, err := exec.Command(p.config.Command, "version").Output()
	if err != nil {
		return fmt.Errorf(
			"Error running \"%s version\": %s", p.config.Command, err.Error())
	}

	versionRe := regexp.MustCompile(`\w (\d+\.\d+[.\d+]*)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return fmt.Errorf(
			"Could not find %s version in output:\n%s", p.config.Command, string(out))
	}

	version := matches[1]
	log.Printf("%s version: %s", p.config.Command, version)
	p.inspecVersion = version

	majVer, err := strconv.ParseUint(strings.Split(version, ".")[0], 10, 0)
	if err != nil {
		return fmt.Errorf("Could not parse major version from \"%s\".", version)
	}
	p.inspecMajVersion = uint(majVer)

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Inspec...")

	for i, envVar := range p.config.InspecEnvVars {
		envVar, err := interpolate.Render(envVar, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate inspec env vars: %s", err)
		}
		p.config.InspecEnvVars[i] = envVar
	}

	for i, arg := range p.config.ExtraArguments {
		arg, err := interpolate.Render(arg, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate inspec extra arguments: %s", err)
		}
		p.config.ExtraArguments[i] = arg
	}

	for i, arg := range p.config.AttributesFiles {
		arg, err := interpolate.Render(arg, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate inspec attributes: %s", err)
		}
		p.config.AttributesFiles[i] = arg
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
	p.adapter = adapter.NewAdapter(p.done, localListener, config, "", ui, comm)

	defer func() {
		log.Print("shutting down the SSH proxy")
		close(p.done)
		p.adapter.Shutdown()
	}()

	go p.adapter.Serve()

	tf, err := ioutil.TempFile(p.config.AttributesDirectory, "packer-provisioner-inspec.*.yml")
	if err != nil {
		return fmt.Errorf("Error preparing packer attributes file: %s", err)
	}
	defer os.Remove(tf.Name())

	w := bufio.NewWriter(tf)
	w.WriteString(fmt.Sprintf("packer_build_name: %s\n", p.config.PackerBuildName))
	w.WriteString(fmt.Sprintf("packer_builder_type: %s\n", p.config.PackerBuilderType))

	if err := w.Flush(); err != nil {
		tf.Close()
		return fmt.Errorf("Error preparing packer attributes file: %s", err)
	}
	tf.Close()
	p.config.AttributesFiles = append(p.config.AttributesFiles, tf.Name())

	if err := p.executeInspec(ui, comm, k.privKeyFile); err != nil {
		return fmt.Errorf("Error executing Inspec: %s", err)
	}

	return nil
}

func (p *Provisioner) executeInspec(ui packer.Ui, comm packer.Communicator, privKeyFile string) error {
	var envvars []string

	args := []string{p.config.SubCommand, p.config.Profile}
	args = append(args, "--backend", p.config.Backend)
	args = append(args, "--host", p.config.Host)

	if p.config.Backend == "ssh" {
		if len(privKeyFile) > 0 {
			args = append(args, "--key-files", privKeyFile)
		}
		args = append(args, "--user", p.config.User)
		args = append(args, "--port", strconv.Itoa(p.config.LocalPort))
	}

	args = append(args, "--input-file")
	args = append(args, p.config.AttributesFiles...)
	args = append(args, p.config.ExtraArguments...)

	if len(p.config.InspecEnvVars) > 0 {
		envvars = append(envvars, p.config.InspecEnvVars...)
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

	ui.Say(fmt.Sprintf("Executing Inspec: %s", strings.Join(cmd.Args, " ")))
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

func validateProfileConfig(name string) error {
	if name == "" {
		return fmt.Errorf("profile must be specified.")
	}
	return nil
}

func validateDirectoryConfig(name string, config string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, name, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s: %s must point to a directory", config, name)
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

	// To support Inspec calling back to us we need to write
	// this file down
	privateKeyDer := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	tf, err := ioutil.TempFile("", "packer-provisioner-inspec.*.key")
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
