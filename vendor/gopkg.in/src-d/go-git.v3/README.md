# go-git [![GoDoc](https://godoc.org/gopkg.in/src-d/go-git.v3?status.svg)](https://godoc.org/gopkg.in/src-d/go-git.v3) [![Build Status](https://travis-ci.org/src-d/go-git.svg)](https://travis-ci.org/src-d/go-git) [![codecov.io](https://codecov.io/github/src-d/go-git/coverage.svg)](https://codecov.io/github/src-d/go-git) [![codebeat badge](https://codebeat.co/badges/b6cb2f73-9e54-483d-89f9-4b95a911f40c)](https://codebeat.co/projects/github-com-src-d-go-git)

A low level and highly extensible git client library for **reading** repositories from git servers.  It is written in Go from scratch, without any C dependencies.

We have been following the open/close principle in its design to facilitate extensions.

*go-git* does not claim to be a replacement of [git2go](https://github.com/libgit2/git2go) as its approach and functionality is quite different.

### ok, but why? ...

At [source{d}](http://sourced.tech) we analyze almost **all** the public open source contributions made to git repositories in the world.

We want to extract detailed information from each GitHub repository, which requires downloading repository packfiles and analyzing them: extracting their code, authors, dates and the languages and ecosystems they use.  We are also interested in knowing who contributes to what, so we can tell top contributors from the more casual ones.

You can obtain all this information using the standard `git` command running over a local clone of a repository, but this simple solution does not scale well over millions of repositories: we want to avoid having local copies of the unpacked repositories in a regular file system; *go-git* allows us to work with an in-memory representation of repositories instead.

### I see... but this is production ready?

*Yes!!!*, we have been using *go-git* at [source{d}](http://sourced.tech) since August 2015 to analyze all GitHub public repositories (i.e. 16M of repositories).

### Coming Soon

Blame support: right now we are using a forward version of a line-tracking
algorithm and we are having some problems handling merges. The plan is to get
merges right and change to a backward line-tracking algorithm soon.

Installation
------------

The recommended way to install *go-git* is:

```
go get -u gopkg.in/src-d/go-git.v3/...
```


Examples
--------

Retrieving the commits for a given repository:

```go
r, err := git.NewRepository("https://github.com/src-d/go-git", nil)
if err != nil {
	panic(err)
}

if err := r.PullDefault(); err != nil {
	panic(err)
}

iter := r.Commits()
defer iter.Close()

for {
	//the commits are not shorted in any special order
	commit, err := iter.Next()
	if err != nil {
		if err == io.EOF {
			break
		}

		panic(err)
	}

	fmt.Println(commit)
}
```

Outputs:
```
commit 2275fa7d0c75d20103f90b0e1616937d5a9fc5e6
Author: Máximo Cuadros <mcuadros@gmail.com>
Date:   2015-10-23 00:44:33 +0200 +0200

commit 35b585759cbf29f8ec428ef89da20705d59f99ec
Author: Carlos Cobo <toqueteos@gmail.com>
Date:   2015-05-20 15:21:37 +0200 +0200

commit 7e3259c191a9de23d88b6077dcb1cd427e925432
Author: Alberto Cortés <alberto@sourced.tech>
Date:   2016-01-21 03:29:57 +0100 +0100

commit 24b8ae50db91f3909b11304014564bffc6fdee79
Author: Alberto Cortés <alberto@sourced.tech>
Date:   2015-12-11 17:57:10 +0100 +0100
...
```

Retrieving the latest commit for a given repository:

```go
r, err := git.NewRepository("https://github.com/src-d/go-git", nil)
if err != nil {
	panic(err)
}

if err := r.PullDefault(); err != nil {
	panic(err)
}

hash, err := r.Remotes[git.DefaultRemoteName].Head()
if err != nil {
	panic(err)
}

commit, err := r.Commit(hash)
if err != nil {
	panic(err)
}

fmt.Println(commit)
```


Acknowledgements
----------------

The earlier versions of the [packfile reader](https://godoc.org/gopkg.in/src-d/go-git.v3/formats/packfile) are based on [git-chain](https://github.com/gitchain/gitchain/blob/master/git/pack.go), project done by [@yrashk](https://github.com/yrashk)


License
-------

MIT, see [LICENSE](LICENSE)
