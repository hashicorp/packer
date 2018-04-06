=============================================
go-vcsurl - Lenient VCS repository URL parser
=============================================

[![xrefs](https://sourcegraph.com/api/repos/github.com/sourcegraph/go-vcsurl/badges/xrefs.png)](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl)
[![funcs](https://sourcegraph.com/api/repos/github.com/sourcegraph/go-vcsurl/badges/funcs.png)](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl)
[![top func](https://sourcegraph.com/api/repos/github.com/sourcegraph/go-vcsurl/badges/top-func.png)](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl)
[![library users](https://sourcegraph.com/api/repos/github.com/sourcegraph/go-vcsurl/badges/library-users.png)](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl)
[![status](https://sourcegraph.com/api/repos/github.com/sourcegraph/go-vcsurl/badges/status.png)](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl)

go-vcsurl parses VCS repository URLs in many common formats.

Note: the public API is experimental and subject to change until further notice.


Usage
=====

Documentation:
[go-vcsurl on Sourcegraph](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl).

Example: [example_test.go](https://github.com/sourcegraph/go-vcsurl/blob/master/example_test.go) ([Sourcegraph](https://sourcegraph.com/github.com/sourcegraph/go-vcsurl/tree/master/example_test.go)):

```go
package vcsurl_test

import (
	"fmt"
	"github.com/sourcegraph/go-vcsurl"
)

func ExampleParse() {
	urls := []string{
		"github.com/alice/libfoo",
		"git://github.com/bob/libbar",
		"code.google.com/p/libqux",
		"https://code.google.com/p/libbaz",
	}
	for i, url := range urls {
		if info, err := vcsurl.Parse(url); err == nil {
			fmt.Printf("%d. %s %s\n", i+1, info.VCS, info.CloneURL)
			fmt.Printf("   name: %s\n", info.Name)
			fmt.Printf("   host: %s\n", info.RepoHost)
		} else {
			fmt.Printf("error parsing %s\n")
		}
	}

	// output:
	// 1. git git://github.com/alice/libfoo.git
	//    name: libfoo
	//    host: github.com
	// 2. git git://github.com/bob/libbar.git
	//    name: libbar
	//    host: github.com
	// 3. hg https://code.google.com/p/libqux
	//    name: libqux
	//    host: code.google.com
	// 4. hg https://code.google.com/p/libbaz
	//    name: libbaz
	//    host: code.google.com
}
```


Running tests
=============

Run `go test`.


Contributors
============

* Quinn Slack <sqs@sourcegraph.com>
