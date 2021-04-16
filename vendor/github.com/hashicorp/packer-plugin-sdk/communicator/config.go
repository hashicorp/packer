//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config,SSH,WinRM,SSHTemporaryKeyPair

package communicator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	helperssh "github.com/hashicorp/packer-plugin-sdk/communicator/ssh"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	packerssh "github.com/hashicorp/packer-plugin-sdk/sdk-internals/communicator/ssh"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/masterzen/winrm"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Config is the common configuration a builder uses to define and configure a Packer
// communicator. Embed this struct in your builder config to implement
// communicator support.
type Config struct {
	// Packer currently supports three kinds of communicators:
	//
	// -   `none` - No communicator will be used. If this is set, most
	//     provisioners also can't be used.
	//
	// -   `ssh` - An SSH connection will be established to the machine. This
	//     is usually the default.
	//
	// -   `winrm` - A WinRM connection will be established.
	//
	// In addition to the above, some builders have custom communicators they
	// can use. For example, the Docker builder has a "docker" communicator
	// that uses `docker exec` and `docker cp` to execute scripts and copy
	// files.
	Type string `mapstructure:"communicator"`
	// We recommend that you enable SSH or WinRM as the very last step in your
	// guest's bootstrap script, but sometimes you may have a race condition
	// where you need Packer to wait before attempting to connect to your
	// guest.
	//
	// If you end up in this situation, you can use the template option
	// `pause_before_connecting`. By default, there is no pause. For example if
	// you set `pause_before_connecting` to `10m` Packer will check whether it
	// can connect, as normal. But once a connection attempt is successful, it
	// will disconnect and then wait 10 minutes before connecting to the guest
	// and beginning provisioning.
	PauseBeforeConnect time.Duration `mapstructure:"pause_before_connecting"`

	SSH   `mapstructure:",squash"`
	WinRM `mapstructure:",squash"`
}

