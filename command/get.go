package command

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"archive/zip"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	git "gopkg.in/src-d/go-git.v3"
	core "gopkg.in/src-d/go-git.v3/core"
)

// GetCommand is a Command implementation download and build template from remote location.
type GetCommand struct {
	Meta
}

func (c *GetCommand) Help() string {
	helpText := `
Usage: packer get [options] repository

  Get remote repository to local dir at specified revision if it present.
  Optional run build with specific template.

Options:

  -d=path                Destination dir to put files.
  -r=string              Remote location to fetch
  -k=false               Keep destination dir after build
  -f=false               Only fetch template
  -t=string              Template to build
  -s=int                 Strip components
`

	return strings.TrimSpace(helpText)
}

func (c *GetCommand) Run(args []string) int {
	var dest, remote, template string
	var keep, fetch bool
	var strip int
	var err error

	flags := c.Meta.FlagSet("get", FlagSetVars)
	flags.Usage = func() { c.Ui.Error(c.Help()) }
	flags.StringVar(&dest, "d", "", "d")
	flags.StringVar(&remote, "r", "", "r")
	flags.StringVar(&template, "t", "all", "t")
	flags.BoolVar(&fetch, "f", false, "f")
	flags.BoolVar(&keep, "k", false, "k")
	flags.IntVar(&strip, "s", 0, "s")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	if dest == "" {
		dest, err = ioutil.TempDir("", "packer-get-")
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
	} else {
		if _, err = os.Stat(dest); err == nil {
			fmt.Printf("err: destination dir must not exists")
			return 2
		}
		if err = os.MkdirAll(dest, os.FileMode(0755)); err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
	}

	u, err := url.Parse(remote)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return 2
	}
	switch u.Scheme {
	default:
		switch filepath.Ext(u.Path) {
		case ".git":
			err = getGit(remote, dest)
		default:
			err = fmt.Errorf("scheme %q not supported", u.Scheme)
		}
	case "http", "https":
		switch filepath.Ext(u.Path) {
		default:

			err = fmt.Errorf("scheme %q not supported", u.Scheme)
		case ".zip", ".tar":
			err = getArchive(remote, dest, strip)
		}
	case "git", "git+http", "git+https":
		err = getGit(remote, dest)
	}
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return 2
	}

	if !keep {
		defer os.RemoveAll(dest)
	}

	if !fetch {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
		err = os.Chdir(dest)
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			return 2
		}
		defer os.Chdir(cwd)
		b := &BuildCommand{c.Meta}
		if template == "all" {
			templates, err := filepath.Glob("*.json")
			if err != nil {
				fmt.Printf("err: %s\n", err.Error())
				return 2
			}
			for _, template = range templates {
				if ret := b.Run(append(flag.Args(), template)); ret != 0 {
					return ret
				}
			}
		} else {
			return b.Run(append(flag.Args(), template))
		}
	}
	return 0
}

func (c *GetCommand) Synopsis() string {
	return "Get template from remote location and built it"
}

func getArchive(src string, dst string, stripComponents int) error {
	tmp, err := ioutil.TempFile("", "packer-get")
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	httpTransport := &http.Transport{
		Dial:            (&net.Dialer{DualStack: true}).Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: httpTransport, Timeout: 30 * time.Second}

	res, err := httpClient.Get(src)
	if err != nil {
		return err
	} else if res.Body == nil {
		return fmt.Errorf("Empty response Body")
	}
	n, err := io.Copy(tmp, res.Body)
	res.Body.Close()
	tmp.Sync()
	if err != nil {
		return err
	}
	u, _ := url.Parse(src)
	switch filepath.Ext(u.Path) {
	case ".zip":
		cr, err := zip.NewReader(tmp, n)
		if err != nil {
			return err
		}
		for _, f := range cr.File {
			name := f.Name
			for idx := 0; idx < stripComponents; idx++ {
				for idxCh, ch := range name {
					if ch == filepath.Separator {
						name = name[idxCh+1:]
						break
					}
				}
			}
			if name == "." || name == "" {
				continue
			}
			if f.Mode().IsDir() {
				if err = os.MkdirAll(name, f.Mode()); err != nil {
					return err
				}
			} else {
				cf, err := f.Open()
				if err != nil {
					return err
				}
				path := filepath.Dir(filepath.Join(dst, name))
				if err = os.MkdirAll(path, os.FileMode(0755)); err != nil {
					return err
				}
				df, err := os.OpenFile(filepath.Join(path, filepath.Base(name)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode())
				if err != nil {
					return err
				}
				_, err = io.Copy(df, cf)
				if err != nil {
					df.Close()
					cf.Close()
					return err
				}
				df.Close()
				cf.Close()
			}
		}
	}
	return nil
}

func getGit(src string, dst string) error {
	var remote, ref string
	var hash core.Hash

	if strings.HasPrefix(src, "git+") {
		src = src[4:]
	}

	u, err := url.Parse(src)
	if err != nil {
		return err
	}

	if idx := strings.Index(u.Path, "@"); idx > 0 {
		if len(u.Path[idx+1:]) == 40 {
			hash = core.Hash(core.NewHash(u.Path[idx+1:]))
		} else {
			ref = u.Path[idx+1:]
		}
		u.Path = u.Path[:idx]
	}
	remote = u.String()

	repo, err := git.NewRepository(remote, nil)
	if err != nil {
		return err
	}

	if ref == "" && !hash.IsZero() {
		err = repo.Pull(git.DefaultRemoteName, "refs/heads/master")
	} else {
		err = repo.Pull(git.DefaultRemoteName, "refs/heads/"+ref)
	}
	if err != nil {
		return err
	}

	if hash.IsZero() {
		hash, err = repo.Remotes[git.DefaultRemoteName].Head()
		if err != nil {
			return err
		}
	}

	commit, err := repo.Commit(hash)
	if err != nil {
		return err
	}

	fiter := commit.Tree().Files()
	defer fiter.Close()

	for {
		f, err := fiter.Next()
		if err != nil {
			if err == io.EOF && f == nil {
				break
			}
			return err
		}
		path := filepath.Dir(filepath.Join(dst, f.Name))
		if err = os.MkdirAll(path, os.FileMode(0755)); err != nil {
			return err
		}
		r, err := f.Reader()
		if err != nil {
			return err
		}
		if f.Mode == 40960 {
			buf, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			dst := string(buf)
			err = os.Symlink(dst, filepath.Join(path, filepath.Base(f.Name)))
			if err != nil {
				return err
			}
		} else {
			fp, err := os.OpenFile(filepath.Join(path, filepath.Base(f.Name)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode)
			if err != nil {
				return err
			}
			_, err = io.Copy(fp, r)
			if err != nil {
				r.Close()
				fp.Close()
				return err
			}
			fp.Close()
		}
		r.Close()
	}
	return nil
}
