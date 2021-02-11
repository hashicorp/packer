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

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

const packerPluginsCheck = "packer-plugins-check"

var (
	docs = flag.Bool("docs", false, "flag to indicate that documentation files should be checked.")
	load = flag.String("load", "", "flag to check if plugin can be loaded by Packer and is compatible with HCL2.")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of "+packerPluginsCheck+":\n")
	fmt.Fprintf(os.Stderr, "\t"+packerPluginsCheck+" [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(packerPluginsCheck + ": ")
	flag.Usage = Usage
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(2)
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

// checkDocumentation looks for the presence of a docs folder with mdx files inside.
// It is not possible to predict the number of mdx files for a given plugin.
// Because of that, finding one file inside the folder is enough to validate the docs existence.
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
			"The plugin must be a multi-component plugin named packer-plugin-{name} to be installable through the `packer init` command.\n")
		return nil
	}
	return fmt.Errorf("plugin's name is not valid")
}

// discoverAndLoad will discover the plugin binary from the current directory and load any builder/provisioner/post-processor
// in the plugin configuration. At least one builder, provisioner, or post-processor should be found to validate the plugin's
// compatibility with Packer.
func discoverAndLoad() error {
	config := packer.PluginConfig{
		PluginMinPort: 10000,
		PluginMaxPort: 25000,
	}
	err := config.Discover()
	if err != nil {
		return err
	}

	// TODO: validate correctness of plugins loaded by checking them against the output of the `describe` command.
	if len(config.Builders.List()) == 0 &&
		len(config.Provisioners.List()) == 0 &&
		len(config.PostProcessors.List()) == 0 {
		return fmt.Errorf("couldn't load any Builder/Provisioner/Post-Processor from the plugin binary")
	}

	return checkHCL2ConfigSpec(config)
}

// checkHCL2ConfigSpec checks if the hcl2spec config is present for the given plugins by validating that ConfigSpec() does not
// return an empty map of specs.
func checkHCL2ConfigSpec(plugins packer.PluginConfig) error {
	var errs *packersdk.MultiError
	for _, b := range plugins.Builders.List() {
		builder, err := plugins.Builders.Start(b)
		if err != nil {
			return packersdk.MultiErrorAppend(err, errs)
		}
		if len(builder.ConfigSpec()) == 0 {
			errs = packersdk.MultiErrorAppend(fmt.Errorf("builder %q does not contain the required hcl2spec configuration", b), errs)
		}
	}
	for _, p := range plugins.Provisioners.List() {
		provisioner, err := plugins.Provisioners.Start(p)
		if err != nil {
			return packersdk.MultiErrorAppend(err, errs)
		}
		if len(provisioner.ConfigSpec()) == 0 {
			errs = packersdk.MultiErrorAppend(fmt.Errorf("provisioner %q does not contain the required hcl2spec configuration", p), errs)
		}
	}
	for _, pp := range plugins.PostProcessors.List() {
		postProcessor, err := plugins.PostProcessors.Start(pp)
		if err != nil {
			return packersdk.MultiErrorAppend(err, errs)
		}
		if len(postProcessor.ConfigSpec()) == 0 {
			errs = packersdk.MultiErrorAppend(fmt.Errorf("post-processor %q does not contain the required hcl2spec configuration", pp), errs)
		}
	}
	for _, d := range plugins.DataSources.List() {
		datasource, err := plugins.DataSources.Start(d)
		if err != nil {
			return packersdk.MultiErrorAppend(err, errs)
		}
		if len(datasource.ConfigSpec()) == 0 {
			errs = packersdk.MultiErrorAppend(fmt.Errorf("datasource %q does not contain the required hcl2spec configuration", d), errs)
		}
	}
	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}
