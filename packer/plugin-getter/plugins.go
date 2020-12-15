package plugingetter

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

// List of plugins
type List []*Plugin

// Plugin describes a required plugin and how it is installed. Usually a list
// of required plugins is generated from a config file. From it we check what
// is actually installed and what needs to happen to get in the desired state.
type Plugin struct {
	// Something like github.com/hashicorp/packer-plugin-amazon
	Identifier *addrs.Plugin

	// VersionConstraints as defined by user. Empty ( to be avoided ) means
	// highest found version.
	VersionConstraints version.Constraints
}

type ListInstallationsOptions struct {
	// Put the folders where plugins could be installed in this list. Paths
	// should be absolute for safety but can also be relative.
	FromFolders []string
	// Usually ".x04" for the 4th API version protocol
	// Should be ".x04.exe" on windows.
	Extension string

	// OS and ARCH usually should be runtime.GOOS and runtime.ARCH, they allow
	// to pick the correct binary.
	OS, ARCH string
}

// ListInstallations lists installed versions of Plugin p from knownFolders.
func (p Plugin) ListInstallations(opts ListInstallationsOptions) ([]Install, error) {
	res := []Install{}
	filenamePrefix := "packer-plugin-" + p.Identifier.Type + "_"
	filenameSuffix := "_" + opts.OS + "_" + opts.ARCH + opts.Extension
	for _, knownFolder := range opts.FromFolders {
		glob := filepath.Join(knownFolder, p.Identifier.Hostname, p.Identifier.Namespace, p.Identifier.Type, filenamePrefix+"*"+filenameSuffix)

		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			fname := filepath.Base(path)
			if fname == "." {
				continue
			}

			// last part could look like packer-plugin-amazon_v1.2.3_darwin_amd64.0_x4
			versionStr := strings.TrimPrefix(fname, filenamePrefix)
			versionStr = strings.TrimSuffix(versionStr, filenameSuffix)
			pv, err := version.NewVersion(versionStr)
			if err != nil {
				// could not be parsed, ignoring the file
				log.Printf("[TRACE]: NewVersion(%q): %v", versionStr, err)
				continue
			}

			// no constraint means always pass
			if !p.VersionConstraints.Check(pv) {
				log.Printf("[TRACE]: version %q of file %q does not match constraint %q", versionStr, path, p.VersionConstraints.String())
				continue
			}

			res = append(res, Install{
				Path:    path,
				Version: versionStr,
			})
		}
	}
	return res, nil
}

// Install describes a plugin installation
type Install struct {
	// Path to where it is installed, if installed.
	// Ex: /usr/azr/.packer.d/plugins/github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v1.2.3_darwin_amd64
	Path string

	// Version of this plugin, if installed and versionned. Ex:
	//  * v1.2.3 for packer-plugin-amazon_v1.2.3_darwin_amd64
	//  * empty  for packer-plugin-amazon
	Version string
}
