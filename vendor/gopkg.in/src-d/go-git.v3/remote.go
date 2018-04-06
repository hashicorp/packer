package git

import (
	"fmt"
	"io"

	"gopkg.in/src-d/go-git.v3/clients"
	"gopkg.in/src-d/go-git.v3/clients/common"
	"gopkg.in/src-d/go-git.v3/core"
)

type Remote struct {
	Endpoint common.Endpoint
	Auth     common.AuthMethod

	upSrv  common.GitUploadPackService
	upInfo *common.GitUploadPackInfo
}

// NewRemote returns a new Remote, using as client http.DefaultClient
func NewRemote(url string) (*Remote, error) {
	return NewAuthenticatedRemote(url, nil)
}

// NewAuthenticatedRemote returns a new Remote using the given AuthMethod, using as
// client http.DefaultClient
func NewAuthenticatedRemote(url string, auth common.AuthMethod) (*Remote, error) {
	end, err := common.NewEndpoint(url)
	if err != nil {
		return nil, err
	}

	upSrv, err := clients.NewGitUploadPackService(url)
	if err != nil {
		return nil, err
	}
	return &Remote{
		Endpoint: end,
		Auth:     auth,
		upSrv:    upSrv,
	}, nil
}

// Connect with the endpoint
func (r *Remote) Connect() error {
	var err error
	if r.Auth == nil {
		err = r.upSrv.Connect(r.Endpoint)
	} else {
		err = r.upSrv.ConnectWithAuth(r.Endpoint, r.Auth)
	}

	if err != nil {
		return err
	}

	return r.retrieveUpInfo()
}

func (r *Remote) retrieveUpInfo() error {
	var err error
	if r.upInfo, err = r.upSrv.Info(); err != nil {
		return err
	}

	return nil
}

// Info returns the git-upload-pack info
func (r *Remote) Info() *common.GitUploadPackInfo {
	return r.upInfo
}

// Capabilities returns the remote capabilities
func (r *Remote) Capabilities() *common.Capabilities {
	return r.upInfo.Capabilities
}

// DefaultBranch returns the name of the remote's default branch
func (r *Remote) DefaultBranch() string {
	return r.upInfo.Capabilities.SymbolicReference("HEAD")
}

// Head returns the Hash of the HEAD
func (r *Remote) Head() (core.Hash, error) {
	return r.Ref(r.DefaultBranch())
}

// Fetch returns a reader using the request
func (r *Remote) Fetch(req *common.GitUploadPackRequest) (io.ReadCloser, error) {
	return r.upSrv.Fetch(req)
}

// FetchDefaultBranch returns a reader for the default branch
func (r *Remote) FetchDefaultBranch() (io.ReadCloser, error) {
	ref, err := r.Ref(r.DefaultBranch())
	if err != nil {
		return nil, err
	}

	req := &common.GitUploadPackRequest{}
	req.Want(ref)

	return r.Fetch(req)
}

// Ref returns the Hash pointing the given refName
func (r *Remote) Ref(refName string) (core.Hash, error) {
	ref, ok := r.upInfo.Refs[refName]
	if !ok {
		return core.NewHash(""), fmt.Errorf("unable to find ref %q", refName)
	}

	return ref, nil
}

// Refs returns the Hash pointing the given refName
func (r *Remote) Refs() map[string]core.Hash {
	return r.upInfo.Refs
}
