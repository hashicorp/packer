package main

import (
	"fmt"
	"os"

	"gopkg.in/src-d/go-git.v3"
)

func main() {
	fmt.Printf("Retrieving latest commit from: %q ...\n", os.Args[1])
	r, err := git.NewRepository(os.Args[1], nil)
	if err != nil {
		panic(err)
	}

	if err := r.Pull(git.DefaultRemoteName, "refs/heads/master"); err != nil {
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
}
