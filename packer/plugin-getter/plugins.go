package plugingetter

import (
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

type Requirements []*Requirement

// Requirement describes a required plugin and how it is installed. Usually a list
// of required plugins is generated from a config file. From it we check what
// is actually installed and what needs to happen to get in the desired state.
type Requirement struct {
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

	Checksummers []Checksummer
}

// ListInstallations lists unique installed versions of Plugin p with opts as a
// filter.
//
// Installations are sorted by version and one binary per version is returned.
// Last binary detected takes precedence: in the order 'FromFolders' option.
//
// You must pass at least one option to Checksumers for a binary to be even
// consider.
func (r Requirement) ListInstallations(opts ListInstallationsOptions) (InstallList, error) {
	res := InstallList{}
	filenamePrefix := "packer-plugin-" + r.Identifier.Type + "_"
	filenameSuffix := "_" + opts.OS + "_" + opts.ARCH + opts.Extension
	for _, knownFolder := range opts.FromFolders {
		glob := filepath.Join(knownFolder, r.Identifier.Hostname, r.Identifier.Namespace, r.Identifier.Type, filenamePrefix+"*"+filenameSuffix)

		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			fname := filepath.Base(path)
			if fname == "." {
				continue
			}

			// base name could look like packer-plugin-amazon_v1.2.3_darwin_amd64.0_x4
			versionStr := strings.TrimPrefix(fname, filenamePrefix)
			versionStr = strings.TrimSuffix(versionStr, filenameSuffix)
			pv, err := version.NewVersion(versionStr)
			if err != nil {
				// could not be parsed, ignoring the file
				log.Printf("[TRACE]: NewVersion(%q): %v", versionStr, err)
				continue
			}

			// no constraint means always pass
			if !r.VersionConstraints.Check(pv) {
				log.Printf("[TRACE]: version %q of file %q does not match constraint %q", versionStr, path, r.VersionConstraints.String())
				continue
			}

			checksumOk := false
			for _, checksummer := range opts.Checksummers {
				if err := checksummer.Checksum(path); err != nil {
					log.Printf("[TRACE]: Checksum(%q) failed: %v", path, err)
					continue
				}
				checksumOk = true
				break
			}
			if !checksumOk {
				log.Printf("[TRACE]: No checksum found for %q ignoring possibly unsafe binary", path)
				continue
			}

			res.InsertSortedUniq(&Install{
				Path:    path,
				Version: versionStr,
			})
		}
	}
	return res, nil
}

// InstallList is a list of installs
type InstallList []*Install

// InsertSortedUniq inserts the installation in the right spot in the list by
// comparing the version lexicographically.
// A Duplicate version will replace any already present version.
func (l *InstallList) InsertSortedUniq(install *Install) {
	pos := sort.Search(len(*l), func(i int) bool { return (*l)[i].Version >= install.Version })
	if len(*l) > pos && (*l)[pos].Version == install.Version {
		(*l)[pos] = install
		return
	}
	(*l) = append((*l), nil)
	copy((*l)[pos+1:], (*l)[pos:])
	(*l)[pos] = install
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
