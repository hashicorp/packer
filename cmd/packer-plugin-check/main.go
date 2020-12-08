// packer-plugin-check is a command used by plugins to validate compatibility and basic configuration
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

	"github.com/hashicorp/packer/packer/plugin"
)

const packerPluginCheck = "packer-plugin-check"

var (
	hcl2spec = flag.Bool("hcl2spec", false, "flag to indicate that hcl2spec files should be checked.")
	docs     = flag.Bool("docs", false, "flag to indicate that documentation files should be checked.")
	load     = flag.String("load", "", "flag to check if plugin can be loaded by Packer.")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of "+packerPluginCheck+":\n")
	fmt.Fprintf(os.Stderr, "\t"+packerPluginCheck+" [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(packerPluginCheck + ": ")
	flag.Usage = Usage
	flag.Parse()

	if flag. NFlag() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if *hcl2spec {
		if err := checkHCL2Specs(); err != nil {
			fmt.Printf(err.Error())
			os.Exit(2)
		}
		fmt.Printf("Plugin succesfully passed hcl2spec check.\n")
	}

	if *docs {
		if err := checkDocumentation(); err != nil {
			fmt.Printf(err.Error())
			os.Exit(2)
		}
		fmt.Printf("Plugin succesfully passed docs check.\n")
	}

	if len(*load) > 0 {
		if err := checkPluginName(*load); err != nil {
			fmt.Printf(err.Error())
			os.Exit(2)
		}
		if err := discoverAndLoad(); err != nil {
			fmt.Printf(err.Error())
			os.Exit(2)
		}
		fmt.Printf("Plugin succesfully passed compatibility check.\n")
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
			if info.Name() == "docs" || info.Name() == ".github" {
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
		return nil
	}
	return fmt.Errorf("no hcl2spec.go files found, make sure to generate them before releasing")
}

// checkDocumentation looks for the presence of a docs folder with mdx files inside.
// It is not possible to predict the number of mdx files for a given plugin.
// Because of that, finding one file inside de folder is enough to validate the knowledge of docs generation.
func checkDocumentation() error {
	// TODO: this should be updated once we have defined what's going to be for plguin's docs
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	docsDir := wd + "/docs"
	stat, err := os.Stat(docsDir)
	if err != nil {
		return fmt.Errorf("could not find docs folter: %s", err.Error())
	}
	if !stat.IsDir() {
		return fmt.Errorf("expecting docs do be a directory of mdx files")
	}

	var mdxFound bool
	_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".mdx" {
			mdxFound = true
			return io.EOF
		}
		return nil
	})

	if mdxFound {
		fmt.Printf("a mdx file was found inside the docs folder\n")
		return nil
	}
	return fmt.Errorf("no docs files found, make sure to have the docs in place before releasing")
}

// checkPluginName checks for the possible valid names for a plugin, packer-plugin-* or packer-[builder|provisioner|post-processor]-*.
// If the name is prefixed with `packer-[builder|provisioner|post-processor]-`, packer won't be able to install it,
// therefore a WARNING will be shown.
func checkPluginName(name string) error {
	if strings.HasPrefix(name, "packer-plugin-") {
		return nil
	}
	if strings.HasPrefix(name, "packer-builder-") ||
		strings.HasPrefix(name, "packer-provisioner-") ||
		strings.HasPrefix(name, "packer-post-processor-") {
		fmt.Printf("\n[WARNING] Plugin is named with old prefix `packer-[builder|provisioner|post-processor]-{name})`. " +
			"These will be detected but Packer cannot install them automatically. " +
			"Update the name to packer-plugin-{name} and the plugin will become available through `packer init` command.\n")
		return nil
	}
	return fmt.Errorf("plugin's name is not valid")
}

// discoverAndLoad will discover the plugin binary from the current directory and load any builder/provisioner/post-processor
// in the plugin configuration. At least one builder, provisioner, or post-processor should be found to validate the plugin's
// compatibility with Packer.
func discoverAndLoad() error {
	config := plugin.Config{
		PluginMinPort: 10000,
		PluginMaxPort: 25000,
	}
	err := config.Discover()
	if err != nil {
		return err
	}

	// TODO: validate correctness of plugins loaded by checking them against the output of the `describe` command.
	builders, provisioners, postProcessors := config.GetPlugins()
	if len(builders) == 0 &&
		len(provisioners) == 0 &&
		len(postProcessors) == 0 {
		return fmt.Errorf("couldn't load any Builder/Provisioner/Post-Processor from the plugin binary")
	}

	return nil
}
