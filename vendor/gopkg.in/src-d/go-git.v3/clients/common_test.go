package clients

import (
	"fmt"
	"io"
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/clients/common"
)

func Test(t *testing.T) { TestingT(t) }

type SuiteCommon struct{}

var _ = Suite(&SuiteCommon{})

func (s *SuiteCommon) TestNewGitUploadPackService(c *C) {
	var tests = [...]struct {
		input    string
		err      bool
		expected string
	}{
		{"://example.com", true, "<nil>"},
		{"badscheme://github.com/src-d/go-git", true, "<nil>"},
		{"http://github.com/src-d/go-git", false, "*http.GitUploadPackService"},
		{"https://github.com/src-d/go-git", false, "*http.GitUploadPackService"},
		{"ssh://github.com/src-d/go-git", false, "*ssh.GitUploadPackService"},
	}

	for i, t := range tests {
		output, err := NewGitUploadPackService(t.input)
		c.Assert(err != nil, Equals, t.err, Commentf("%d) %q: wrong error value", i, t.input))
		c.Assert(typeAsString(output), Equals, t.expected, Commentf("%d) %q: wrong type", i, t.input))
	}
}

type dummyProtocolService struct{}

func newDummyProtocolService() common.GitUploadPackService {
	return &dummyProtocolService{}
}

func (s *dummyProtocolService) Connect(url common.Endpoint) error {
	return nil
}

func (s *dummyProtocolService) ConnectWithAuth(url common.Endpoint, auth common.AuthMethod) error {
	return nil
}

func (s *dummyProtocolService) Info() (*common.GitUploadPackInfo, error) {
	return nil, nil
}

func (s *dummyProtocolService) Fetch(r *common.GitUploadPackRequest) (io.ReadCloser, error) {
	return nil, nil
}

func (s *SuiteCommon) TestInstallProtocol(c *C) {
	var tests = [...]struct {
		scheme  string
		service common.GitUploadPackService
		panic   bool
	}{
		{"panic", nil, true},
		{"newscheme", newDummyProtocolService(), false},
		{"http", newDummyProtocolService(), false},
	}

	for i, t := range tests {
		if t.panic {
			fmt.Println(t.service == nil)
			c.Assert(func() { InstallProtocol(t.scheme, t.service) }, PanicMatches, `nil service`)
			continue
		}

		InstallProtocol(t.scheme, t.service)
		c.Assert(typeAsString(KnownProtocols[t.scheme]), Equals, typeAsString(t.service), Commentf("%d) wrong service", i))
		// reset to default protocols after installing
		if v, ok := DefaultProtocols[t.scheme]; ok {
			InstallProtocol(t.scheme, v)
		}
	}
}

func typeAsString(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
