package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh/agent"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/core"
)

type SuiteRemote struct{}

var _ = Suite(&SuiteRemote{})

const (
	fixRepo             = "git@github.com:tyba/git-fixture.git"
	fixRepoBadVcs       = "www.example.com"
	fixRepoNonGit       = "https://code.google.com/p/go"
	fixGitRepoNonGithub = "https://bitbucket.org/user/repo.git"
)

func (s *SuiteRemote) TestConnect(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Connect(fixRepo), Equals, ErrAuthRequired)
}

// We will use a running ssh agent for testing
// ssh authentication.
type sshAgentConn struct {
	pipe net.Conn
	auth *PublicKeysCallback
}

// Opens a pipe with the ssh agent and uses the pipe
// as the implementer of the public key callback function.
func newSSHAgentConn() (*sshAgentConn, error) {
	pipe, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return &sshAgentConn{
		pipe: pipe,
		auth: &PublicKeysCallback{
			User:     "git",
			Callback: agent.NewClient(pipe).Signers,
		},
	}, nil
}

// Closes the pipe with the ssh agent
func (c *sshAgentConn) close() error {
	return c.pipe.Close()
}

func (s *SuiteRemote) SetUpSuite(c *C) {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		c.Skip("SSH_AUTH_SOCK is not set")
	}
}

func (s *SuiteRemote) TestConnectWithPublicKeysCallback(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	defer func() { c.Assert(r.Disconnect(), IsNil) }()
	c.Assert(r.connected, Equals, true)
	c.Assert(r.auth, Equals, agent.auth)
}

func (s *SuiteRemote) TestConnectBadVcs(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepoBadVcs, nil), ErrorMatches, fmt.Sprintf(".*%s.*", fixRepoBadVcs))
}

func (s *SuiteRemote) TestConnectNonGit(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepoNonGit, nil), Equals, ErrUnsupportedVCS)
}

func (s *SuiteRemote) TestConnectNonGithub(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixGitRepoNonGithub, nil), Equals, ErrUnsupportedRepo)
}

// A mock implementation of client.common.AuthMethod
// to test non ssh auth method detection.
type mockAuth struct{}

func (*mockAuth) Name() string   { return "" }
func (*mockAuth) String() string { return "" }

func (s *SuiteRemote) TestConnectWithAuthWrongType(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, &mockAuth{}), Equals, ErrInvalidAuthMethod)
	c.Assert(r.connected, Equals, false)
}

func (s *SuiteRemote) TestAlreadyConnected(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	defer func() { c.Assert(r.Disconnect(), IsNil) }()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), Equals, ErrAlreadyConnected)
	c.Assert(r.connected, Equals, true)
}

func (s *SuiteRemote) TestDisconnect(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	c.Assert(r.Disconnect(), IsNil)
	c.Assert(r.connected, Equals, false)
}

func (s *SuiteRemote) TestDisconnectedWhenNonConnected(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Disconnect(), Equals, ErrNotConnected)
}

func (s *SuiteRemote) TestAlreadyDisconnected(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	c.Assert(r.Disconnect(), IsNil)
	c.Assert(r.Disconnect(), Equals, ErrNotConnected)
	c.Assert(r.connected, Equals, false)
}

func (s *SuiteRemote) TestServeralConnections(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	c.Assert(r.Disconnect(), IsNil)

	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	c.Assert(r.connected, Equals, true)
	c.Assert(r.Disconnect(), IsNil)
	c.Assert(r.connected, Equals, false)

	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	c.Assert(r.connected, Equals, true)
	c.Assert(r.Disconnect(), IsNil)
	c.Assert(r.connected, Equals, false)
}

func (s *SuiteRemote) TestInfoNotConnected(c *C) {
	r := NewGitUploadPackService()
	_, err := r.Info()
	c.Assert(err, Equals, ErrNotConnected)
}

func (s *SuiteRemote) TestDefaultBranch(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	defer func() { c.Assert(r.Disconnect(), IsNil) }()

	info, err := r.Info()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.SymbolicReference("HEAD"), Equals, "refs/heads/master")
}

func (s *SuiteRemote) TestCapabilities(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	defer func() { c.Assert(r.Disconnect(), IsNil) }()

	info, err := r.Info()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.Get("agent").Values, HasLen, 1)
}

func (s *SuiteRemote) TestFetchNotConnected(c *C) {
	r := NewGitUploadPackService()
	pr := &common.GitUploadPackRequest{}
	pr.Want(core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	_, err := r.Fetch(pr)
	c.Assert(err, Equals, ErrNotConnected)
}

func (s *SuiteRemote) TestFetch(c *C) {
	agent, err := newSSHAgentConn()
	c.Assert(err, IsNil)
	defer func() { c.Assert(agent.close(), IsNil) }()

	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(fixRepo, agent.auth), IsNil)
	defer func() { c.Assert(r.Disconnect(), IsNil) }()

	pr := &common.GitUploadPackRequest{}
	pr.Want(core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	reader, err := r.Fetch(pr)
	c.Assert(err, IsNil)

	b, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(b, HasLen, 85374)
}
