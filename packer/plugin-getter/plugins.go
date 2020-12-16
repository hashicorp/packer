package plugingetter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type BinaryInstallationOptions struct {
	// Usually ".0_x4" for the 4th API version protocol
	// Should be ".0_x4.exe" on windows.
	Extension string
	// OS and ARCH usually should be runtime.GOOS and runtime.ARCH, they allow
	// to pick the correct binary.
	OS, ARCH string

	Checksummers []Checksummer
}

type ListInstallationsOptions struct {
	// FromFolders where plugins could be installed. Paths should be absolute for
	// safety but can also be relative.
	FromFolders []string

	BinaryInstallationOptions
}

func (pr Requirement) filenamePrefix() string {
	return "packer-plugin-" + pr.Identifier.Type + "_"
}

func (opts BinaryInstallationOptions) filenameSuffix() string {
	return "_" + opts.OS + "_" + opts.ARCH + opts.Extension
}

// ListInstallations lists unique installed versions of plugin Requirement pr
// with opts as a filter.
//
// Installations are sorted by version and one binary per version is returned.
// Last binary detected takes precedence: in the order 'FromFolders' option.
//
// You must pass at least one option to Checksumers for a binary to be even
// consider.
func (pr Requirement) ListInstallations(opts ListInstallationsOptions) (InstallList, error) {
	res := InstallList{}
	filenamePrefix := pr.filenamePrefix()
	filenameSuffix := opts.filenameSuffix()
	for _, knownFolder := range opts.FromFolders {
		glob := filepath.Join(knownFolder, pr.Identifier.Hostname, pr.Identifier.Namespace, pr.Identifier.Type, filenamePrefix+"*"+filenameSuffix)

		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("ListInstallations: %q failed to list binaries in folder: %v", pr.Identifier.String(), err)
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
			if !pr.VersionConstraints.Check(pv) {
				log.Printf("[TRACE]: version %q of file %q does not match constraint %q", versionStr, path, pr.VersionConstraints.String())
				continue
			}

			checksumOk := false
			for _, checksummer := range opts.Checksummers {

				cs, err := checksummer.GetChecksumOfFile(path)
				if err != nil {
					log.Printf("[TRACE]: GetChecksumOfFile(%q) failed: %v", path, err)
					continue
				}

				if err := checksummer.ChecksumFile(cs, path); err != nil {
					log.Printf("[TRACE]: ChecksumFile(%q) failed: %v", path, err)
					continue
				}
				checksumOk = true
				break
			}
			if !checksumOk {
				log.Printf("[TRACE]: No checksum found for %q ignoring possibly unsafe binary", path)
				continue
			}

			res.InsertSortedUniq(&Installation{
				BinaryPath: path,
				Version:    versionStr,
			})
		}
	}
	return res, nil
}

// InstallList is a list of installs
type InstallList []*Installation

// InsertSortedUniq inserts the installation in the right spot in the list by
// comparing the version lexicographically.
// A Duplicate version will replace any already present version.
func (l *InstallList) InsertSortedUniq(install *Installation) {
	pos := sort.Search(len(*l), func(i int) bool { return (*l)[i].Version >= install.Version })
	if len(*l) > pos && (*l)[pos].Version == install.Version {
		(*l)[pos] = install
		return
	}
	(*l) = append((*l), nil)
	copy((*l)[pos+1:], (*l)[pos:])
	(*l)[pos] = install
}

// Installation describes a plugin installation
type Installation struct {
	// path to where binary is installed, if installed.
	// Ex: /usr/azr/.packer.d/plugins/github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v1.2.3_darwin_amd64
	BinaryPath string

	// Version of this plugin, if installed and versionned. Ex:
	//  * v1.2.3 for packer-plugin-amazon_v1.2.3_darwin_.0_x5
	//  * empty  for packer-plugin-amazon
	Version string
}

// InstallOptions describes the possible options for installing the plugin that
// fits the plugin Requirement.
type InstallOptions struct {
	// Any downloaded binary and checksum file will be put in this folder.
	//
	InFolders []string

	// If empty then we will try to fetch it.
	Version string

	BinaryInstallationOptions
}

type GetOptions struct {
	PluginRequirement *Requirement

	// If empty then we will try to fetch it.
	Version string

	BinaryInstallationOptions
}

