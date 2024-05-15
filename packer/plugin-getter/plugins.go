// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugingetter

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	goversion "github.com/hashicorp/go-version"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"golang.org/x/mod/semver"
)

type Requirements []*Requirement

// Requirement describes a required plugin and how it is installed. Usually a list
// of required plugins is generated from a config file. From it we check what
// is actually installed and what needs to happen to get in the desired state.
type Requirement struct {
	// Plugin accessor as defined in the config file.
	// For Packer, using :
	//  required_plugins { amazon = {...} }
	// Will set Accessor to `amazon`.
	Accessor string

	// Something like github.com/hashicorp/packer-plugin-amazon, from the
	// previous example.
	Identifier *addrs.Plugin

	// VersionConstraints as defined by user. Empty ( to be avoided ) means
	// highest found version.
	VersionConstraints goversion.Constraints
}

type BinaryInstallationOptions struct {
	// The API version with which to check remote compatibility
	//
	// They're generally extracted from the SDK since it's what Packer Core
	// supports as far as the protocol goes
	APIVersionMajor, APIVersionMinor string
	// OS and ARCH usually should be runtime.GOOS and runtime.ARCH, they allow
	// to pick the correct binary.
	OS, ARCH string

	// Ext is ".exe" on windows
	Ext string

	Checksummers []Checksummer

	// ReleasesOnly may be set by commands like validate or build, and
	// forces Packer to not consider plugin pre-releases.
	ReleasesOnly bool
}

type ListInstallationsOptions struct {
	// The directory in which to look for when installing plugins
	PluginDirectory string

	BinaryInstallationOptions
}

// RateLimitError is returned when a getter is being rate limited.
type RateLimitError struct {
	SetableEnvVar string
	ResetTime     time.Time
	Err           error
}

func (rlerr *RateLimitError) Error() string {
	s := fmt.Sprintf("Plugin host rate limited the plugin getter. Try again in %s.\n", time.Until(rlerr.ResetTime))
	if rlerr.SetableEnvVar != "" {
		s += fmt.Sprintf("HINT: Set the %s env var with a token to get more requests.\n", rlerr.SetableEnvVar)
	}
	s += rlerr.Err.Error()
	return s
}

// PrereleaseInstallError is returned when a getter encounters the install of a pre-release version.
type PrereleaseInstallError struct {
	PluginSrc string
	Err       error
}

func (e *PrereleaseInstallError) Error() string {
	var s strings.Builder
	s.WriteString(e.Err.Error() + "\n")
	s.WriteString("Remote installation of pre-release plugin versions is unsupported.\n")
	s.WriteString("This is likely an upstream issue, which should be reported.\n")
	s.WriteString("If you require this specific version of the plugin, download the binary and install it manually.\n")
	s.WriteString("\npacker plugins install --path '<plugin_binary>' " + e.PluginSrc)
	return s.String()
}

// ContinuableInstallError describe a failed getter install that is
// capable of falling back to next available version.
type ContinuableInstallError struct {
	Err error
}

func (e *ContinuableInstallError) Error() string {
	return fmt.Sprintf("Continuing to next available version: %s", e.Err)
}

func (pr Requirement) FilenamePrefix() string {
	if pr.Identifier == nil {
		return "packer-plugin-"
	}

	return "packer-plugin-" + pr.Identifier.Name() + "_"
}

func (opts BinaryInstallationOptions) FilenameSuffix() string {
	return "_" + opts.OS + "_" + opts.ARCH + opts.Ext
}