// The SSH config defines configuration for the SSH communicator.
type SSH struct {
	// The address to SSH to. This usually is automatically configured by the
	// builder.
	SSHHost string `mapstructure:"ssh_host"`
	// The port to connect to SSH. This defaults to `22`.
	SSHPort int `mapstructure:"ssh_port"`
	// The username to connect to SSH with. Required if using SSH.
	SSHUsername string `mapstructure:"ssh_username"`
	// A plaintext password to use to authenticate with SSH.
	SSHPassword string `mapstructure:"ssh_password"`
	// If specified, this is the key that will be used for SSH with the
	// machine. The key must match a key pair name loaded up into the remote.
	// By default, this is blank, and Packer will generate a temporary keypair
	// unless [`ssh_password`](#ssh_password) is used.
	// [`ssh_private_key_file`](#ssh_private_key_file) or
	// [`ssh_agent_auth`](#ssh_agent_auth) must be specified when
	// [`ssh_keypair_name`](#ssh_keypair_name) is utilized.
	SSHKeyPairName string `mapstructure:"ssh_keypair_name" undocumented:"true"`
	// The name of the temporary key pair to generate. By default, Packer
	// generates a name that looks like `packer_<UUID>`, where &lt;UUID&gt; is
	// a 36 character unique identifier.
	SSHTemporaryKeyPairName string `mapstructure:"temporary_key_pair_name" undocumented:"true"`
	SSHTemporaryKeyPair     `mapstructure:",squash"`
	// This overrides the value of ciphers supported by default by golang.
	// The default value is [
	//   "aes128-gcm@openssh.com",
	//   "chacha20-poly1305@openssh.com",
	//   "aes128-ctr", "aes192-ctr", "aes256-ctr",
	// ]
	//
	// Valid options for ciphers include:
	// "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
	// "chacha20-poly1305@openssh.com",
	// "arcfour256", "arcfour128", "arcfour", "aes128-cbc", "3des-cbc",
	SSHCiphers []string `mapstructure:"ssh_ciphers"`
	// If true, Packer will attempt to remove its temporary key from
	// `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
	// mostly cosmetic option, since Packer will delete the temporary private
	// key from the host system regardless of whether this is set to true
	// (unless the user has set the `-debug` flag). Defaults to "false";
	// currently only works on guests with `sed` installed.
	SSHClearAuthorizedKeys bool `mapstructure:"ssh_clear_authorized_keys"`
	// If set, Packer will override the value of key exchange (kex) altorighms
	// supported by default by golang. Acceptable values include:
	// "curve25519-sha256@libssh.org", "ecdh-sha2-nistp256",
	// "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
	// "diffie-hellman-group14-sha1", and "diffie-hellman-group1-sha1".
	SSHKEXAlgos []string `mapstructure:"ssh_key_exchange_algorithms"`
	// Path to a PEM encoded private key file to use to authenticate with SSH.
	// The `~` can be used in path and will be expanded to the home directory
	// of current user.
	SSHPrivateKeyFile string `mapstructure:"ssh_private_key_file" undocumented:"true"`
	// Path to user certificate used to authenticate with SSH.
	// The `~` can be used in path and will be expanded to the
	// home directory of current user.
	SSHCertificateFile string `mapstructure:"ssh_certificate_file"`
	// If `true`, a PTY will be requested for the SSH connection. This defaults
	// to `false`.
	SSHPty bool `mapstructure:"ssh_pty"`
	// The time to wait for SSH to become available. Packer uses this to
	// determine when the machine has booted so this is usually quite long.
	// Example value: `10m`.
	SSHTimeout time.Duration `mapstructure:"ssh_timeout"`
	// Deprecated in favor of SSHTimeout
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout" undocumented:"true"`
	// If true, the local SSH agent will be used to authenticate connections to
	// the source instance. No temporary keypair will be created, and the
	// values of [`ssh_password`](#ssh_password) and
	// [`ssh_private_key_file`](#ssh_private_key_file) will be ignored. The
	// environment variable `SSH_AUTH_SOCK` must be set for this option to work
	// properly.
	SSHAgentAuth bool `mapstructure:"ssh_agent_auth" undocumented:"true"`
	// If true, SSH agent forwarding will be disabled. Defaults to `false`.
	SSHDisableAgentForwarding bool `mapstructure:"ssh_disable_agent_forwarding"`
	// The number of handshakes to attempt with SSH once it can connect. This
	// defaults to `10`.
	SSHHandshakeAttempts int `mapstructure:"ssh_handshake_attempts"`
	// A bastion host to use for the actual SSH connection.
	SSHBastionHost string `mapstructure:"ssh_bastion_host"`
	// The port of the bastion host. Defaults to `22`.
	SSHBastionPort int `mapstructure:"ssh_bastion_port"`
	// If `true`, the local SSH agent will be used to authenticate with the
	// bastion host. Defaults to `false`.
	SSHBastionAgentAuth bool `mapstructure:"ssh_bastion_agent_auth"`
	// The username to connect to the bastion host.
	SSHBastionUsername string `mapstructure:"ssh_bastion_username"`
	// The password to use to authenticate with the bastion host.
	SSHBastionPassword string `mapstructure:"ssh_bastion_password"`
	// If `true`, the keyboard-interactive used to authenticate with bastion host.
	SSHBastionInteractive bool `mapstructure:"ssh_bastion_interactive"`
	// Path to a PEM encoded private key file to use to authenticate with the
	// bastion host. The `~` can be used in path and will be expanded to the
	// home directory of current user.
	SSHBastionPrivateKeyFile string `mapstructure:"ssh_bastion_private_key_file"`
	// Path to user certificate used to authenticate with bastion host.
	// The `~` can be used in path and will be expanded to the
	//home directory of current user.
	SSHBastionCertificateFile string `mapstructure:"ssh_bastion_certificate_file"`
	// `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
	// File Transfer Protocol.
	SSHFileTransferMethod string `mapstructure:"ssh_file_transfer_method"`
	// A SOCKS proxy host to use for SSH connection
	SSHProxyHost string `mapstructure:"ssh_proxy_host"`
	// A port of the SOCKS proxy. Defaults to `1080`.
	SSHProxyPort int `mapstructure:"ssh_proxy_port"`
	// The optional username to authenticate with the proxy server.
	SSHProxyUsername string `mapstructure:"ssh_proxy_username"`
	// The optional password to use to authenticate with the proxy server.
	SSHProxyPassword string `mapstructure:"ssh_proxy_password"`
	// How often to send "keep alive" messages to the server. Set to a negative
	// value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.
	SSHKeepAliveInterval time.Duration `mapstructure:"ssh_keep_alive_interval"`
	// The amount of time to wait for a remote command to end. This might be
	// useful if, for example, packer hangs on a connection after a reboot.
	// Example: `5m`. Disabled by default.
	SSHReadWriteTimeout time.Duration `mapstructure:"ssh_read_write_timeout"`

	// Tunneling

	//
	SSHRemoteTunnels []string `mapstructure:"ssh_remote_tunnels"`
	//
	SSHLocalTunnels []string `mapstructure:"ssh_local_tunnels"`

	// SSH Internals
	SSHPublicKey  []byte `mapstructure:"ssh_public_key" undocumented:"true"`
	SSHPrivateKey []byte `mapstructure:"ssh_private_key" undocumented:"true"`
}

