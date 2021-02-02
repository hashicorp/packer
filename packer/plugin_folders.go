package packer

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/pathing"
)

// PluginFolders returns the list of known plugin folders based on system.
func PluginFolders(dirs ...string) []string {
	res := []string{}

	if path, err := os.Executable(); err != nil {
		log.Printf("[ERR] Error finding executable: %v", err)
	} else {
		res = append(res, path)
	}

	res = append(res, dirs...)

	if cd, err := pathing.ConfigDir(); err != nil {
		log.Printf("[ERR] Error loading config directory: %v", err)
	} else {
		res = append(res, filepath.Join(cd, "plugins"))
	}

	if packerPluginPath := os.Getenv("PACKER_PLUGIN_PATH"); packerPluginPath != "" {
		res = append(res, strings.Split(packerPluginPath, string(os.PathListSeparator))...)
	}

	return res
}
