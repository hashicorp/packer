package gcs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/go-getter/v2"
	"google.golang.org/api/iterator"
)

// Getter is a Getter implementation that will download a module from
// a GCS bucket.
type Getter struct {}

func (g *Getter) Mode(ctx context.Context, u *url.URL) (getter.Mode, error) {

	// Parse URL
	bucket, object, err := g.parseURL(u)
	if err != nil {
		return 0, err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return 0, err
	}
	iter := client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})
	for {
		obj, err := iter.Next()
		if err != nil && err != iterator.Done {
			return 0, err
		}

		if err == iterator.Done {
			break
		}
		if strings.HasSuffix(obj.Name, "/") {
			// A directory matched the prefix search, so this must be a directory
			return getter.ModeDir, nil
		} else if obj.Name != object {
			// A file matched the prefix search and doesn't have the same name
			// as the query, so this must be a directory
			return getter.ModeDir, nil
		}
	}
	// There are no directories or subdirectories, and if a match was returned,
	// it was exactly equal to the prefix search. So return File mode
	return getter.ModeFile, nil
}

func (g *Getter) Get(ctx context.Context, req *getter.Request) error {
	// Parse URL
	bucket, object, err := g.parseURL(req.URL())
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Iterate through all matching objects.
	iter := client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})
	for {
		obj, err := iter.Next()
		if err != nil && err != iterator.Done {
			return err
		}
		if err == iterator.Done {
			break
		}

		if !strings.HasSuffix(obj.Name, "/") {
			// Get the object destination path
			objDst, err := filepath.Rel(object, obj.Name)
			if err != nil {
				return err
			}
			objDst = filepath.Join(req.Dst, objDst)
			// Download the matching object.
			err = g.getObject(ctx, client, objDst, bucket, obj.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Getter) GetFile(ctx context.Context, req *getter.Request) error {
	// Parse URL
	bucket, object, err := g.parseURL(req.URL())
	if err != nil {
		return err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	return g.getObject(ctx, client, req.Dst, bucket, object)
}

func (g *Getter) getObject(ctx context.Context, client *storage.Client, dst, bucket, object string) error {
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = getter.Copy(ctx, f, rc)
	return err
}

func (g *Getter) parseURL(u *url.URL) (bucket, path string, err error) {
	if strings.Contains(u.Host, "googleapis.com") {
		hostParts := strings.Split(u.Host, ".")
		if len(hostParts) != 3 {
			err = fmt.Errorf("URL is not a valid GCS URL")
			return
		}

		pathParts := strings.SplitN(u.Path, "/", 5)
		if len(pathParts) != 5 {
			err = fmt.Errorf("URL is not a valid GCS URL")
			return
		}
		bucket = pathParts[3]
		path = pathParts[4]
	}
	return
}

func (g *Getter) Detect(req *getter.Request) (bool, error) {
	src := req.Src
	if len(src) == 0 {
		return false, nil
	}

	if req.Forced != "" {
		// There's a getter being forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the forced one
			// Don't use it to try to download the artifact
			return false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return false, nil
	}

	src, ok, err := new(Detector).Detect(src, req.Pwd)
	if err != nil {
		return ok, err
	}
	if ok {
		req.Src = src
		return ok, nil
	}

	return false, nil
}

func (g *Getter) validScheme(scheme string) bool {
	return scheme == "gcs"
}
