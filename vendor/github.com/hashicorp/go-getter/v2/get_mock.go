package getter

import (
	"context"
	"net/url"
)

// MockGetter is an implementation of Getter that can be used for tests.
type MockGetter struct {

	// Proxy, if set, will be called after recording the calls below.
	// If it isn't set, then the *Err values will be returned.
	Proxy Getter

	GetCalled bool
	GetDst    string
	GetURL    *url.URL
	GetErr    error

	GetFileCalled bool
	GetFileDst    string
	GetFileURL    *url.URL
	GetFileErr    error
}

func (g *MockGetter) Get(ctx context.Context, req *Request) error {
	g.GetCalled = true
	g.GetDst = req.Dst
	g.GetURL = req.u

	if g.Proxy != nil {
		return g.Proxy.Get(ctx, req)
	}

	return g.GetErr
}

func (g *MockGetter) GetFile(ctx context.Context, req *Request) error {
	g.GetFileCalled = true
	g.GetFileDst = req.Dst
	g.GetFileURL = req.u

	if g.Proxy != nil {
		return g.Proxy.GetFile(ctx, req)
	}
	return g.GetFileErr
}

func (g *MockGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if l := len(u.Path); l > 0 && u.Path[l-1:] == "/" {
		return ModeDir, nil
	}
	return ModeFile, nil
}

func (g *MockGetter) Detect(req *Request) (bool, error) {
	if g.Proxy != nil {
		return g.Proxy.Detect(req)
	}
	return true, nil
}
