// This is the main package for the `packer` application.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("%#v\n", os.Args)
	fmt.Println("Hello, world.")
}
