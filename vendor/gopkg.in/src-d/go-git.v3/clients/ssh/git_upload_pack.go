// Package ssh implements a ssh client for go-git.
//
// The Connect() method is not allowed in ssh, use ConnectWithAuth() instead.
package ssh

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/formats/pktline"

	"github.com/sourcegraph/go-vcsurl"
	"golang.org/x/crypto/ssh"
)

// New errors introduced by this package.
var (
	ErrInvalidAuthMethod      = errors.New("invalid ssh auth method")
	ErrAuthRequired           = errors.New("cannot connect: auth required")
	ErrNotConnected           = errors.New("not connected")
	ErrAlreadyConnected       = errors.New("already connected")
	ErrUploadPackAnswerFormat = errors.New("git-upload-pack bad answer format")
	ErrUnsupportedVCS         = errors.New("only git is supported")
	ErrUnsupportedRepo        = errors.New("only github.com is supported")
)

// GitUploadPackService holds the service information.
// The zero value is safe to use.
// TODO: remove NewGitUploadPackService().
type GitUploadPackService struct {
	connected bool
	vcs       *vcsurl.RepoInfo
	client    *ssh.Client
	auth      AuthMethod
}

// NewGitUploadPackService initialises a GitUploadPackService.
// TODO: remove this, as the struct is zero-value safe.
func NewGitUploadPackService() *GitUploadPackService {
	return &GitUploadPackService{}
}

// Connect cannot be used with SSH clients and always return
// ErrAuthRequired. Use ConnectWithAuth instead.
func (s *GitUploadPackService) Connect(ep common.Endpoint) (err error) {
	return ErrAuthRequired
}

// ConnectWithAuth connects to ep using SSH. Authentication is handled
// by auth.
func (s *GitUploadPackService) ConnectWithAuth(ep common.Endpoint, auth common.AuthMethod) (err error) {
	if s.connected {
		return ErrAlreadyConnected
	}

	s.vcs, err = vcsurl.Parse(string(ep))
	if err != nil {
		return err
	}

	url, err := vcsToURL(s.vcs)
	if err != nil {
		return
	}

	var ok bool
	s.auth, ok = auth.(AuthMethod)
	if !ok {
		return ErrInvalidAuthMethod
	}

	s.client, err = ssh.Dial("tcp", url.Host, s.auth.clientConfig())
	if err != nil {
		return err
	}

	s.connected = true
	return
}

func vcsToURL(vcs *vcsurl.RepoInfo) (u *url.URL, err error) {
	if vcs.VCS != vcsurl.Git {
		return nil, ErrUnsupportedVCS
	}
	if vcs.RepoHost != vcsurl.GitHub {
		return nil, ErrUnsupportedRepo
	}
	s := "ssh://git@" + string(vcs.RepoHost) + ":22/" + vcs.FullName
	u, err = url.Parse(s)
	return
}

// Info returns the GitUploadPackInfo of the repository.
// The client must be connected with the repository (using
// the ConnectWithAuth() method) before using this
// method.
func (s *GitUploadPackService) Info() (i *common.GitUploadPackInfo, err error) {
	if !s.connected {
		return nil, ErrNotConnected
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		// the session can be closed by the other endpoint,
		// therefore we must ignore a close error.
		_ = session.Close()
	}()

	out, err := session.Output("git-upload-pack " + s.vcs.FullName + ".git")
	if err != nil {
		return nil, err
	}

	i = common.NewGitUploadPackInfo()
	return i, i.Decode(pktline.NewDecoder(bytes.NewReader(out)))
}

// Disconnect the SSH client.
func (s *GitUploadPackService) Disconnect() (err error) {
	if !s.connected {
		return ErrNotConnected
	}
	s.connected = false
	return s.client.Close()
}

// Fetch retrieves the GitUploadPack form the repository.
// You must be connected to the repository before using this method
// (using the ConnectWithAuth() method).
// TODO: fetch should really reuse the info session instead of openning a new
// one
func (s *GitUploadPackService) Fetch(r *common.GitUploadPackRequest) (rc io.ReadCloser, err error) {
	if !s.connected {
		return nil, ErrNotConnected
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		// the session can be closed by the other endpoint,
		// therefore we must ignore a close error.
		_ = session.Close()
	}()

	si, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	so, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		fmt.Fprintln(si, r.String())
		err = si.Close()
	}()

	err = session.Start("git-upload-pack " + s.vcs.FullName + ".git")
	if err != nil {
		return nil, err
	}
	// TODO: inestigate this *ExitError type (command fails or
	// doesn't complete successfully), as it is happenning all
	// the time, but everyting seems to work fine.
	err = session.Wait()
	if err != nil {
		if _, ok := err.(*ssh.ExitError); !ok {
			return nil, err
		}
	}

	// read until the header of the second answer
	soBuf := bufio.NewReader(so)
	token := "0000"
	for {
		var line string
		line, err = soBuf.ReadString('\n')
		if err == io.EOF {
			return nil, ErrUploadPackAnswerFormat
		}
		if line[0:len(token)] == token {
			break
		}
	}

	data, err := ioutil.ReadAll(soBuf)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	return ioutil.NopCloser(buf), nil
}
