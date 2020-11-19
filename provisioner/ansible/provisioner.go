//go:generate mapstructure-to-hcl2 -type Config
//go:generate struct-markdown

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
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/adapter"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	// The command to invoke ansible. Defaults to
	//  `ansible-playbook`. If you would like to provide a more complex command,
	//  for example, something that sets up a virtual environment before calling
	//  ansible, take a look at the ansible wrapper guide below for inspiration.
	//  Please note that Packer expects Command to be a path to an executable.
	//  Arbitrary bash scripting will not work and needs to go inside an
	//  executable script.
	Command string `mapstructure:"command"`
	// Extra arguments to pass to Ansible.
	// These arguments _will not_ be passed through a shell and arguments should
	// not be quoted. Usage example:
	//
	// ```json
	//   "extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
	// ```
	//
	// If you are running a Windows build on AWS, Azure, Google Compute, or OpenStack
	// and would like to access the auto-generated password that Packer uses to
	// connect to a Windows instance via WinRM, you can use the template variable
	// `{{.WinRMPassword}}` in this option. For example:
	//
	// ```json
	//   "extra_arguments": [
	//     "--extra-vars", "winrm_password={{ .WinRMPassword }}"
	//   ]
	// ```
	ExtraArguments []string `mapstructure:"extra_arguments"`
	// Environment variables to set before
	//   running Ansible. Usage example:
	//
	//   ```json
	//     "ansible_env_vars": [ "ANSIBLE_HOST_KEY_CHECKING=False", "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes -o ControlMaster=auto -o ControlPersist=60s'", "ANSIBLE_NOCOLOR=True" ]
	//   ```
	//
	//   This is a [template engine](/docs/templates/engine). Therefore, you
	//   may use user variables and template functions in this field.
	//
	//   For example, if you are running a Windows build on AWS, Azure,
	//   Google Compute, or OpenStack and would like to access the auto-generated
	//   password that Packer uses to connect to a Windows instance via WinRM, you
	//   can use the template variable `{{.WinRMPassword}}` in this option. Example:
	//
	//   ```json
	//   "ansible_env_vars": [ "WINRM_PASSWORD={{.WinRMPassword}}" ],
	//   ```
	AnsibleEnvVars []string `mapstructure:"ansible_env_vars"`
	// The playbook to be run by Ansible.
	PlaybookFile string `mapstructure:"playbook_file" required:"true"`
	// Specifies --ssh-extra-args on command line defaults to -o IdentitiesOnly=yes
	AnsibleSSHExtraArgs []string `mapstructure:"ansible_ssh_extra_args"`
	// The groups into which the Ansible host should
	//  be placed. When unspecified, the host is not associated with any groups.
	Groups []string `mapstructure:"groups"`
	// The groups which should be present in
	//  inventory file but remain empty.
	EmptyGroups []string `mapstructure:"empty_groups"`
	//  The alias by which the Ansible host should be
	// known. Defaults to `default`. This setting is ignored when using a custom
	// inventory file.
	HostAlias string `mapstructure:"host_alias"`
	// The `ansible_user` to use. Defaults to the user running
	//  packer, NOT the user set for your communicator. If you want to use the same
	//  user as the communicator, you will need to manually set it again in this
	//  field.
	User string `mapstructure:"user"`
	// The port on which to attempt to listen for SSH
	//  connections. This value is a starting point. The provisioner will attempt
	//  listen for SSH connections on the first available of ten ports, starting at
	//  `local_port`. A system-chosen port is used when `local_port` is missing or
	//  empty.
	LocalPort int `mapstructure:"local_port"`
	// The SSH key that will be used to run the SSH
	//  server on the host machine to forward commands to the target machine.
	//  Ansible connects to this server and will validate the identity of the
	//  server using the system known_hosts. The default behavior is to generate
	//  and use a onetime key. Host key checking is disabled via the
	//  `ANSIBLE_HOST_KEY_CHECKING` environment variable if the key is generated.
	SSHHostKeyFile string `mapstructure:"ssh_host_key_file"`
	// The SSH public key of the Ansible
	//  `ssh_user`. The default behavior is to generate and use a onetime key. If
	//  this key is generated, the corresponding private key is passed to
	//  `ansible-playbook` with the `-e ansible_ssh_private_key_file` option.
	SSHAuthorizedKeyFile string `mapstructure:"ssh_authorized_key_file"`
	// The command to run on the machine being
	//  provisioned by Packer to handle the SFTP protocol that Ansible will use to
	//  transfer files. The command should read and write on stdin and stdout,
	//  respectively. Defaults to `/usr/lib/sftp-server -e`.
	SFTPCmd string `mapstructure:"sftp_command"`
	// Check if ansible is installed prior to
	//  running. Set this to `true`, for example, if you're going to install
	//  ansible during the packer run.
	SkipVersionCheck bool `mapstructure:"skip_version_check"`
	UseSFTP          bool `mapstructure:"use_sftp"`
	// The directory in which to place the
	//  temporary generated Ansible inventory file. By default, this is the
	//  system-specific temporary file location. The fully-qualified name of this
	//  temporary file will be passed to the `-i` argument of the `ansible` command
	//  when this provisioner runs ansible. Specify this if you have an existing
	//  inventory directory with `host_vars` `group_vars` that you would like to
	//  use in the playbook that this provisioner will run.
	InventoryDirectory string `mapstructure:"inventory_directory"`
	// This template represents the format for the lines added to the temporary
	// inventory file that Packer will create to run Ansible against your image.
	// The default for recent versions of Ansible is:
	// "{{ .HostAlias }} ansible_host={{ .Host }} ansible_user={{ .User }} ansible_port={{ .Port }}\n"
	// Available template engines are: This option is a template engine;
	// variables available to you include the examples in the default (Host,
	// HostAlias, User, Port) as well as any variables available to you via the
	// "build" template engine.
	InventoryFileTemplate string `mapstructure:"inventory_file_template"`
	// The inventory file to use during provisioning.
	//  When unspecified, Packer will create a temporary inventory file and will
	//  use the `host_alias`.
	InventoryFile string `mapstructure:"inventory_file"`
	// If `true`, the Ansible provisioner will
	//  not delete the temporary inventory file it creates in order to connect to
	//  the instance. This is useful if you are trying to debug your ansible run
	//  and using "--on-error=ask" in order to leave your instance running while you
	//  test your playbook. this option is not used if you set an `inventory_file`.
	KeepInventoryFile bool `mapstructure:"keep_inventory_file"`
	// A requirements file which provides a way to
	//  install roles or collections with the [ansible-galaxy
	//  cli](https://docs.ansible.com/ansible/latest/galaxy/user_guide.html#the-ansible-galaxy-command-line-tool)
	//  on the local machine before executing `ansible-playbook`. By default, this is empty.
	GalaxyFile string `mapstructure:"galaxy_file"`
	// The command to invoke ansible-galaxy. By default, this is
	// `ansible-galaxy`.
	GalaxyCommand string `mapstructure:"galaxy_command"`
	// Force overwriting an existing role.
	//  Adds `--force` option to `ansible-galaxy` command. By default, this is
	//  `false`.
	GalaxyForceInstall bool `mapstructure:"galaxy_force_install"`
	// The path to the directory on your local system in which to
	//   install the roles. Adds `--roles-path /path/to/your/roles` to
	//   `ansible-galaxy` command. By default, this is empty, and thus `--roles-path`
	//   option is not added to the command.
	RolesPath string `mapstructure:"roles_path"`
	// The path to the directory on your local system in which to
	//   install the collections. Adds `--collections-path /path/to/your/collections` to
	//   `ansible-galaxy` command. By default, this is empty, and thus `--collections-path`
	//   option is not added to the command.
	CollectionsPath string `mapstructure:"collections_path"`
	// When `true`, set up a localhost proxy adapter
	// so that Ansible has an IP address to connect to, even if your guest does not
	// have an IP address. For example, the adapter is necessary for Docker builds
	// to use the Ansible provisioner. If you set this option to `false`, but
	// Packer cannot find an IP address to connect Ansible to, it will
	// automatically set up the adapter anyway.
	//
	//  In order for Ansible to connect properly even when use_proxy is false, you
	// need to make sure that you are either providing a valid username and ssh key
	// to the ansible provisioner directly, or that the username and ssh key
	// being used by the ssh communicator will work for your needs. If you do not
	// provide a user to ansible, it will use the user associated with your
	// builder, not the user running Packer.
	//  use_proxy=false is currently only supported for SSH and WinRM.
	//
	// Currently, this defaults to `true` for all connection types. In the future,
	// this option will be changed to default to `false` for SSH and WinRM
	// connections where the provisioner has access to a host IP.
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

	setupAdapterFunc   func(ui packersdk.Ui, comm packersdk.Communicator) (string, error)
	executeAnsibleFunc func(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	p.done = make(chan struct{})

	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "ansible",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"inventory_file_template",
			},
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

	var errs *packersdk.MultiError
	err = validateFileConfig(p.config.PlaybookFile, "playbook_file", true)
	if err != nil {
		errs = packersdk.MultiErrorAppend(errs, err)
	}

	// Check that the galaxy file exists, if configured
	if len(p.config.GalaxyFile) > 0 {
		err = validateFileConfig(p.config.GalaxyFile, "galaxy_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the authorized key file exists
	if len(p.config.SSHAuthorizedKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHAuthorizedKeyFile, "ssh_authorized_key_file", true)
		if err != nil {
			log.Println(p.config.SSHAuthorizedKeyFile, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}
	if len(p.config.SSHHostKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHHostKeyFile, "ssh_host_key_file", true)
		if err != nil {
			log.Println(p.config.SSHHostKeyFile, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	} else {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_HOST_KEY_CHECKING=False")
	}

	if !p.config.UseSFTP {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_SCP_IF_SSH=True")
	}

	if p.config.LocalPort > 65535 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("local_port: %d must be a valid port", p.config.LocalPort))
	}

	if len(p.config.InventoryDirectory) > 0 {
		err = validateInventoryDirectoryConfig(p.config.InventoryDirectory)
		if err != nil {
			log.Println(p.config.InventoryDirectory, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if !p.config.SkipVersionCheck {
		err = p.getVersion()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if p.config.User == "" {
		p.config.userWasEmpty = true
		usr, err := user.Current()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		} else {
			p.config.User = usr.Username
		}
	}
	if p.config.User == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("user: could not determine current user from environment."))
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

func (p *Provisioner) setupAdapter(ui packersdk.Ui, comm packersdk.Communicator) (string, error) {
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
				return nil, fmt.Errorf("authentication failed: %s is not a valid user", user)
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

	// If user has defiend their own inventory template, use it.
	hostTemplate := p.config.InventoryFileTemplate
	if hostTemplate == "" {
		// figure out which inventory line template to use
		hostTemplate = DefaultSSHInventoryFilev2
		if p.ansibleMajVersion < 2 {
			hostTemplate = DefaultSSHInventoryFilev1
		}
		if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
			hostTemplate = DefaultWinRMInventoryFilev2
		}
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

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
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
			SSHAgentAuth := generatedData["SSHAgentAuth"].(bool)
			if SSHPrivateKeyFile != "" || SSHAgentAuth {
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

func (p *Provisioner) executeGalaxy(ui packersdk.Ui, comm packersdk.Communicator) error {
	galaxyFile := filepath.ToSlash(p.config.GalaxyFile)

	// ansible-galaxy install -r requirements.yml
	roleArgs := []string{"install", "-r", galaxyFile}
	// Instead of modifying args depending on config values and removing or modifying values from
	// the slice between role and collection installs, just use 2 slices and simplify everything
	collectionArgs := []string{"collection", "install", "-r", galaxyFile}
	// Add force to arguments
	if p.config.GalaxyForceInstall {
		roleArgs = append(roleArgs, "-f")
		collectionArgs = append(collectionArgs, "-f")
	}

	// Add roles_path argument if specified
	if p.config.RolesPath != "" {
		roleArgs = append(roleArgs, "-p", filepath.ToSlash(p.config.RolesPath))
	}
	// Add collections_path argument if specified
	if p.config.CollectionsPath != "" {
		collectionArgs = append(collectionArgs, "-p", filepath.ToSlash(p.config.CollectionsPath))
	}

	// Search galaxy_file for roles and collections keywords
	f, err := ioutil.ReadFile(galaxyFile)
	if err != nil {
		return err
	}
	hasRoles, _ := regexp.Match(`(?m)^roles:`, f)
	hasCollections, _ := regexp.Match(`(?m)^collections:`, f)

	// If if roles keyword present (v2 format), or no collections keywork present (v1), install roles
	if hasRoles || !hasCollections {
		if roleInstallError := p.invokeGalaxyCommand(roleArgs, ui, comm); roleInstallError != nil {
			return roleInstallError
		}
	}

	// If collections keyword present (v2 format), install collections
	if hasCollections {
		if collectionInstallError := p.invokeGalaxyCommand(collectionArgs, ui, comm); collectionInstallError != nil {
			return collectionInstallError
		}
	}

	return nil
}

// Intended to be invoked from p.executeGalaxy depending on the Ansible Galaxy parameters passed to Packer
func (p *Provisioner) invokeGalaxyCommand(args []string, ui packersdk.Ui, comm packersdk.Communicator) error {
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

	//Setting up AnsibleEnvVars at begining so additional checks can take them into account
	if len(p.config.AnsibleEnvVars) > 0 {
		envVars = append(envVars, p.config.AnsibleEnvVars...)
	}

	if p.config.PackerBuildName != "" {
		// HCL configs don't currently have the PakcerBuildName. Don't
		// cause weirdness with a half-set variable
		args = append(args, "-e", fmt.Sprintf("packer_build_name=%q", p.config.PackerBuildName))
	}

	args = append(args, "-e", fmt.Sprintf("packer_builder_type=%s", p.config.PackerBuilderType))

	// expose packer_http_addr extra variable
	if httpAddr != commonsteps.HttpAddrNotImplemented {
		args = append(args, "-e", fmt.Sprintf("packer_http_addr=%s", httpAddr))
	}

	if p.generatedData["ConnType"] == "ssh" && len(privKeyFile) > 0 {
		// Add ssh extra args to set IdentitiesOnly
		if len(p.config.AnsibleSSHExtraArgs) > 0 {
			var sshExtraArgs string
			for _, sshExtraArg := range p.config.AnsibleSSHExtraArgs {
				sshExtraArgs = sshExtraArgs + sshExtraArg
			}
			args = append(args, "--ssh-extra-args", fmt.Sprintf("'%s'", sshExtraArgs))
		} else {
			args = append(args, "--ssh-extra-args", "'-o IdentitiesOnly=yes'")
		}
	}

	args = append(args, p.config.ExtraArguments...)

	// Add password to ansible call.
	if !checkArg("ansible_password", args) && p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
		args = append(args, "-e", fmt.Sprintf("ansible_password=%s", p.generatedData["Password"]))
	}

	if !checkArg("ansible_password", args) && len(privKeyFile) > 0 {
		// "-e ansible_ssh_private_key_file" is preferable to "--private-key"
		// because it is a higher priority variable and therefore won't get
		// overridden by dynamic variables. See #5852 for more details.
		args = append(args, "-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", privKeyFile))
	}

	if checkArg("ansible_password", args) && p.generatedData["ConnType"] == "ssh" {
		if !checkArg("ansible_host_key_checking", args) && !checkArg("ANSIBLE_HOST_KEY_CHECKING", envVars) {
			args = append(args, "-e", "ansible_host_key_checking=False")
		}
	}
	// This must be the last arg appended to args
	args = append(args, "-i", inventory, playbook)
	return args, envVars
}

func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error {
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
	if checkArg("ansible_password", args) {
		usePass, ok := p.generatedData["Password"]
		if ok && usePass != "" {
			sanitized = strings.Replace(sanitized, usePass.(string), "*****", -1)
		}
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

//checkArg Evaluates if argname is in args
func checkArg(argname string, args []string) bool {
	for _, arg := range args {
		for _, ansibleArg := range strings.Split(arg, "=") {
			if ansibleArg == argname {
				return true
			}
		}
	}
	return false
}
