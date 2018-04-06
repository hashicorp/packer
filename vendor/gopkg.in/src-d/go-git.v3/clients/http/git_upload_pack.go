package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/formats/pktline"
)

type GitUploadPackService struct {
	Client *http.Client

	endpoint common.Endpoint
	auth     HTTPAuthMethod
}

func NewGitUploadPackService() *GitUploadPackService {
	return &GitUploadPackService{
		Client: http.DefaultClient,
	}
}

func (s *GitUploadPackService) Connect(url common.Endpoint) error {
	s.endpoint = url

	return nil
}

func (s *GitUploadPackService) ConnectWithAuth(url common.Endpoint, auth common.AuthMethod) error {
	httpAuth, ok := auth.(HTTPAuthMethod)
	if !ok {
		return InvalidAuthMethodErr
	}

	s.endpoint = url
	s.auth = httpAuth

	return nil
}

func (s *GitUploadPackService) Info() (*common.GitUploadPackInfo, error) {
	url := fmt.Sprintf("%s/info/refs?service=%s", s.endpoint, common.GitUploadPackServiceName)
	res, err := s.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	i := common.NewGitUploadPackInfo()
	return i, i.Decode(pktline.NewDecoder(res.Body))
}

func (s *GitUploadPackService) Fetch(r *common.GitUploadPackRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s", s.endpoint, common.GitUploadPackServiceName)
	res, err := s.doRequest("POST", url, r.Reader())
	if err != nil {
		return nil, err
	}

	h := make([]byte, 8)
	if _, err := res.Body.Read(h); err != nil {
		return nil, core.NewUnexpectedError(err)
	}

	return res.Body, nil
}

func (s *GitUploadPackService) doRequest(method, url string, content *strings.Reader) (*http.Response, error) {
	var body io.Reader
	if content != nil {
		body = content
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, core.NewPermanentError(err)
	}

	s.applyHeadersToRequest(req, content)
	s.applyAuthToRequest(req)

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, core.NewUnexpectedError(err)
	}

	if err := NewHTTPError(res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *GitUploadPackService) applyHeadersToRequest(req *http.Request, content *strings.Reader) {
	req.Header.Add("User-Agent", "git/1.0")
	req.Header.Add("Host", "github.com")

	if content == nil {
		req.Header.Add("Accept", "*/*")
	} else {
		req.Header.Add("Accept", "application/x-git-upload-pack-result")
		req.Header.Add("Content-Type", "application/x-git-upload-pack-request")
		req.Header.Add("Content-Length", string(content.Len()))
	}
}

func (s *GitUploadPackService) applyAuthToRequest(req *http.Request) {
	if s.auth == nil {
		return
	}

	s.auth.setAuth(req)
}