// When no ssh credentials are specified, Packer will generate a temporary SSH
// keypair for the instance. You can change the algorithm type and bits
// settings.
type SSHTemporaryKeyPair struct {
	// `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
	//
	// Specifies the type of key to create. The possible values are 'dsa',
	// 'ecdsa', 'ed25519', or 'rsa'.
	SSHTemporaryKeyPairType string `mapstructure:"temporary_key_pair_type"`
	// Specifies the number of bits in the key to create. For RSA keys, the
	// minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
	// bits is considered sufficient. DSA keys must be exactly 1024 bits as
	// specified by FIPS 186-2. For ECDSA keys, bits determines the key length
	// by selecting from one of three elliptic curve sizes: 256, 384 or 521
	// bits. Attempting to use bit lengths other than these three values for
	// ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
	// ignored.
	SSHTemporaryKeyPairBits int `mapstructure:"temporary_key_pair_bits"`
}

// The WinRM config defines configuration for the WinRM communicator.
type WinRM struct {
	// The username to use to connect to WinRM.
	WinRMUser string `mapstructure:"winrm_username"`
	// The password to use to connect to WinRM.
	WinRMPassword string `mapstructure:"winrm_password"`
	// The address for WinRM to connect to.
	//
	// NOTE: If using an Amazon EBS builder, you can specify the interface
	// WinRM connects to via
	// [`ssh_interface`](/docs/builders/amazon-ebs#ssh_interface)
	WinRMHost string `mapstructure:"winrm_host"`
	// Setting this to `true` adds the remote
	// `host:port` to the `NO_PROXY` environment variable. This has the effect of
	// bypassing any configured proxies when connecting to the remote host.
	// Default to `false`.
	WinRMNoProxy bool `mapstructure:"winrm_no_proxy"`
	// The WinRM port to connect to. This defaults to `5985` for plain
	// unencrypted connection and `5986` for SSL when `winrm_use_ssl` is set to
	// true.
	WinRMPort int `mapstructure:"winrm_port"`
	// The amount of time to wait for WinRM to become available. This defaults
	// to `30m` since setting up a Windows machine generally takes a long time.
	WinRMTimeout time.Duration `mapstructure:"winrm_timeout"`
	// If `true`, use HTTPS for WinRM.
	WinRMUseSSL bool `mapstructure:"winrm_use_ssl"`
	// If `true`, do not check server certificate chain and host name.
	WinRMInsecure bool `mapstructure:"winrm_insecure"`
	// If `true`, NTLMv2 authentication (with session security) will be used
	// for WinRM, rather than default (basic authentication), removing the
	// requirement for basic authentication to be enabled within the target
	// guest. Further reading for remote connection authentication can be found
	// [here](https://msdn.microsoft.com/en-us/library/aa384295(v=vs.85).aspx).
	WinRMUseNTLM            bool `mapstructure:"winrm_use_ntlm"`
	WinRMTransportDecorator func() winrm.Transporter
}