// A Getter helps get the appropriate files to download a binary.
type Getter interface {
	// Get:
	//  * 'releases'
	//  * 'sha256'
	//  * 'binary'
	Get(what string, opts GetOptions) (io.ReadCloser, error)
}

type Release struct {
	Version string `json:"version"`
}

type Releases []Release

func (r Releases) Len() int           { return len(r) }
func (r Releases) Less(i, j int) bool { return r[i].Version < r[j].Version }
func (r Releases) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

var _ sort.Interface = Releases{}

func ParseReleases(f io.ReadCloser) (Releases, error) {
	var releases []Release
	defer f.Close()
	return releases, json.NewDecoder(f).Decode(&releases)
}

func (pr *Requirement) InstallLatest(opts InstallOptions) (InstallList, error) {
	log.Printf("[TRACE] Installing the %s plugin ...", pr.Identifier.ForDisplay())

	var getters []Getter

	getOpts := GetOptions{
		pr,
		opts.Version,
		opts.BinaryInstallationOptions,
	}

	if getOpts.Version == "" {
		for _, getter := range getters {
			releasesFile, err := getter.Get("releases", getOpts)
			if err != nil {
				err := fmt.Errorf("%q getter could not get release: %w", getter, err)
				log.Printf("[TRACE] %s", err.Error())
				continue
			}

			releases, err := ParseReleases(releasesFile)
			if err != nil {
				err := fmt.Errorf("could not parse release: %w", err)
				log.Printf("[TRACE] %s", err.Error())
				continue
			}
			if len(releases) == 0 {
				err := fmt.Errorf("no release found")
				log.Printf("[TRACE] %s", err.Error())
				continue
			}
			sort.Sort(releases)
			getOpts.Version = releases[0].Version
			break
		}
	}

	if getOpts.Version == "" {
		err := fmt.Errorf("no release version found")
		return nil, err
	}

	folder := opts.InFolders[len(opts.InFolders)-1]
	expectedFilename := filepath.Join(folder, pr.filenamePrefix()+getOpts.Version+getOpts.filenameSuffix())

	log.Printf("[TRACE] Installing the %q version for the %s plugin in %q...", getOpts.Version, pr.Identifier.ForDisplay(), folder)

	var checksum *Checksum
	for _, checksummer := range opts.Checksummers {
		// First check if checksum file is already here in the expected
		// download folder. Here we want to download a binary so we only check
		// for an existing checksum file from the folder we want to download
		// into.
		cs, err := checksummer.GetChecksumOfFile(expectedFilename)
		if err == nil && len(cs) > 0 {
			checksum = &Checksum{
				Expected:    cs,
				Checksummer: checksummer,
			}
			log.Printf("[TRACE] found a pre-exising %q checksum file", checksummer.Type)
			break
		}
	}

	for _, getter := range getters {
		for _, checksummer := range opts.Checksummers {

			// First check if checksum file is already here in the expected
			// download folder. Here we want to download a binary so we only check
			// for an existing checksum file from the folder we want to download
			// into.
			cs, err := checksummer.GetChecksumOfFile(expectedFilename)
			if err == nil && len(cs) > 0 {
				checksum = &Checksum{
					Expected:    cs,
					Checksummer: checksummer,
				}
				log.Printf("[TRACE] found a pre-exising %q checksum file", checksummer.Type)
				break
			}
			log.Printf("[TRACE] no %q file found, downloading", expectedFilename+checksum.FileExt())

			checksumFile, err := getter.Get(checksummer.Type, getOpts)
			if err != nil {
				return nil, err
			}
			cs, err = checksummer.ParseChecksum(checksumFile)
			checksumFile.Close()
			if err != nil {
				log.Printf("[TRACE] could not parse %s checksum: %v", checksummer.Type, err)
				continue
			}
			if err := ioutil.WriteFile(expectedFilename+checksum.FileExt(), cs, 0666); err != nil {
				return nil, fmt.Errorf("Could write checksum file %w", err)
			}
			checksum = &Checksum{
				Expected:    cs,
				Checksummer: checksummer,
			}
		}
	}

	// binary, err := getter.Get("binary", getOpts)

	// os.Open(name string)

	return nil, fmt.Errorf("not implemented")
}
