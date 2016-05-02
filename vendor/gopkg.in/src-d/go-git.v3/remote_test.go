package git

import (
	"gopkg.in/src-d/go-git.v3/clients/http"
	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/formats/packfile"
	"gopkg.in/src-d/go-git.v3/storage/memory"

	. "gopkg.in/check.v1"
)

type SuiteRemote struct{}

var _ = Suite(&SuiteRemote{})

const RepositoryFixture = "https://github.com/tyba/git-fixture"

func (s *SuiteRemote) TestNewAuthenticatedRemote(c *C) {
	auth := &http.BasicAuth{}
	r, err := NewAuthenticatedRemote(RepositoryFixture, auth)
	c.Assert(err, IsNil)
	c.Assert(r.Auth, Equals, auth)
}

func (s *SuiteRemote) TestConnect(c *C) {
	r, err := NewRemote(RepositoryFixture)
	c.Assert(err, IsNil)
	c.Assert(r.Connect(), IsNil)
}

func (s *SuiteRemote) TestDefaultBranch(c *C) {
	r, err := NewRemote(RepositoryFixture)
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Connect(), IsNil)
	c.Assert(r.DefaultBranch(), Equals, "refs/heads/master")
}

func (s *SuiteRemote) TestCapabilities(c *C) {
	r, err := NewRemote(RepositoryFixture)
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Connect(), IsNil)
	c.Assert(r.Capabilities().Get("agent").Values, HasLen, 1)
}

func (s *SuiteRemote) TestFetchDefaultBranch(c *C) {
	r, err := NewRemote(RepositoryFixture)
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Connect(), IsNil)

	reader, err := r.FetchDefaultBranch()
	c.Assert(err, IsNil)

	pr := packfile.NewReader(reader)

	storage := memory.NewObjectStorage()
	_, err = pr.Read(storage)
	c.Assert(err, IsNil)
	c.Assert(storage.Objects, HasLen, 28)
}

func (s *SuiteRemote) TestHead(c *C) {
	r, err := NewRemote(RepositoryFixture)
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)

	err = r.Connect()
	c.Assert(err, IsNil)

	hash, err := r.Head()
	c.Assert(err, IsNil)
	c.Assert(hash, Equals, core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
}