// The ConfigSpec funcs are used by the Packer core to parse HCL2 templates.
func (c *SSH) ConfigSpec() hcldec.ObjectSpec { return c.FlatMapstructure().HCL2Spec() }

// The ConfigSpec funcs are used by the Packer core to parse HCL2 templates.
func (c *WinRM) ConfigSpec() hcldec.ObjectSpec { return c.FlatMapstructure().HCL2Spec() }

// Configure parses the json template into the Config structs
func (c *SSH) Configure(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, nil, raws...)
	return nil, err
}

// Configure parses the json template into the Config structs
func (c *WinRM) Configure(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, nil, raws...)
	return nil, err
}

var (
	_ packersdk.ConfigurableCommunicator = new(SSH)
	_ packersdk.ConfigurableCommunicator = new(WinRM)
)

// SSHInterface defines whether to use public or private, addresses, and whether
// to use IPv4 or IPv6.
type SSHInterface struct {
	// One of `public_ip`, `private_ip`, `public_dns`, or `private_dns`. If
	// set, either the public IP address, private IP address, public DNS name
	// or private DNS name will used as the host for SSH. The default behaviour
	// if inside a VPC is to use the public IP address if available, otherwise
	// the private IP address will be used. If not in a VPC the public DNS name
	// will be used. Also works for WinRM.
	//
	// Where Packer is configured for an outbound proxy but WinRM traffic
	// should be direct, `ssh_interface` must be set to `private_dns` and
	// `<region>.compute.internal` included in the `NO_PROXY` environment
	// variable.
	SSHInterface string `mapstructure:"ssh_interface"`
	// The IP version to use for SSH connections, valid values are `4` and `6`.
	// Useful on dual stacked instances where the default behavior is to
	// connect via whichever IP address is returned first from the OpenStack
	// API.
	SSHIPVersion string `mapstructure:"ssh_ip_version"`
}

// ReadSSHPrivateKeyFile returns the SSH private key bytes.
func (c *Config) ReadSSHPrivateKeyFile() ([]byte, error) {
	var privateKey []byte

	if c.SSHPrivateKeyFile != "" {
		keyPath, err := pathing.ExpandUser(c.SSHPrivateKeyFile)
		if err != nil {
			return []byte{}, fmt.Errorf("Error expanding path for SSH private key: %s", err)
		}

		privateKey, err = ioutil.ReadFile(keyPath)
		if err != nil {
			return privateKey, fmt.Errorf("Error on reading SSH private key: %s", err)
		}
	}
	return privateKey, nil
}

// SSHConfigFunc returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the private key
// or password.
func (c *Config) SSHConfigFunc() func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		sshConfig := &ssh.ClientConfig{
			User:            c.SSHUsername,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if len(c.SSHCiphers) != 0 {
			sshConfig.Config.Ciphers = c.SSHCiphers
		}

		if len(c.SSHKEXAlgos) != 0 {
			sshConfig.Config.KeyExchanges = c.SSHKEXAlgos
		}

		if c.SSHAgentAuth {
			authSock := os.Getenv("SSH_AUTH_SOCK")
			if authSock == "" {
				return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
			}

			sshAgent, err := net.Dial("unix", authSock)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
			}

			sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		}

		var privateKeys [][]byte
		if c.SSHPrivateKeyFile != "" {
			privateKey, err := c.ReadSSHPrivateKeyFile()
			if err != nil {
				return nil, err
			}
			privateKeys = append(privateKeys, privateKey)
		}

		// aws,alicloud,cloudstack,digitalOcean,oneAndOne,openstack,oracle & profitbricks key
		if iKey, hasKey := state.GetOk("privateKey"); hasKey {
			privateKeys = append(privateKeys, []byte(iKey.(string)))
		}

		if len(c.SSHPrivateKey) != 0 {
			privateKeys = append(privateKeys, c.SSHPrivateKey)
		}

		certPath := ""
		if c.SSHCertificateFile != "" {
			var err error
			certPath, err = pathing.ExpandUser(c.SSHCertificateFile)
			if err != nil {
				return nil, err
			}
		}

		for _, key := range privateKeys {

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				return nil, fmt.Errorf("Error on parsing SSH private key: %s", err)
			}

			if certPath != "" {
				signer, err = helperssh.ReadCertificate(certPath, signer)
				if err != nil {
					return nil, err
				}
			}

			sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
		}

		if c.SSHPassword != "" {
			sshConfig.Auth = append(sshConfig.Auth,
				ssh.Password(c.SSHPassword),
				ssh.KeyboardInteractive(packerssh.PasswordKeyboardInteractive(c.SSHPassword)),
			)
		}
		return sshConfig, nil
	}
}

