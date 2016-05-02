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
