package getter

import (
	"context"
	"net/url"
)

// SmbGetter is a Getter implementation that will download a module from
// a shared folder using samba scheme.
type SmbGetter struct {
	getter
}

func (g *SmbGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	path := "//" + u.Host + u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}
	return mode(path)
}

func (g *SmbGetter) Get(ctx context.Context, req *Request) error {
	path := "//" + req.u.Host + req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}
	return get(path, req)
}

func (g *SmbGetter) GetFile(ctx context.Context, req *Request) error {
	path := "//" + req.u.Host + req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}
	return getFile(path, req, ctx)
}