// Port returns the port that will be used for access based on config.
func (c *Config) Port() int {
	switch c.Type {
	case "ssh":
		return c.SSHPort
	case "winrm":
		return c.WinRMPort
	default:
		return 0
	}
}

// Host returns the host that will be used for access based on config.
func (c *Config) Host() string {
	switch c.Type {
	case "ssh":
		return c.SSHHost
	case "winrm":
		return c.WinRMHost
	default:
		return ""
	}
}

// User returns the user that will be used for access based on config.
func (c *Config) User() string {
	switch c.Type {
	case "ssh":
		return c.SSHUsername
	case "winrm":
		return c.WinRMUser
	default:
		return ""
	}
}

// Password returns the password that will be used for access based on config.
func (c *Config) Password() string {
	switch c.Type {
	case "ssh":
		return c.SSHPassword
	case "winrm":
		return c.WinRMPassword
	default:
		return ""
	}
}

func (c *Config) Prepare(ctx *interpolate.Context) []error {
	if c.Type == "" {
		c.Type = "ssh"
	}

	var errs []error
	switch c.Type {
	case "ssh":
		if es := c.prepareSSH(ctx); len(es) > 0 {
			errs = append(errs, es...)
		}
	case "winrm":
		if es := c.prepareWinRM(ctx); len(es) > 0 {
			errs = append(errs, es...)
		}
	case "docker", "dockerWindowsContainer", "none":
		break
	default:
		return []error{fmt.Errorf("Communicator type %s is invalid", c.Type)}
	}

	return errs
}

