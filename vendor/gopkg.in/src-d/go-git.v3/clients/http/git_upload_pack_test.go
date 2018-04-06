package http

import (
	"io/ioutil"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/core"
)

type SuiteRemote struct{}

var _ = Suite(&SuiteRemote{})

const RepositoryFixture = "https://github.com/tyba/git-fixture"

func (s *SuiteRemote) TestConnect(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Connect(RepositoryFixture), IsNil)
}

func (s *SuiteRemote) TestConnectWithAuth(c *C) {
	auth := &BasicAuth{}
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(RepositoryFixture, auth), IsNil)
	c.Assert(r.auth, Equals, auth)
}

type mockAuth struct{}

func (*mockAuth) Name() string   { return "" }
func (*mockAuth) String() string { return "" }

func (s *SuiteRemote) TestConnectWithAuthWrongType(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.ConnectWithAuth(RepositoryFixture, &mockAuth{}), Equals, InvalidAuthMethodErr)
}

func (s *SuiteRemote) TestDefaultBranch(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Connect(RepositoryFixture), IsNil)

	info, err := r.Info()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.SymbolicReference("HEAD"), Equals, "refs/heads/master")
}

func (s *SuiteRemote) TestCapabilities(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Connect(RepositoryFixture), IsNil)

	info, err := r.Info()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.Get("agent").Values, HasLen, 1)
}

func (s *SuiteRemote) TestFetch(c *C) {
	r := NewGitUploadPackService()
	c.Assert(r.Connect(RepositoryFixture), IsNil)

	req := &common.GitUploadPackRequest{}
	req.Want(core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	reader, err := r.Fetch(req)
	c.Assert(err, IsNil)

	b, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(b, HasLen, 85374)
}
