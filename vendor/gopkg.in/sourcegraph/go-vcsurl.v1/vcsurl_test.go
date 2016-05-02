package vcsurl

import (
	"github.com/kr/pretty"
	"testing"
)

var (
	githubUserRepo = RepoInfo{
		CloneURL: "git://github.com/user/repo.git",
		VCS:      Git,
		RepoHost: GitHub,
		Username: "user",
		Name:     "repo",
		FullName: "user/repo",
		Rev:      "asdf",
	}
	googleCodeRepo = RepoInfo{
		CloneURL: "https://code.google.com/p/go",
		VCS:      Mercurial,
		RepoHost: GoogleCode,
		Name:     "go",
		FullName: "go",
	}
	cpythonRepo = RepoInfo{
		CloneURL: "http://hg.python.org/cpython",
		VCS:      Mercurial,
		RepoHost: PythonOrg,
		Name:     "cpython",
		FullName: "cpython",
	}
	bitbucketHgRepo = RepoInfo{
		CloneURL: "https://bitbucket.org/user/repo",
		VCS:      Mercurial,
		RepoHost: Bitbucket,
		Username: "user",
		Name:     "repo",
		FullName: "user/repo",
	}
	bitbucketGitRepo = RepoInfo{
		CloneURL: "https://bitbucket.org/user/repo.git",
		VCS:      Git,
		RepoHost: Bitbucket,
		Username: "user",
		Name:     "repo",
		FullName: "user/repo",
	}
	launchpadRepo = RepoInfo{
		CloneURL: "bzr://launchpad.net/repo",
		VCS:      Bazaar,
		RepoHost: Launchpad,
		Username: "",
		Name:     "repo",
		FullName: "repo",
	}
)

func TestParse(t *testing.T) {
	tests := []struct {
		url  string
		rid  string
		info RepoInfo
	}{
		{"github.com/user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"http://github.com/user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"http://github.com/user/repo.git#asdf", "github.com/user/repo", githubUserRepo},
		{"https://github.com/user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"https://github.com/user/repo.git#asdf", "github.com/user/repo", githubUserRepo},
		{"git://github.com/user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"git://github.com/user/repo.git#asdf", "github.com/user/repo", githubUserRepo},
		{"git+ssh://github.com/user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"git+ssh://github.com/user/repo.git#asdf", "github.com/user/repo", githubUserRepo},
		{"git@github.com:user/repo#asdf", "github.com/user/repo", githubUserRepo},
		{"git@github.com:user/repo.git#asdf", "github.com/user/repo", githubUserRepo},

		{"code.google.com/p/go", "code.google.com/p/go", googleCodeRepo},
		{"https://code.google.com/p/go", "code.google.com/p/go", googleCodeRepo},

		{"hg.python.org/cpython", "hg.python.org/cpython", cpythonRepo},
		{"http://hg.python.org/cpython", "hg.python.org/cpython", cpythonRepo},

		{"bitbucket.org/user/repo", "bitbucket.org/user/repo", bitbucketHgRepo},
		{"https://bitbucket.org/user/repo", "bitbucket.org/user/repo", bitbucketHgRepo},
		{"http://bitbucket.org/user/repo", "bitbucket.org/user/repo", bitbucketHgRepo},

		{"bitbucket.org/user/repo.git", "bitbucket.org/user/repo", bitbucketGitRepo},
		{"https://bitbucket.org/user/repo.git", "bitbucket.org/user/repo", bitbucketGitRepo},
		{"http://bitbucket.org/user/repo.git", "bitbucket.org/user/repo", bitbucketGitRepo},

		{"http://launchpad.net/repo", "launchpad.net/repo", launchpadRepo},
		{"bzr://launchpad.net/repo", "launchpad.net/repo", launchpadRepo},
		{"bzr+ssh://launchpad.net/repo", "launchpad.net/repo", launchpadRepo},

		// subpaths
		{"http://github.com/user/repo/subpath#asdf", "github.com/user/repo", githubUserRepo},
		{"git@github.com:user/repo.git/subpath#asdf", "github.com/user/repo", githubUserRepo},
		{"https://code.google.com/p/go/subpath", "code.google.com/p/go", googleCodeRepo},

		// other repo hosts
		{"git://example.com/foo", "example.com/foo", RepoInfo{
			CloneURL: "git://example.com/foo",
			VCS:      Git,
			RepoHost: "example.com",
			Name:     "foo",
			FullName: "foo",
		}},
		{"https://example.com/foo.git", "example.com/foo", RepoInfo{
			CloneURL: "https://example.com/foo.git",
			VCS:      Git,
			RepoHost: "example.com",
			Name:     "foo",
			FullName: "foo",
		}},
		{"https://example.com/git/foo", "example.com/foo", RepoInfo{
			CloneURL: "https://example.com/git/foo",
			VCS:      Git,
			RepoHost: "example.com",
			Name:     "foo",
			FullName: "git/foo",
		}},
	}

	for _, test := range tests {
		info, err := Parse(test.url)
		if err != nil {
			t.Errorf("clone URL %q: got error: %s", test.url, err)
			continue
		}
		if test.info != *info {
			t.Errorf("%s: %v", test.url, pretty.Diff(test.info, *info))
		}
	}
}

func TestLink(t *testing.T) {
	tests := []struct {
		repo RepoInfo
		link string
	}{
		{githubUserRepo, "https://github.com/user/repo"},
		{bitbucketHgRepo, "https://bitbucket.org/user/repo"},
		{bitbucketGitRepo, "https://bitbucket.org/user/repo"},
		{googleCodeRepo, "https://code.google.com/p/go"},
	}

	for _, test := range tests {
		link := test.repo.Link()
		if test.link != link {
			t.Errorf("%s: want link %q, got %q", test.repo.CloneURL, test.link, link)
		}
	}
}
