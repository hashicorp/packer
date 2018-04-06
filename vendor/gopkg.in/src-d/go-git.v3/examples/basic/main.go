package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/src-d/go-git.v3"
)

func main() {
	fmt.Printf("Retrieving %q ...\n", os.Args[1])
	r, err := git.NewRepository(os.Args[1], nil)
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
}
