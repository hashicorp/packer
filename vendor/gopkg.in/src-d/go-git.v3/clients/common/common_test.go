package common

import (
	"bytes"
	"encoding/base64"
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/formats/pktline"
)

func Test(t *testing.T) { TestingT(t) }

type SuiteCommon struct{}

var _ = Suite(&SuiteCommon{})

func (s *SuiteCommon) TestNewEndpoint(c *C) {
	e, err := NewEndpoint("git@github.com:user/repository.git")
	c.Assert(err, IsNil)
	c.Assert(e, Equals, Endpoint("https://github.com/user/repository.git"))
}

func (s *SuiteCommon) TestNewEndpointWrongForgat(c *C) {
	e, err := NewEndpoint("foo")
	c.Assert(err, Not(IsNil))
	c.Assert(e, Equals, Endpoint(""))
}

func (s *SuiteCommon) TestEndpointService(c *C) {
	e, _ := NewEndpoint("git@github.com:user/repository.git")
	c.Assert(e.Service("foo"), Equals, "https://github.com/user/repository.git/info/refs?service=foo")
}

const CapabilitiesFixture = "6ecf0ef2c2dffb796033e5a02219af86ec6584e5 HEADmulti_ack thin-pack side-band side-band-64k ofs-delta shallow no-progress include-tag multi_ack_detailed no-done symref=HEAD:refs/heads/master agent=git/2:2.4.8~dbussink-fix-enterprise-tokens-compilation-1167-gc7006cf"

func (s *SuiteCommon) TestCapabilitiesSymbolicReference(c *C) {
	cap := NewCapabilities()
	cap.Decode(CapabilitiesFixture)
	c.Assert(cap.SymbolicReference("HEAD"), Equals, "refs/heads/master")
}

const GitUploadPackInfoFixture = "MDAxZSMgc2VydmljZT1naXQtdXBsb2FkLXBhY2sKMDAwMDAxMGM2ZWNmMGVmMmMyZGZmYjc5NjAzM2U1YTAyMjE5YWY4NmVjNjU4NGU1IEhFQUQAbXVsdGlfYWNrIHRoaW4tcGFjayBzaWRlLWJhbmQgc2lkZS1iYW5kLTY0ayBvZnMtZGVsdGEgc2hhbGxvdyBuby1wcm9ncmVzcyBpbmNsdWRlLXRhZyBtdWx0aV9hY2tfZGV0YWlsZWQgbm8tZG9uZSBzeW1yZWY9SEVBRDpyZWZzL2hlYWRzL21hc3RlciBhZ2VudD1naXQvMjoyLjQuOH5kYnVzc2luay1maXgtZW50ZXJwcmlzZS10b2tlbnMtY29tcGlsYXRpb24tMTE2Ny1nYzcwMDZjZgowMDNmZThkM2ZmYWI1NTI4OTVjMTliOWZjZjdhYTI2NGQyNzdjZGUzMzg4MSByZWZzL2hlYWRzL2JyYW5jaAowMDNmNmVjZjBlZjJjMmRmZmI3OTYwMzNlNWEwMjIxOWFmODZlYzY1ODRlNSByZWZzL2hlYWRzL21hc3RlcgowMDNlYjhlNDcxZjU4YmNiY2E2M2IwN2JkYTIwZTQyODE5MDQwOWMyZGI0NyByZWZzL3B1bGwvMS9oZWFkCjAwMDA="

func (s *SuiteCommon) TestGitUploadPackInfo(c *C) {
	b, _ := base64.StdEncoding.DecodeString(GitUploadPackInfoFixture)

	i := NewGitUploadPackInfo()
	err := i.Decode(pktline.NewDecoder(bytes.NewBuffer(b)))
	c.Assert(err, IsNil)

	ref := i.Capabilities.SymbolicReference("HEAD")
	c.Assert(ref, Equals, "refs/heads/master")
	c.Assert(i.Refs[ref].String(), Equals, "6ecf0ef2c2dffb796033e5a02219af86ec6584e5")
	c.Assert(i.Head.String(), Equals, "6ecf0ef2c2dffb796033e5a02219af86ec6584e5")
}

func (s *SuiteCommon) TestGitUploadPackInfoEmpty(c *C) {
	b := bytes.NewBuffer(nil)

	i := NewGitUploadPackInfo()
	err := i.Decode(pktline.NewDecoder(b))
	c.Assert(err, ErrorMatches, "permanent.*empty.*")
}

func (s *SuiteCommon) TestCapabilitiesDecode(c *C) {
	cap := NewCapabilities()
	cap.Decode("symref=foo symref=qux thin-pack")

	c.Assert(cap.m, HasLen, 2)
	c.Assert(cap.Get("symref").Values, DeepEquals, []string{"foo", "qux"})
	c.Assert(cap.Get("thin-pack").Values, DeepEquals, []string{""})
}

func (s *SuiteCommon) TestCapabilitiesSet(c *C) {
	cap := NewCapabilities()
	cap.Add("symref", "foo", "qux")
	cap.Set("symref", "bar")

	c.Assert(cap.m, HasLen, 1)
	c.Assert(cap.Get("symref").Values, DeepEquals, []string{"bar"})
}

func (s *SuiteCommon) TestCapabilitiesSetEmpty(c *C) {
	cap := NewCapabilities()
	cap.Set("foo", "bar")

	c.Assert(cap.Get("foo").Values, HasLen, 1)
}

func (s *SuiteCommon) TestCapabilitiesAdd(c *C) {
	cap := NewCapabilities()
	cap.Add("symref", "foo", "qux")
	cap.Add("thin-pack")

	c.Assert(cap.String(), Equals, "symref=foo symref=qux thin-pack")
}

func (s *SuiteCommon) TestGitUploadPackEncode(c *C) {
	info := NewGitUploadPackInfo()
	info.Capabilities.Add("symref", "HEAD:refs/heads/master")

	info.Head = core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5")
	info.Refs = map[string]core.Hash{
		"refs/heads/master": info.Head,
	}

	c.Assert(info.String(), Equals,
		"001e# service=git-upload-pack\n"+
			"000000506ecf0ef2c2dffb796033e5a02219af86ec6584e5 HEAD\x00symref=HEAD:refs/heads/master\n"+
			"003f6ecf0ef2c2dffb796033e5a02219af86ec6584e5 refs/heads/master\n"+
			"0000",
	)
}

func (s *SuiteCommon) TestGitUploadPackRequest(c *C) {
	r := &GitUploadPackRequest{}
	r.Want(core.NewHash("d82f291cde9987322c8a0c81a325e1ba6159684c"))
	r.Want(core.NewHash("2b41ef280fdb67a9b250678686a0c3e03b0a9989"))
	r.Have(core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	c.Assert(r.String(), Equals,
		"0032want d82f291cde9987322c8a0c81a325e1ba6159684c\n"+
			"0032want 2b41ef280fdb67a9b250678686a0c3e03b0a9989\n"+
			"0032have 6ecf0ef2c2dffb796033e5a02219af86ec6584e5\n0000"+
			"0009done\n",
	)
}