// getPluginBinaries lists the plugin binaries installed locally.
//
// Each plugin binary must be in the right hierarchy (not root) and has to be
// conforming to the packer-plugin-<name>_<version>_<API>_<os>_<arch> convention.
func (pr Requirement) getPluginBinaries(opts ListInstallationsOptions) ([]string, error) {
	var matches []string

	rootdir := opts.PluginDirectory
	if pr.Identifier != nil {
		rootdir = filepath.Join(rootdir, path.Dir(pr.Identifier.Source))
	}

	if _, err := os.Lstat(rootdir); err != nil {
		log.Printf("Directory %q does not exist, the plugin likely isn't installed locally yet.", rootdir)
		return matches, nil
	}

	err := filepath.WalkDir(rootdir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// No need to inspect directory entries, we can continue walking
		if d.IsDir() {
			return nil
		}

		// Skip plugins installed at root, only those in a hierarchy should be considered valid
		if filepath.Dir(path) == opts.PluginDirectory {
			return nil
		}

		// If the binary's name doesn't start with packer-plugin-, we skip it.
		if !strings.HasPrefix(filepath.Base(path), pr.FilenamePrefix()) {
			return nil
		}
		// If the binary's name doesn't match the expected convention, we skip it
		if !strings.HasSuffix(filepath.Base(path), opts.FilenameSuffix()) {
			return nil
		}

		matches = append(matches, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	retMatches := make([]string, 0, len(matches))
	// Don't keep plugins that are nested too deep in the hierarchy
	for _, match := range matches {
		dir := strings.Replace(filepath.Dir(match), opts.PluginDirectory, "", 1)
		parts := strings.FieldsFunc(dir, func(r rune) bool {
			return r == '/'
		})
		if len(parts) > 16 {
			log.Printf("[WARN] plugin %q ignored, too many levels of depth: %d (max 16)", match, len(parts))
			continue
		}

		retMatches = append(retMatches, match)
	}

	return retMatches, err
}

// ListInstallations lists unique installed versions of plugin Requirement pr
// with opts as a filter.
//
// Installations are sorted by version and one binary per version is returned.
// Last binary detected takes precedence: in the order 'FromFolders' option.
//
// At least one opts.Checksumers must be given for a binary to be even
// considered.
func (pr Requirement) ListInstallations(opts ListInstallationsOptions) (InstallList, error) {
	res := InstallList{}
	log.Printf("[TRACE] listing potential installations for %q that match %q. %#v", pr.Identifier, pr.VersionConstraints, opts)

	matches, err := pr.getPluginBinaries(opts)
	if err != nil {
		return nil, fmt.Errorf("ListInstallations: failed to list installed plugins: %s", err)
	}

	for _, path := range matches {
		fname := filepath.Base(path)
		if fname == "." {
			continue
		}

		checksumOk := false
		for _, checksummer := range opts.Checksummers {

			cs, err := checksummer.GetCacheChecksumOfFile(path)
			if err != nil {
				log.Printf("[TRACE] GetChecksumOfFile(%q) failed: %v", path, err)
				continue
			}

			if err := checksummer.ChecksumFile(cs, path); err != nil {
				log.Printf("[TRACE] ChecksumFile(%q) failed: %v", path, err)
				continue
			}
			checksumOk = true
			break
		}
		if !checksumOk {
			log.Printf("[TRACE] No checksum found for %q ignoring possibly unsafe binary", path)
			continue
		}

		// base name could look like packer-plugin-amazon_v1.2.3_x5.1_darwin_amd64.exe
		versionsStr := strings.TrimPrefix(fname, pr.FilenamePrefix())
		versionsStr = strings.TrimSuffix(versionsStr, opts.FilenameSuffix())

		if pr.Identifier == nil {
			if idx := strings.Index(versionsStr, "_"); idx > 0 {
				versionsStr = versionsStr[idx+1:]
			}
		}

		describeInfo, err := GetPluginDescription(path)
		if err != nil {
			log.Printf("failed to call describe on %q: %s", path, err)
			continue
		}

		// versionsStr now looks like v1.2.3_x5.1 or amazon_v1.2.3_x5.1
		parts := strings.SplitN(versionsStr, "_", 2)
		pluginVersionStr, protocolVersionStr := parts[0], parts[1]
		ver, err := goversion.NewVersion(pluginVersionStr)
		if err != nil {
			// could not be parsed, ignoring the file
			log.Printf("found %q with an incorrect %q version, ignoring it. %v", path, pluginVersionStr, err)
			continue
		}

		if fmt.Sprintf("v%s", ver.String()) != pluginVersionStr {
			log.Printf("version %q in path is non canonical, this could introduce ambiguity and is not supported, ignoring it.", pluginVersionStr)
			continue
		}

		if ver.Prerelease() != "" && opts.ReleasesOnly {
			log.Printf("ignoring pre-release plugin %q", path)
			continue
		}

		if ver.Metadata() != "" {
			log.Printf("found version %q with metadata in the name, this could introduce ambiguity and is not supported, ignoring it.", pluginVersionStr)
			continue
		}

		descVersion, err := goversion.NewVersion(describeInfo.Version)
		if err != nil {
			log.Printf("malformed reported version string %q: %s, ignoring", describeInfo.Version, err)
			continue
		}

		if ver.Compare(descVersion) != 0 {
			log.Printf("plugin %q reported version %q while its name implies version %q, ignoring", path, describeInfo.Version, pluginVersionStr)
			continue
		}

		preRel := descVersion.Prerelease()
		if preRel != "" && preRel != "dev" {
			log.Printf("invalid plugin pre-release version %q, only development or release binaries are accepted", pluginVersionStr)
		}

		// Check the API version matches between path and describe
		if describeInfo.APIVersion != protocolVersionStr {
			log.Printf("plugin %q reported API version %q while its name implies version %q, ignoring", path, describeInfo.APIVersion, protocolVersionStr)
			continue
		}

		// no constraint means always pass, this will happen for implicit
		// plugin requirements and when we list all plugins.
		//
		// Note: we use the raw version name here, without the pre-release
		// suffix, as otherwise constraints reject them, which is not
		// what we want by default.
		if !pr.VersionConstraints.Check(ver.Core()) {
			log.Printf("[TRACE] version %q of file %q does not match constraint %q", pluginVersionStr, path, pr.VersionConstraints.String())
			continue
		}

		if err := opts.CheckProtocolVersion(protocolVersionStr); err != nil {
			log.Printf("[NOTICE] binary %s requires protocol version %s that is incompatible "+
				"with this version of Packer. %s", path, protocolVersionStr, err)
			continue
		}

		res = append(res, &Installation{
			BinaryPath: path,
			Version:    pluginVersionStr,
			APIVersion: describeInfo.APIVersion,
		})
	}

	sort.Sort(res)

	return res, nil
}

// InstallList is a list of installed plugins (binaries) with their versions,
// ListInstallations should be used to get an InstallList.
//
// ListInstallations sorts binaries by version and one binary per version is
// returned.
type InstallList []*Installation

func (l InstallList) String() string {
	v := &strings.Builder{}
	v.Write([]byte("["))
	for i, inst := range l {
		if i > 0 {
			v.Write([]byte(","))
		}
		fmt.Fprintf(v, "%v", *inst)
	}
	v.Write([]byte("]"))
	return v.String()
}

// Len is the number of elements in the collection.
func (l InstallList) Len() int {
	return len(l)
}

var rawPluginName = regexp.MustCompile("packer-plugin-[^_]+")

// Less reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
//
// Less must describe a transitive ordering:
//   - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// See Float64Slice.Less for a correct implementation for floating-point values.
func (l InstallList) Less(i, j int) bool {
	lowPluginPath := l[i]
	hiPluginPath := l[j]

	lowRawPluginName := rawPluginName.FindString(path.Base(lowPluginPath.BinaryPath))
	hiRawPluginName := rawPluginName.FindString(path.Base(hiPluginPath.BinaryPath))

	// We group by path, then by descending order for the versions
	//
	// i.e. if the path are not the same, we can return the plain
	// lexicographic order, otherwise, we'll do a semver-conscious
	// version comparison for sorting.
	if lowRawPluginName != hiRawPluginName {
		return lowRawPluginName < hiRawPluginName
	}

	verCmp := semver.Compare(lowPluginPath.Version, hiPluginPath.Version)
	if verCmp != 0 {
		return verCmp < 0
	}

	// Ignore errors here, they are already validated when populating the InstallList
	loAPIVer, _ := NewAPIVersion(lowPluginPath.APIVersion)
	hiAPIVer, _ := NewAPIVersion(hiPluginPath.APIVersion)

	if loAPIVer.Major != hiAPIVer.Major {
		return loAPIVer.Major < hiAPIVer.Major
	}

	return loAPIVer.Minor < hiAPIVer.Minor
}

// Swap swaps the elements with indexes i and j.
func (l InstallList) Swap(i, j int) {
	tmp := l[i]
	l[i] = l[j]
	l[j] = tmp
}

// Installation describes a plugin installation
type Installation struct {
	// Path to where binary is installed.
	// Ex: /usr/azr/.packer.d/plugins/github.com/hashicorp/amazon/packer-plugin-amazon_v1.2.3_darwin_amd64
	BinaryPath string

	// Version of this plugin. Ex:
	//  * v1.2.3 for packer-plugin-amazon_v1.2.3_darwin_x5
	Version string

	// API version for the plugin. Ex:
	//  * 5.0 for packer-plugin-amazon_v1.2.3_darwin_x5.0
	//  * 5.1 for packer-plugin-amazon_v1.2.3_darwin_x5.1
	APIVersion string
}

// InstallOptions describes the possible options for installing the plugin that
// fits the plugin Requirement.
type InstallOptions struct {
	// Different means to get releases, sha256 and binary files.
	Getters []Getter

	// The directory in which the plugins should be installed
	PluginDirectory string

	// Forces installation of the plugin, even if already installed.
	Force bool

	BinaryInstallationOptions
}

type GetOptions struct {
	PluginRequirement *Requirement

	BinaryInstallationOptions

	version *goversion.Version

	expectedZipFilename string
}

// ExpectedZipFilename is the filename of the zip we expect to find, the
// value is known only after parsing the checksum file file.
func (gp *GetOptions) ExpectedZipFilename() string {
	return gp.expectedZipFilename
}

type APIVersion struct {
	Major int
	Minor int
}

func NewAPIVersion(apiVersion string) (APIVersion, error) {
	ver := APIVersion{}

	apiVersion = strings.TrimPrefix(strings.TrimSpace(apiVersion), "x")
	parts := strings.Split(apiVersion, ".")
	if len(parts) < 2 {
		return ver, fmt.Errorf(
			"Invalid remote protocol: %q, expected something like '%s.%s'",
			apiVersion, pluginsdk.APIVersionMajor, pluginsdk.APIVersionMinor,
		)
	}

	vMajor, err := strconv.Atoi(parts[0])
	if err != nil {
		return ver, err
	}
	ver.Major = vMajor

	vMinor, err := strconv.Atoi(parts[1])
	if err != nil {
		return ver, err
	}
	ver.Minor = vMinor

	return ver, nil
}

var localAPIVersion APIVersion

func (binOpts *BinaryInstallationOptions) CheckProtocolVersion(remoteProt string) error {
	// no protocol version check
	if binOpts.APIVersionMajor == "" && binOpts.APIVersionMinor == "" {
		return nil
	}

	localVersion := localAPIVersion
	if binOpts.APIVersionMajor != pluginsdk.APIVersionMajor ||
		binOpts.APIVersionMinor != pluginsdk.APIVersionMinor {
		var err error

		localVersion, err = NewAPIVersion(fmt.Sprintf("x%s.%s", binOpts.APIVersionMajor, binOpts.APIVersionMinor))
		if err != nil {
			return fmt.Errorf("Failed to parse API Version from constraints: %s", err)
		}
	}

	remoteVersion, err := NewAPIVersion(remoteProt)
	if err != nil {
		return err
	}

	if localVersion.Major != remoteVersion.Major {
		return fmt.Errorf("Unsupported remote protocol MAJOR version %d. The current MAJOR protocol version is %d."+
			" This version of Packer can only communicate with plugins using that version.", remoteVersion.Major, localVersion.Major)
	}

	if remoteVersion.Minor > localVersion.Minor {
		return fmt.Errorf("Unsupported remote protocol MINOR version %d. The supported MINOR protocol versions are version %d and below. "+
			"Please upgrade Packer or use an older version of the plugin if possible.", remoteVersion.Minor, localVersion.Minor)
	}

	return nil
}

func (gp *GetOptions) Version() string {
	return "v" + gp.version.String()
}

// A Getter helps get the appropriate files to download a binary.
type Getter interface {
	// Get allows Packer to know more information about releases of a plugin in
	// order to decide which version to install. Get behaves similarly to an
	// HTTP server. Packer will stream responses from get in order to do what's
	// needed. In order to minimize the amount of requests done, Packer is
	// strict on filenames and we highly recommend on automating releases.
	// In the future, Packer will make it possible to ship plugin getters as
	// binaries this is why Packer streams from the output of get, which will
	// then be a command.
	//
	//  * 'releases', get 'releases' should return the complete list of Releases
	//    in JSON format following the format of the Release struct. It is also
	//    possible to read GetOptions to filter for a smaller response. Some
	//    getters don't. Packer will then decide the highest compatible
	//    version of the plugin to install by using the sha256 function.
	//
	//  * 'sha256', get 'sha256' should return a SHA256SUMS txt file. It will be
	//    called with the highest possible & user allowed version from get
	//   'releases'. Packer will check if the release has a binary matching what
	//    Packer can install and use. If so, get 'binary' will be called;
	//    otherwise, lower versions will be checked.
	//    For version 1.0.0 of the 'hashicorp/amazon' builder, the GitHub getter
	//    will fetch the following URL:
	//    https://github.com/hashicorp/packer-plugin-amazon/releases/download/v1.0.0/packer-plugin-amazon_v1.0.0_SHA256SUMS
	//    This URL can be parameterized to the following one:
	//    https://github.com/{plugin.path}/releases/download/{plugin.version}/packer-plugin-{plugin.name}_{plugin.version}_SHA256SUMS
	//    If Packer is running on Linux AMD 64, then Packer will check for the
	//    existence of a packer-plugin-amazon_v1.0.0_x5.0_linux_amd64 checksum in
	//    that file. This filename can be parameterized to the following one:
	//    packer-plugin-{plugin.name}_{plugin.version}_x{proto_ver.major}.{proto_ver._minor}_{os}_{arch}
	//
	//    See
	//    https://github.com/hashicorp/packer-plugin-scaffolding/blob/main/.goreleaser.yml
	//    and
	//    https://www.packer.io/docs/plugins/creation#plugin-development-basics
	//    to learn how to create and automate your releases and for docs on
	//    plugin development basics.
	//
	//  * get 'zip' is called once we know what version we want and that it is
	//    compatible with the OS and Packer. Zip expects an io stream of a zip
	//    file containing a binary. For version 1.0.0 of the 'hashicorp/amazon'
	//    builder and on darwin_amd64, the GitHub getter will fetch the
	//    following ZIP:
	//    https://github.com/hashicorp/packer-plugin-amazon/releases/download/v1.0.0/packer-plugin-amazon_v1.0.0_x5.0_darwin_amd64.zip
	//    this zip is expected to contain a
	//    packer-plugin-amazon_v1.0.0_x5.0_linux_amd64 file that will be checksum
	//    verified then copied to the correct plugin location.
	Get(what string, opts GetOptions) (io.ReadCloser, error)
}

type Release struct {
	Version string `json:"version"`
}

func ParseReleases(f io.ReadCloser) ([]Release, error) {
	var releases []Release
	defer f.Close()
	return releases, json.NewDecoder(f).Decode(&releases)
}

type ChecksumFileEntry struct {
	Filename                  string `json:"filename"`
	Checksum                  string `json:"checksum"`
	ext, binVersion, os, arch string
	protVersion               string
}

func (e ChecksumFileEntry) Ext() string         { return e.ext }
func (e ChecksumFileEntry) BinVersion() string  { return e.binVersion }
func (e ChecksumFileEntry) ProtVersion() string { return e.protVersion }
func (e ChecksumFileEntry) Os() string          { return e.os }
func (e ChecksumFileEntry) Arch() string        { return e.arch }

// a file inside will look like so:
//
//	packer-plugin-comment_v0.2.12_x5.0_freebsd_amd64.zip
func (e *ChecksumFileEntry) init(req *Requirement) (err error) {
	filename := e.Filename
	res := strings.TrimPrefix(filename, req.FilenamePrefix())
	// res now looks like v0.2.12_x5.0_freebsd_amd64.zip

	e.ext = filepath.Ext(res)

	res = strings.TrimSuffix(res, e.ext)
	// res now looks like v0.2.12_x5.0_freebsd_amd64

	parts := strings.Split(res, "_")
	// ["v0.2.12", "x5.0", "freebsd", "amd64"]
	if len(parts) < 4 {
		return fmt.Errorf("malformed filename expected %s{version}_x{protocol-version}_{os}_{arch}", req.FilenamePrefix())
	}

	e.binVersion, e.protVersion, e.os, e.arch = parts[0], parts[1], parts[2], parts[3]

	return err
}

func (e *ChecksumFileEntry) validate(expectedVersion string, installOpts BinaryInstallationOptions) error {
	if e.binVersion != expectedVersion {
		return fmt.Errorf("wrong version: '%s' does not match expected %s ", e.binVersion, expectedVersion)
	}
	if e.os != installOpts.OS || e.arch != installOpts.ARCH {
		return fmt.Errorf("wrong system, expected %s_%s ", installOpts.OS, installOpts.ARCH)
	}

	return installOpts.CheckProtocolVersion(e.protVersion)
}

func ParseChecksumFileEntries(f io.Reader) ([]ChecksumFileEntry, error) {
	var entries []ChecksumFileEntry
	return entries, json.NewDecoder(f).Decode(&entries)
}

func (pr *Requirement) InstallLatest(opts InstallOptions) (*Installation, error) {

	getters := opts.Getters

	log.Printf("[TRACE] getting available versions for the %s plugin", pr.Identifier)
	versions := goversion.Collection{}
	var errs *multierror.Error
	for _, getter := range getters {

		releasesFile, err := getter.Get("releases", GetOptions{
			PluginRequirement:         pr,
			BinaryInstallationOptions: opts.BinaryInstallationOptions,
		})
		if err != nil {
			errs = multierror.Append(errs, err)
			log.Printf("[TRACE] %s", err.Error())
			continue
		}

		releases, err := ParseReleases(releasesFile)
		if err != nil {
			err := fmt.Errorf("could not parse release: %w", err)
			errs = multierror.Append(errs, err)
			log.Printf("[TRACE] %s", err.Error())
			continue
		}
		if len(releases) == 0 {
			err := fmt.Errorf("no release found")
			errs = multierror.Append(errs, err)
			log.Printf("[TRACE] %s", err.Error())
			continue
		}
		for _, release := range releases {
			v, err := goversion.NewVersion(release.Version)
			if err != nil {
				err := fmt.Errorf("could not parse release version %s. %w", release.Version, err)
				errs = multierror.Append(errs, err)
				log.Printf("[TRACE] %s, ignoring it", err.Error())
				continue
			}
			if pr.VersionConstraints.Check(v) {
				versions = append(versions, v)
			}
		}
		if len(versions) == 0 {
			err := fmt.Errorf("no matching version found in releases. In %v", releases)
			errs = multierror.Append(errs, err)
			log.Printf("[TRACE] %s", err.Error())
			continue
		}

		break
	}

	if len(versions) == 0 {
		if errs.Len() == 0 {
			err := fmt.Errorf("no release version found for constraints: %q", pr.VersionConstraints.String())
			errs = multierror.Append(errs, err)
		}
		return nil, errs
	}

	// Here we want to try every release in order, starting from the highest one
	// that matches the requirements. The system and protocol version need to
	// match too.
	sort.Sort(sort.Reverse(versions))
	log.Printf("[DEBUG] will try to install: %s", versions)

	for _, version := range versions {
		//TODO(azr): split in its own InstallVersion(version, opts) function

		outputFolder := filepath.Join(
			// Pick last folder as it's the one with the highest priority
			opts.PluginDirectory,
			// add expected full path
			filepath.Join(pr.Identifier.Parts()...),
		)

		log.Printf("[TRACE] fetching checksums file for the %q version of the %s plugin in %q...", version, pr.Identifier, outputFolder)

		var checksum *FileChecksum
		for _, getter := range getters {
			if checksum != nil {
				break
			}
			for _, checksummer := range opts.Checksummers {
				if checksum != nil {
					break
				}
				checksumFile, err := getter.Get(checksummer.Type, GetOptions{
					PluginRequirement:         pr,
					BinaryInstallationOptions: opts.BinaryInstallationOptions,
					version:                   version,
				})
				if err != nil {
					err := fmt.Errorf("could not get %s checksum file for %s version %s. Is the file present on the release and correctly named ? %w", checksummer.Type, pr.Identifier, version, err)
					errs = multierror.Append(errs, err)
					log.Printf("[TRACE] %s", err)
					continue
				}
				entries, err := ParseChecksumFileEntries(checksumFile)
				_ = checksumFile.Close()
				if err != nil {
					err := fmt.Errorf("could not parse %s checksumfile: %v. Make sure the checksum file contains a checksum and a binary filename per line", checksummer.Type, err)
					errs = multierror.Append(errs, err)
					log.Printf("[TRACE] %s", err)
					continue
				}

				for _, entry := range entries {
					if err := entry.init(pr); err != nil {
						err := fmt.Errorf("could not parse checksum filename %s. Is it correctly formatted ? %s", entry.Filename, err)
						errs = multierror.Append(errs, err)
						log.Printf("[TRACE] %s", err)
						continue
					}
					if err := entry.validate("v"+version.String(), opts.BinaryInstallationOptions); err != nil {
						continue
					}

					log.Printf("[TRACE] About to get: %s", entry.Filename)

					cs, err := checksummer.ParseChecksum(strings.NewReader(entry.Checksum))
					if err != nil {
						err := fmt.Errorf("could not parse %s checksum: %s. Make sure the checksum file contains the checksum and only the checksum", checksummer.Type, err)
						errs = multierror.Append(errs, err)
						log.Printf("[TRACE] %s", err)
						continue
					}

					checksum = &FileChecksum{
						Filename:    entry.Filename,
						Expected:    cs,
						Checksummer: checksummer,
					}
					expectedZipFilename := checksum.Filename
					expectedBinaryFilename := strings.TrimSuffix(expectedZipFilename, filepath.Ext(expectedZipFilename)) + opts.BinaryInstallationOptions.Ext
					outputFileName := filepath.Join(outputFolder, expectedBinaryFilename)

					for _, potentialChecksumer := range opts.Checksummers {
						// First check if a local checksum file is already here in the expected
						// download folder. Here we want to download a binary so we only check
						// for an existing checksum file from the folder we want to download
						// into.
						cs, err := potentialChecksumer.GetCacheChecksumOfFile(outputFileName)
						if err == nil && len(cs) > 0 {
							localChecksum := &FileChecksum{
								Expected:    cs,
								Checksummer: potentialChecksumer,
							}

							log.Printf("[TRACE] found a pre-existing %q checksum file", potentialChecksumer.Type)
							// if outputFile is there and matches the checksum: do nothing more.
							if err := localChecksum.ChecksumFile(localChecksum.Expected, outputFileName); err == nil && !opts.Force {
								log.Printf("[INFO] %s v%s plugin is already correctly installed in %q", pr.Identifier, version, outputFileName)
								return nil, nil // success
							}
						}
					}

					for _, getter := range getters {
						// start fetching binary
						remoteZipFile, err := getter.Get("zip", GetOptions{
							PluginRequirement:         pr,
							BinaryInstallationOptions: opts.BinaryInstallationOptions,
							version:                   version,
							expectedZipFilename:       expectedZipFilename,
						})
						if err != nil {
							errs = multierror.Append(errs,
								fmt.Errorf("could not get binary for %s version %s. Is the file present on the release and correctly named ? %s",
									pr.Identifier, version, err))
							continue
						}
						// create temporary file that will receive a temporary binary.zip
						tmpFile, err := tmp.File("packer-plugin-*.zip")
						if err != nil {
							err = fmt.Errorf("could not create temporary file to download plugin: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						defer func() {
							tmpFilePath := tmpFile.Name()
							tmpFile.Close()
							os.Remove(tmpFilePath)
						}()
						// write binary to tmp file
						_, err = io.Copy(tmpFile, remoteZipFile)
						_ = remoteZipFile.Close()
						if err != nil {
							err := fmt.Errorf("Error getting plugin, trying another getter: %w", err)
							errs = multierror.Append(errs, err)
							continue
						}
						if _, err := tmpFile.Seek(0, 0); err != nil {
							err := fmt.Errorf("Error seeking beginning of temporary file for checksumming, continuing: %w", err)
							errs = multierror.Append(errs, err)
							continue
						}
						// verify that the checksum for the zip is what we expect.
						if err := checksum.Checksummer.Checksum(checksum.Expected, tmpFile); err != nil {
							err := fmt.Errorf("%w. Is the checksum file correct ? Is the binary file correct ?", err)
							errs = multierror.Append(errs, err)
							continue
						}
						zr, err := zip.OpenReader(tmpFile.Name())
						if err != nil {
							errs = multierror.Append(errs, fmt.Errorf("zip : %v", err))
							return nil, errs
						}

						var copyFrom io.ReadCloser
						for _, f := range zr.File {
							if f.Name != expectedBinaryFilename {
								continue
							}
							copyFrom, err = f.Open()
							if err != nil {
								multierror.Append(errs, fmt.Errorf("failed to open temp file: %w", err))
								return nil, errs
							}
							break
						}
						if copyFrom == nil {
							err := fmt.Errorf("could not find a %q file in zipfile", expectedBinaryFilename)
							errs = multierror.Append(errs, err)
							return nil, errs
						}

						var outputFileData bytes.Buffer
						if _, err := io.Copy(&outputFileData, copyFrom); err != nil {
							err := fmt.Errorf("extract file: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						tmpBinFileName := filepath.Join(os.TempDir(), expectedBinaryFilename)
						tmpOutputFile, err := os.OpenFile(tmpBinFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
						if err != nil {
							err = fmt.Errorf("could not create temporary file to download plugin: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						defer func() {
							os.Remove(tmpBinFileName)
						}()

						if _, err := tmpOutputFile.Write(outputFileData.Bytes()); err != nil {
							err := fmt.Errorf("extract file: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						tmpOutputFile.Close()

						if err := checkVersion(tmpBinFileName, pr.Identifier.String(), version); err != nil {
							errs = multierror.Append(errs, err)
							var continuableError *ContinuableInstallError
							if errors.As(err, &continuableError) {
								continue
							}
							return nil, errs
						}

						// create directories if need be
						if err := os.MkdirAll(outputFolder, 0755); err != nil {
							err := fmt.Errorf("could not create plugin folder %q: %w", outputFolder, err)
							errs = multierror.Append(errs, err)
							log.Printf("[TRACE] %s", err.Error())
							return nil, errs
						}
						outputFile, err := os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
						if err != nil {
							err = fmt.Errorf("could not create final plugin binary file: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						if _, err := outputFile.Write(outputFileData.Bytes()); err != nil {
							err = fmt.Errorf("could not write final plugin binary file: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						outputFile.Close()

						cs, err := checksum.Checksummer.Sum(&outputFileData)
						if err != nil {
							err := fmt.Errorf("failed to checksum binary file: %s", err)
							errs = multierror.Append(errs, err)
							log.Printf("[WARNING] %v, ignoring", err)
						}
						if err := os.WriteFile(outputFileName+checksum.Checksummer.FileExt(), []byte(hex.EncodeToString(cs)), 0644); err != nil {
							err := fmt.Errorf("failed to write local binary checksum file: %s", err)
							errs = multierror.Append(errs, err)
							log.Printf("[WARNING] %v, ignoring", err)
							os.Remove(outputFileName)
							continue
						}

						// Success !!
						return &Installation{
							BinaryPath: strings.ReplaceAll(outputFileName, "\\", "/"),
							Version:    "v" + version.String(),
						}, nil
					}
				}
			}
		}
	}

	if errs.ErrorOrNil() == nil {
		err := fmt.Errorf("could not find a local nor a remote checksum for plugin %q %q", pr.Identifier, pr.VersionConstraints)
		errs = multierror.Append(errs, err)
	}
	errs = multierror.Append(errs, fmt.Errorf("could not install any compatible version of plugin %q", pr.Identifier))
	return nil, errs
}

func GetPluginDescription(pluginPath string) (pluginsdk.SetDescription, error) {
	out, err := exec.Command(pluginPath, "describe").Output()
	if err != nil {
		return pluginsdk.SetDescription{}, err
	}

	desc := pluginsdk.SetDescription{}
	err = json.Unmarshal(out, &desc)

	return desc, err
}

// checkVersion checks the described version of a plugin binary against the requested version constriant.
// A ContinuableInstallError is returned upon a version mismatch to indicate that the caller should try the next
// available version. A PrereleaseInstallError is returned to indicate an unsupported version install.
func checkVersion(binPath string, identifier string, version *goversion.Version) error {
	desc, err := GetPluginDescription(binPath)
	if err != nil {
		err := fmt.Errorf("failed to describe plugin binary %q: %s", binPath, err)
		return &ContinuableInstallError{Err: err}
	}
	descVersion, err := goversion.NewSemver(desc.Version)
	if err != nil {
		err := fmt.Errorf("invalid self-reported version %q: %s", desc.Version, err)
		return &ContinuableInstallError{Err: err}
	}
	if descVersion.Core().Compare(version.Core()) != 0 {
		err := fmt.Errorf("binary reported version (%q) is different from the expected %q, skipping", desc.Version, version.String())
		return &ContinuableInstallError{Err: err}
	}
	if version.Prerelease() != "" {
		return &PrereleaseInstallError{
			PluginSrc: identifier,
			Err:       errors.New("binary reported a pre-release version of " + version.String()),
		}
	}
	// Since only final releases can be installed remotely, a non-empty prerelease version
	// means something's not right on the release, as it should report a final version.
	//
	// Therefore to avoid surprises (and avoid being able to install a version that
	// cannot be loaded), we error here, and advise users to manually install the plugin if they
	// need it.
	if descVersion.Prerelease() != "" {
		return &PrereleaseInstallError{
			PluginSrc: identifier,
			Err:       errors.New("binary reported a pre-release version of " + descVersion.String()),
		}
	}
	return nil
}

func init() {
	var err error
	// Should never error if both components are set
	localAPIVersion, err = NewAPIVersion(fmt.Sprintf("x%s.%s", pluginsdk.APIVersionMajor, pluginsdk.APIVersionMinor))
	if err != nil {
		panic(fmt.Sprintf("malformed API version in Packer. This is a programming error, please open an error to report it."))
	}
}
