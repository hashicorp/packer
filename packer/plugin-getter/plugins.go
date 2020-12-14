package plugingetter

import (
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

// List of plugins
type List []*Plugin

// Plugins entity
type Plugin struct {
	// Something like github.com/hashicorp/packer-plugin-amazon
	Identifier *addrs.Plugin

	// VersionConstraints as defined by user.
	VersionConstraints version.Constraints

	// Could be defined by user or taken from expected checksum file.
	ExpectedChecksum string

	Install struct {
		// Path to where it is installed, if installed.
		// Ex: /usr/azr/.packer.d/plugins/packer-plugin-amazon_v1.2.3_darwin_amd64
		Path string

		// Version of this plugin, if installed and versionned. Ex:
		//  * v1.2.3 for /usr/azr/.packer.d/plugins/packer-plugin-amazon_v1.2.3_darwin_amd64
		//  * empty  for /usr/azr/.packer.d/plugins/packer-plugin-amazon
		Version string

		// Checksum of the binary, if installed.
		Checksum string
	}
}

var _ sort.Interface = List{}

func (l List) Len() int           { return len(l) }
func (l List) Less(i, j int) bool { return l[i].Identifier.String() < l[j].Identifier.String() }
func (l List) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
