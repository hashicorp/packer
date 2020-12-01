// plugin-check is a command used by plugins to validate compatibility and basic configuration
// to work with Packer.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const pluginCheck = "plugin-check"

var (
	hcl2spec = flag.Bool("hcl2spec", false, "flag to indicate that hcl2spec files should be checked.")
	website  = flag.Bool("website", false, "flag to indicate that website files should be checked.")
	load     = flag.Bool("load", false, "flag to indicate that the binary should be loaded to check compatibility.")
	binary   = flag.String("binary", "", "the binary name that will be loaded.")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of "+pluginCheck+":\n")
	fmt.Fprintf(os.Stderr, "\t"+pluginCheck+" [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(pluginCheck + ": ")
	flag.Usage = Usage
	flag.Parse()

	if *hcl2spec == false && *website == false && *load == false {
		flag.Usage()
		os.Exit(2)
	}

	if *hcl2spec {
		_ = checkHCL2Specs()
	}

	if *website {
		_ = checkWebsite()
	}

	if *load {
		if len(*binary) == 0 {
			flag.Usage()
			os.Exit(2)
		}

		_ = loadBinary()
	}
}

// checkHCL2Specs looks for the presence of a hcl2spec.go file in the current directory.
// It is not possible to predict the number of hcl2spec.go files for a given plugin.
// Because of that, finding one file is enough to validate the knowledge of hcl2spec generation.
func checkHCL2Specs() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var hcl2found bool
	_ = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if info.Name() == "website" || info.Name() == ".github" {
				return filepath.SkipDir
			}
		} else {
			if strings.HasSuffix(path, "hcl2spec.go") {
				hcl2found = true
				return io.EOF
			}
		}
		return nil
	})

	if hcl2found {
		fmt.Printf("a hcl2spec file was found \n")
		return nil
	}
	return fmt.Errorf("No hcl2spec.go files found. Please, make sure to generate them before releasing.")
}

// checkWebsite looks for the presence of a website folder with mdx files inside.
// It is not possible to predict the number of mdx files for a given plugin.
// Because of that, finding one file inside de folder is enough to validate the knowledge of website generation.
func checkWebsite() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	websiteDir := wd + "/website"
	stat, err := os.Stat(websiteDir)
	if err != nil {
		return fmt.Errorf("could not find website folter: %s", err.Error())
	}
	if !stat.IsDir() {
		return fmt.Errorf("expecting website do be a directory of mdx files")
	}

	var mdxFound bool
	_ = filepath.Walk(websiteDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".mdx" {
			mdxFound = true
			return io.EOF
		}
		return nil
	})

	if mdxFound {
		fmt.Printf("a mdx file was found inside website folder\n")
		return nil
	}
	return fmt.Errorf("No website files found. Please, make sure to generate them before releasing.")
}

func loadBinary() error {
	//TODO
	return nil
}
