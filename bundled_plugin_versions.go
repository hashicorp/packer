package main

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/packer/packer"
	"golang.org/x/mod/modfile"
)

//go:embed go.mod
var mod string

var pluginRegex = regexp.MustCompile("packer-plugin-.*$")

func GetBundledPluginVersions() map[string]packer.PluginSpec {
	pluginSpecs := map[string]packer.PluginSpec{}

	mods, err := modfile.Parse("", []byte(mod), nil)
	if err != nil {
		panic(fmt.Sprintf("failed to parse embedded modfile: %s", err))
	}

	for _, req := range mods.Require {
		if pluginRegex.MatchString(req.Mod.Path) {
			pluginName := pluginRegex.FindString(req.Mod.Path)
			pluginShortName := strings.Replace(pluginName, "packer-plugin-", "", 1)
			pluginSpecs[pluginShortName] = packer.PluginSpec{
				Name:    pluginShortName,
				Version: fmt.Sprintf("bundled (%s)", req.Mod.Version),
				Path:    os.Args[0],
			}
		}
	}

	return pluginSpecs
}
