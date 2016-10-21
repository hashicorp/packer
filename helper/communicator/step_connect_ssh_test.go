package communicator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// startAgent sets ssh-agent environment variables
func startAgent(t *testing.T) func() {
	if testing.Short() {
		// ssh-agent is not always available, and the key
		// types supported vary by platform.
		t.Skip("skipping test due to -short or availability")
	}

	bin, err := exec.LookPath("ssh-agent")
	if err != nil {
		t.Skip("could not find ssh-agent")
	}

	cmd := exec.Command(bin, "-s")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("cmd.Output: %v", err)
	}

	/* Output looks like:
	   SSH_AUTH_SOCK=/tmp/ssh-P65gpcqArqvH/agent.15541; export SSH_AUTH_SOCK;
	           SSH_AGENT_PID=15542; export SSH_AGENT_PID;
	           echo Agent pid 15542;
	*/
	fields := bytes.Split(out, []byte(";"))
	line := bytes.SplitN(fields[0], []byte("="), 2)
	line[0] = bytes.TrimLeft(line[0], "\n")
	if string(line[0]) != "SSH_AUTH_SOCK" {
		t.Fatalf("could not find key SSH_AUTH_SOCK in %q", fields[0])
	}
	socket := string(line[1])
	t.Logf("Socket value: %v", socket)

	origSocket := os.Getenv("SSH_AUTH_SOCK")
	if err := os.Setenv("SSH_AUTH_SOCK", socket); err != nil {
		t.Fatalf("could not set SSH_AUTH_SOCK environment variable: %v", err)
	}

	line = bytes.SplitN(fields[2], []byte("="), 2)
	line[0] = bytes.TrimLeft(line[0], "\n")
	if string(line[0]) != "SSH_AGENT_PID" {
		t.Fatalf("could not find key SSH_AGENT_PID in %q", fields[2])
	}
	pidStr := line[1]
	t.Logf("Agent PID: %v", string(pidStr))
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		t.Fatalf("Atoi(%q): %v", pidStr, err)
	}

	return func() {
		proc, _ := os.FindProcess(pid)
		if proc != nil {
			proc.Kill()
		}

		os.Setenv("SSH_AUTH_SOCK", origSocket)
		os.RemoveAll(filepath.Dir(socket))
	}
}

func TestSSHAgent(t *testing.T) {
	cleanup := startAgent(t)
	defer cleanup()

	if auth := sshAgent(); auth == nil {
		t.Error("Want `ssh.AuthMethod`, got `nil`")
	}
}

func TestSSHBastionConfig(t *testing.T) {
	pemPath := TestPEM(t)
	tests := []struct {
		in     *Config
		errStr string
		want   int
		fn     func() func()
	}{
		{
			in:   &Config{SSHDisableAgent: true},
			want: 0,
		},
		{
			in:   &Config{SSHDisableAgent: false},
			want: 0,
			fn: func() func() {
				cleanup := startAgent(t)
				os.Unsetenv("SSH_AUTH_SOCK")
				return cleanup
			},
		},
		{
			in: &Config{
				SSHDisableAgent:      false,
				SSHBastionPassword:   "foobar",
				SSHBastionPrivateKey: pemPath,
			},
			want: 4,
			fn: func() func() {
				cleanup := startAgent(t)
				return cleanup
			},
		},
		{
			in: &Config{
				SSHBastionPrivateKey: pemPath,
			},
			want:   0,
			errStr: "Failed to read key '" + pemPath + "': no key found",
			fn: func() func() {
				os.Truncate(pemPath, 0)
				return func() {
					if err := os.Remove(pemPath); err != nil {
						t.Fatalf("os.Remove: %v", err)
					}
				}
			},
		},
	}

	for _, c := range tests {
		func() {
			if c.fn != nil {
				defered := c.fn()
				defer defered()
			}
			bConf, err := sshBastionConfig(c.in)
			if err != nil {
				if err.Error() != c.errStr {
					t.Errorf("want error %v, got %q", c.errStr, err)
				}
				return
			}

			if len(bConf.Auth) != c.want {
				t.Errorf("want %v ssh.AuthMethod, got %v ssh.AuthMethod", c.want, len(bConf.Auth))
			}
		}()
	}
}