func (c *Config) prepareSSH(ctx *interpolate.Context) []error {
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.SSHTimeout == 0 {
		c.SSHTimeout = 5 * time.Minute
	}

	if c.SSHKeepAliveInterval == 0 {
		c.SSHKeepAliveInterval = 5 * time.Second
	}

	if c.SSHHandshakeAttempts == 0 {
		c.SSHHandshakeAttempts = 10
	}

	if c.SSHBastionHost != "" {
		if c.SSHBastionPort == 0 {
			c.SSHBastionPort = 22
		}

		if c.SSHBastionPrivateKeyFile == "" && c.SSHPrivateKeyFile != "" {
			c.SSHBastionPrivateKeyFile = c.SSHPrivateKeyFile
		}

		if c.SSHBastionCertificateFile == "" && c.SSHCertificateFile != "" {
			c.SSHBastionCertificateFile = c.SSHCertificateFile
		}

	}

	if c.SSHProxyHost != "" {
		if c.SSHProxyPort == 0 {
			c.SSHProxyPort = 1080
		}
	}

	if c.SSHFileTransferMethod == "" {
		c.SSHFileTransferMethod = "scp"
	}

	// Backwards compatibility
	if c.SSHWaitTimeout != 0 {
		c.SSHTimeout = c.SSHWaitTimeout
	}

	// Validation
	var errs []error
	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified\n  Note: some builders used to default ssh_username to \"root\"."))
	}

	if c.SSHPrivateKeyFile != "" {
		path, err := pathing.ExpandUser(c.SSHPrivateKeyFile)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_private_key_file is invalid: %s", err))
		} else if _, err := os.Stat(path); err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_private_key_file is invalid: %s", err))
		} else {
			if c.SSHCertificateFile != "" {
				certPath, err := pathing.ExpandUser(c.SSHCertificateFile)
				if err != nil {
					errs = append(errs, fmt.Errorf("invalid identity certificate: #{err}"))
				}

				if _, err := helperssh.FileSignerWithCert(path, certPath); err != nil {
					errs = append(errs, fmt.Errorf(
						"ssh_private_key_file is invalid: %s", err))
				}
			} else {
				if _, err := helperssh.FileSigner(path); err != nil {
					errs = append(errs, fmt.Errorf(
						"ssh_private_key_file is invalid: %s", err))
				}
			}
		}
	}

	if c.SSHBastionHost != "" && !c.SSHBastionAgentAuth {
		if c.SSHBastionPassword == "" && c.SSHBastionPrivateKeyFile == "" {
			errs = append(errs, errors.New(
				"ssh_bastion_password or ssh_bastion_private_key_file must be specified"))
		} else if c.SSHBastionPrivateKeyFile != "" {
			path, err := pathing.ExpandUser(c.SSHBastionPrivateKeyFile)
			if err != nil {
				errs = append(errs, fmt.Errorf(
					"ssh_bastion_private_key_file is invalid: %s", err))
			} else if _, err := os.Stat(path); err != nil {
				errs = append(errs, fmt.Errorf(
					"ssh_bastion_private_key_file is invalid: %s", err))
			} else {
				if c.SSHBastionCertificateFile != "" {
					certPath, err := pathing.ExpandUser(c.SSHBastionCertificateFile)
					if err != nil {
						errs = append(errs, fmt.Errorf("invalid identity certificate: #{err}"))
					}
					if _, err := helperssh.FileSignerWithCert(path, certPath); err != nil {
						errs = append(errs, fmt.Errorf(
							"ssh_bastion_private_key_file is invalid: %s", err))
					}
				} else {
					if _, err := helperssh.FileSigner(path); err != nil {
						errs = append(errs, fmt.Errorf(
							"ssh_bastion_private_key_file is invalid: %s", err))
					}
				}
			}
		}
	}

	if c.SSHFileTransferMethod != "scp" && c.SSHFileTransferMethod != "sftp" {
		errs = append(errs, fmt.Errorf(
			"ssh_file_transfer_method ('%s') is invalid, valid methods: sftp, scp",
			c.SSHFileTransferMethod))
	}

	if c.SSHBastionHost != "" && c.SSHProxyHost != "" {
		errs = append(errs, errors.New("please specify either ssh_bastion_host or ssh_proxy_host, not both"))
	}

	for _, v := range c.SSHLocalTunnels {
		_, err := helperssh.ParseTunnelArgument(v, packerssh.UnsetTunnel)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_local_tunnels ('%s') is invalid: %s", v, err))
		}
	}

	for _, v := range c.SSHRemoteTunnels {
		_, err := helperssh.ParseTunnelArgument(v, packerssh.UnsetTunnel)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_remote_tunnels ('%s') is invalid: %s", v, err))
		}
	}

	return errs
}

func (c *Config) prepareWinRM(ctx *interpolate.Context) (errs []error) {
	if c.WinRMPort == 0 && c.WinRMUseSSL {
		c.WinRMPort = 5986
	} else if c.WinRMPort == 0 {
		c.WinRMPort = 5985
	}

	if c.WinRMTimeout == 0 {
		c.WinRMTimeout = 30 * time.Minute
	}

	if c.WinRMUseNTLM == true {
		c.WinRMTransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }
	}

	if c.WinRMUser == "" {
		errs = append(errs, errors.New("winrm_username must be specified."))
	}

	return errs
}
