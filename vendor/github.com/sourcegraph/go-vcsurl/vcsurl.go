package vcsurl

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type RepoHost string

const (
	GitHub     RepoHost = "github.com"
	Bitbucket  RepoHost = "bitbucket.org"
	GoogleCode RepoHost = "code.google.com"
	PythonOrg  RepoHost = "hg.python.org"
	Launchpad  RepoHost = "launchpad.net"
)

type VCS string

const (
	Git       VCS = "git"
	Mercurial VCS = "hg"
	Bazaar    VCS = "bzr"
)

// RepoInfo describes a VCS repository.
type RepoInfo struct {
	CloneURL string   // clone URL
	VCS      VCS      // VCS type
	RepoHost RepoHost // repo hosting site
	Username string   // username of repo owner on repo hosting site
	Name     string   // base name of repo on repo hosting site
	FullName string   // full name of repo on repo hosting site
	Rev      string   // a specific revision (commit ID, branch, etc.)
}

// Link returns the URL to the repository that is intended for access by humans
// using a Web browser (i.e., not the URL to the API resource).
func (r *RepoInfo) Link() string {
	switch r.RepoHost {
	case GoogleCode:
		return fmt.Sprintf("https://code.google.com/p/%s", r.FullName)
	default:
		return (&url.URL{Scheme: "https", Host: string(r.RepoHost), Path: "/" + r.FullName}).String()
	}
}

var removeDotGit = regexp.MustCompile(`\.git$`)

// Parses a string that resembles a VCS repository URL. See TestParse for a list of supported URL
// formats.
func Parse(spec string) (info *RepoInfo, err error) {
	if strings.HasPrefix(spec, "git@github.com:") {
		spec = strings.Replace(spec, "git@github.com:", "git://github.com/", 1)
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(spec); err == nil {
		if parsedURL.Scheme == "" {
			spec = "https://" + spec
			if parsedURL, err = url.Parse(spec); err != nil {
				return nil, err
			}
		}

		info = new(RepoInfo)

		info.CloneURL = parsedURL.String()
		info.RepoHost = RepoHost(parsedURL.Host)
		info.Rev = parsedURL.Fragment

		if info.RepoHost == GitHub || parsedURL.Scheme == "git" {
			info.VCS = Git
		} else if info.RepoHost == GoogleCode && parsedURL.Scheme == "https" {
			info.VCS = Mercurial
		} else if info.RepoHost == Bitbucket && (parsedURL.Scheme == "https" || parsedURL.Scheme == "http") {
			if !strings.HasSuffix(parsedURL.Path, ".git") {
				info.VCS = Mercurial
			}
		} else if info.RepoHost == Launchpad {
			info.VCS = Bazaar
		}

		path := parsedURL.Path
		switch info.RepoHost {
		case GitHub:
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				info.Username = parts[1]
				parts[2] = removeDotGit.ReplaceAllLiteralString(parts[2], "")
				info.Name = parts[2]
				info.FullName = parts[1] + "/" + parts[2]
				info.CloneURL = "git://github.com/" + info.FullName + ".git"
			}
		case GoogleCode:
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				if parts[1] == "p" {
					info.Name = parts[2]
					info.FullName = info.Name
					info.CloneURL = "https://code.google.com/p/" + info.FullName
				}
			}
		case PythonOrg:
			parts := strings.Split(path, "/")
			if len(parts) >= 2 {
				info.CloneURL = "http://hg.python.org" + path
				info.VCS = Mercurial
				info.Name = parts[len(parts)-1]
				info.FullName = strings.Join(parts[1:], "/")
			}
		case Bitbucket:
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				info.Username = parts[1]
				if strings.HasSuffix(parts[2], ".git") {
					info.VCS = Git
					parts[2] = strings.TrimSuffix(parts[2], ".git")
				}
				info.Name = parts[2]
				info.FullName = parts[1] + "/" + parts[2]
				info.CloneURL = "https://bitbucket.org/" + info.FullName
				if info.VCS == Git {
					info.CloneURL += ".git"
				}
			}
		default:
			if len(path) == 0 {
				return nil, fmt.Errorf("empty path in repo spec: %q", spec)
			}
			path = path[1:] // remove leading slash
			path = removeDotGit.ReplaceAllLiteralString(path, "")
			info.FullName = path
			info.Name = filepath.Base(path)
			if strings.Contains(spec, "git") {
				info.VCS = Git
			} else if strings.Contains(spec, "hg") || strings.Contains(spec, "mercurial") {
				info.VCS = Mercurial
			}
		}

		if info.RepoHost == Launchpad {
			parsedURL.Scheme = "bzr"
			info.CloneURL = parsedURL.String()
		}

		if info.Name == "" || info.FullName == "" {
			return nil, fmt.Errorf("unable to determine name or full name for repo spec %q", spec)
		}

		return info, nil
	}
	return nil, err
}
