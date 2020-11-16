package communicator

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/masterzen/winrm"
)

func testConfig() *Config {
	return &Config{
		SSH: SSH{
			SSHUsername: "root",
		},
	}
}

func TestConfigType(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.Type != "ssh" {
		t.Fatalf("bad: %#v", c)
	}
}

func TestConfig_none(t *testing.T) {
	c := &Config{Type: "none"}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}
}

func TestConfig_badtype(t *testing.T) {
	c := &Config{Type: "foo"}
	if err := c.Prepare(testContext(t)); len(err) != 1 {
		t.Fatalf("bad: %#v", err)
	}
}

func TestConfig_winrm_noport(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser: "admin",
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5985 {
		t.Fatalf("WinRMPort doesn't match default port 5985 when SSL is not enabled and no port is specified.")
	}

}

func TestConfig_winrm_noport_ssl(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser:   "admin",
			WinRMUseSSL: true,
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5986 {
		t.Fatalf("WinRMPort doesn't match default port 5986 when SSL is enabled and no port is specified.")
	}

}

func TestConfig_winrm_port(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser: "admin",
			WinRMPort: 5509,
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5509 {
		t.Fatalf("WinRMPort doesn't match custom port 5509 when SSL is not enabled.")
	}

}

func TestConfig_winrm_port_ssl(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser:   "admin",
			WinRMPort:   5510,
			WinRMUseSSL: true,
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5510 {
		t.Fatalf("WinRMPort doesn't match custom port 5510 when SSL is enabled.")
	}

}

func TestConfig_winrm_use_ntlm(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser:    "admin",
			WinRMUseNTLM: true,
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMTransportDecorator == nil {
		t.Fatalf("WinRMTransportDecorator not set.")
	}

	expected := &winrm.ClientNTLM{}
	actual := c.WinRMTransportDecorator()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("WinRMTransportDecorator isn't ClientNTLM.")
	}

}

func TestSSHBastion(t *testing.T) {
	c := &Config{
		Type: "ssh",
		SSH: SSH{
			SSHUsername:        "root",
			SSHBastionHost:     "mybastionhost.company.com",
			SSHBastionPassword: "test",
		},
	}

	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.SSHBastionCertificateFile != "" {
		t.Fatalf("Identity certificate somehow set")
	}

	if c.SSHPrivateKeyFile != "" {
		t.Fatalf("Private key file somehow set")
	}

}

func TestSSHConfigFunc_ciphers(t *testing.T) {
	state := new(multistep.BasicStateBag)

	// No ciphers set
	c := &Config{
		Type: "ssh",
	}

	f := c.SSHConfigFunc()
	sshConfig, _ := f(state)
	if sshConfig.Config.Ciphers != nil {
		t.Fatalf("Shouldn't set SSHCiphers if communicator config option " +
			"ssh_ciphers is unset.")
	}

	// Ciphers are set
	c = &Config{
		Type: "ssh",
		SSH: SSH{
			SSHCiphers: []string{"partycipher"},
		},
	}
	f = c.SSHConfigFunc()
	sshConfig, _ = f(state)
	if sshConfig.Config.Ciphers == nil {
		t.Fatalf("Shouldn't set SSHCiphers if communicator config option " +
			"ssh_ciphers is unset.")
	}
	if sshConfig.Config.Ciphers[0] != "partycipher" {
		t.Fatalf("ssh_ciphers should be a direct passthrough.")
	}
	if c.SSHCertificateFile != "" {
		t.Fatalf("Identity certificate somehow set")
	}
}

func TestSSHConfigFunc_kexAlgos(t *testing.T) {
	state := new(multistep.BasicStateBag)

	// No ciphers set
	c := &Config{
		Type: "ssh",
	}

	f := c.SSHConfigFunc()
	sshConfig, _ := f(state)
	if sshConfig.Config.KeyExchanges != nil {
		t.Fatalf("Shouldn't set KeyExchanges if communicator config option " +
			"ssh_key_exchange_algorithms is unset.")
	}

	// Ciphers are set
	c = &Config{
		Type: "ssh",
		SSH: SSH{
			SSHKEXAlgos: []string{"partyalgo"},
		},
	}
	f = c.SSHConfigFunc()
	sshConfig, _ = f(state)
	if sshConfig.Config.KeyExchanges == nil {
		t.Fatalf("Should set SSHKEXAlgos if communicator config option " +
			"ssh_key_exchange_algorithms is set.")
	}
	if sshConfig.Config.KeyExchanges[0] != "partyalgo" {
		t.Fatalf("ssh_key_exchange_algorithms should be a direct passthrough.")
	}
	if c.SSHCertificateFile != "" {
		t.Fatalf("Identity certificate somehow set")
	}
}

func TestConfig_winrm(t *testing.T) {
	c := &Config{
		Type: "winrm",
		WinRM: WinRM{
			WinRMUser: "admin",
		},
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}
}

func testContext(t *testing.T) *interpolate.Context {
	return nil
}
