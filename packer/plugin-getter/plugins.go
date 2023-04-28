package plugingetter

import (
	"archive/zip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/hcl2template/addrs"
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
	VersionConstraints version.Constraints

	// was this require implicitly guessed ?
	Implicit bool
}

type BinaryInstallationOptions struct {
	//
	APIVersionMajor, APIVersionMinor string
	// OS and ARCH usually should be runtime.GOOS and runtime.ARCH, they allow
	// to pick the correct binary.
	OS, ARCH string

	// Ext is ".exe" on windows
	Ext string

	Checksummers []Checksummer
}

type ListInstallationsOptions struct {
	// FromFolders where plugins could be installed. Paths should be absolute for
	// safety but can also be relative.
	FromFolders []string

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

func (pr Requirement) FilenamePrefix() string {
	if pr.Identifier == nil {
		return "packer-plugin-"
	}
	return "packer-plugin-" + pr.Identifier.Type + "_"
}

func (opts BinaryInstallationOptions) FilenameSuffix() string {
	return "_" + opts.OS + "_" + opts.ARCH + opts.Ext
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
	FilenamePrefix := pr.FilenamePrefix()
	filenameSuffix := opts.FilenameSuffix()
	log.Printf("[TRACE] listing potential installations for %q that match %q. %#v", pr.Identifier, pr.VersionConstraints, opts)
	for _, knownFolder := range opts.FromFolders {
		glob := ""
		if pr.Identifier == nil {
			glob = filepath.Join(knownFolder, "*", "*", "*", FilenamePrefix+"*"+filenameSuffix)
		} else {
			glob = filepath.Join(knownFolder, pr.Identifier.Hostname, pr.Identifier.Namespace, pr.Identifier.Type, FilenamePrefix+"*"+filenameSuffix)
		}

		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("ListInstallations: %q failed to list binaries in folder: %v", pr.Identifier.String(), err)
		}
		for _, path := range matches {
			fname := filepath.Base(path)
			if fname == "." {
				continue
			}

			// base name could look like packer-plugin-amazon_v1.2.3_x5.1_darwin_amd64.exe
			versionsStr := strings.TrimPrefix(fname, FilenamePrefix)
			versionsStr = strings.TrimSuffix(versionsStr, filenameSuffix)

			if pr.Identifier == nil {
				if idx := strings.Index(versionsStr, "_"); idx > 0 {
					versionsStr = versionsStr[idx+1:]
				}
			}

			// versionsStr now looks like v1.2.3_x5.1 or amazon_v1.2.3_x5.1
			parts := strings.SplitN(versionsStr, "_", 2)
			pluginVersionStr, protocolVerionStr := parts[0], parts[1]
			pv, err := version.NewVersion(pluginVersionStr)
			if err != nil {
				// could not be parsed, ignoring the file
				log.Printf("found %q with an incorrect %q version, ignoring it. %v", path, pluginVersionStr, err)
				continue
			}

			// no constraint means always pass, this will happen for implicit
			// plugin requirements and when we list all plugins.
			if !pr.VersionConstraints.Check(pv) {
				log.Printf("[TRACE] version %q of file %q does not match constraint %q", pluginVersionStr, path, pr.VersionConstraints.String())
				continue
			}

			if err := opts.CheckProtocolVersion(protocolVerionStr); err != nil {
				log.Printf("[NOTICE] binary %s requires protocol version %s that is incompatible "+
					"with this version of Packer. %s", path, protocolVerionStr, err)
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

			res.InsertSortedUniq(&Installation{
				BinaryPath: path,
				Version:    pluginVersionStr,
			})
		}
	}
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

// InsertSortedUniq inserts the installation in the right spot in the list by
// comparing the version lexicographically.
// A Duplicate version will replace any already present version.
func (l *InstallList) InsertSortedUniq(install *Installation) {
	pos := sort.Search(len(*l), func(i int) bool { return (*l)[i].Version >= install.Version })
	if len(*l) > pos && (*l)[pos].Version == install.Version {
		// already detected, let's ignore any new foundings, this way any plugin
		// close to cwd or the packer exec takes precedence; this will be better
		// for plugin development/tests.
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
	//  * v1.2.3 for packer-plugin-amazon_v1.2.3_darwin_x5
	//  * empty  for packer-plugin-amazon
	Version string
}

// InstallOptions describes the possible options for installing the plugin that
// fits the plugin Requirement.
type InstallOptions struct {
	// Different means to get releases, sha256 and binary files.
	Getters []Getter

	// Any downloaded binary and checksum file will be put in the last possible
	// folder of this list.
	InFolders []string

	BinaryInstallationOptions
}

type GetOptions struct {
	PluginRequirement *Requirement

	BinaryInstallationOptions

	version *version.Version

	expectedZipFilename string
}

// ExpectedZipFilename is the filename of the zip we expect to find, the
// value is known only after parsing the checksum file file.
func (gp *GetOptions) ExpectedZipFilename() string {
	return gp.expectedZipFilename
}

func (binOpts *BinaryInstallationOptions) CheckProtocolVersion(remoteProt string) error {
	remoteProt = strings.TrimPrefix(remoteProt, "x")
	parts := strings.Split(remoteProt, ".")
	if len(parts) < 2 {
		return fmt.Errorf("Invalid remote protocol: %q, expected something like '%s.%s'", remoteProt, binOpts.APIVersionMajor, binOpts.APIVersionMinor)
	}
	vMajor, vMinor := parts[0], parts[1]

	// no protocol version check
	if binOpts.APIVersionMajor == "" && binOpts.APIVersionMinor == "" {
		return nil
	}

	if vMajor != binOpts.APIVersionMajor {
		return fmt.Errorf("Unsupported remote protocol MAJOR version %q. The current MAJOR protocol version is %q."+
			" This version of Packer can only communicate with plugins using that version.", vMajor, binOpts.APIVersionMajor)
	}

	if vMinor == binOpts.APIVersionMinor {
		return nil
	}

	vMinori, err := strconv.Atoi(vMinor)
	if err != nil {
		return err
	}

	APIVersoinMinori, err := strconv.Atoi(binOpts.APIVersionMinor)
	if err != nil {
		return err
	}

	if vMinori > APIVersoinMinori {
		return fmt.Errorf("Unsupported remote protocol MINOR version %q. The supported MINOR protocol versions are version %q and bellow."+
			"Please upgrade Packer or use an older version of the plugin if possible.", vMinor, binOpts.APIVersionMinor)
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
	versions := version.Collection{}
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
			v, err := version.NewVersion(release.Version)
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
			opts.InFolders[len(opts.InFolders)-1],
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
						err := fmt.Errorf("ignoring invalid remote binary %s: %s", entry.Filename, err)
						errs = multierror.Append(errs, err)
						log.Printf("[TRACE] %s", err)
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

					for _, outputFolder := range opts.InFolders {
						potentialOutputFilename := filepath.Join(
							outputFolder,
							filepath.Join(pr.Identifier.Parts()...),
							expectedBinaryFilename,
						)
						for _, potentialChecksumer := range opts.Checksummers {
							// First check if a local checksum file is already here in the expected
							// download folder. Here we want to download a binary so we only check
							// for an existing checksum file from the folder we want to download
							// into.
							cs, err := potentialChecksumer.GetCacheChecksumOfFile(potentialOutputFilename)
							if err == nil && len(cs) > 0 {
								localChecksum := &FileChecksum{
									Expected:    cs,
									Checksummer: potentialChecksumer,
								}

								log.Printf("[TRACE] found a pre-exising %q checksum file", potentialChecksumer.Type)
								// if outputFile is there and matches the checksum: do nothing more.
								if err := localChecksum.ChecksumFile(localChecksum.Expected, potentialOutputFilename); err == nil {
									log.Printf("[INFO] %s v%s plugin is already correctly installed in %q", pr.Identifier, version, potentialOutputFilename)
									return nil, nil // success
								}
							}
						}
					}

					// The last folder from the installation list is where we will install.
					outputFileName := filepath.Join(outputFolder, expectedBinaryFilename)

					// create directories if need be
					if err := os.MkdirAll(outputFolder, 0755); err != nil {
						err := fmt.Errorf("could not create plugin folder %q: %w", outputFolder, err)
						errs = multierror.Append(errs, err)
						log.Printf("[TRACE] %s", err.Error())
						return nil, errs
					}

					for _, getter := range getters {
						// create temporary file that will receive a temporary binary.zip
						tmpFile, err := tmp.File("packer-plugin-*.zip")
						if err != nil {
							err = fmt.Errorf("could not create temporary file to dowload plugin: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						defer tmpFile.Close()

						// start fetching binary
						remoteZipFile, err := getter.Get("zip", GetOptions{
							PluginRequirement:         pr,
							BinaryInstallationOptions: opts.BinaryInstallationOptions,
							version:                   version,
							expectedZipFilename:       expectedZipFilename,
						})
						if err != nil {
							err := fmt.Errorf("could not get binary for %s version %s. Is the file present on the release and correctly named ? %s", pr.Identifier, version, err)
							errs = multierror.Append(errs, err)
							log.Printf("[TRACE] %v", err)
							continue
						}

						// write binary to tmp file
						_, err = io.Copy(tmpFile, remoteZipFile)
						_ = remoteZipFile.Close()
						if err != nil {
							err := fmt.Errorf("Error getting plugin, trying another getter: %w", err)
							errs = multierror.Append(errs, err)
							log.Printf("[TRACE] %s", err)
							continue
						}

						if _, err := tmpFile.Seek(0, 0); err != nil {
							err := fmt.Errorf("Error seeking begining of temporary file for checksumming, continuing: %w", err)
							errs = multierror.Append(errs, err)
							log.Printf("[TRACE] %s", err)
							continue
						}

						// verify that the checksum for the zip is what we expect.
						if err := checksum.Checksummer.Checksum(checksum.Expected, tmpFile); err != nil {
							err := fmt.Errorf("%w. Is the checksum file correct ? Is the binary file correct ?", err)
							errs = multierror.Append(errs, err)
							log.Printf("%s, truncating the zipfile", err)
							if err := tmpFile.Truncate(0); err != nil {
								log.Printf("[TRACE] %v", err)
							}
							continue
						}

						tmpFileStat, err := tmpFile.Stat()
						if err != nil {
							err := fmt.Errorf("failed to stat: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}

						zr, err := zip.NewReader(tmpFile, tmpFileStat.Size())
						if err != nil {
							err := fmt.Errorf("zip : %v", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}

						var copyFrom io.ReadCloser
						for _, f := range zr.File {
							if f.Name != expectedBinaryFilename {
								continue
							}
							copyFrom, err = f.Open()
							if err != nil {
								err := fmt.Errorf("failed to open temp file: %w", err)
								errs = multierror.Append(errs, err)
								return nil, errs
							}
							break
						}
						if copyFrom == nil {
							err := fmt.Errorf("could not find a %s file in zipfile", checksum.Filename)
							errs = multierror.Append(errs, err)
							return nil, errs
						}

						outputFile, err := os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
						if err != nil {
							err := fmt.Errorf("failed to create %s: %w", outputFileName, err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}
						defer outputFile.Close()

						if _, err := io.Copy(outputFile, copyFrom); err != nil {
							err := fmt.Errorf("extract file: %w", err)
							errs = multierror.Append(errs, err)
							return nil, errs
						}

						if _, err := outputFile.Seek(0, 0); err != nil {
							err := fmt.Errorf("Error seeking begining of binary file for checksumming: %w", err)
							errs = multierror.Append(errs, err)
							log.Printf("[WARNING] %v, ignoring", err)
						}

						cs, err := checksum.Checksummer.Sum(outputFile)
						if err != nil {
							err := fmt.Errorf("failed to checksum binary file: %s", err)
							errs = multierror.Append(errs, err)
							log.Printf("[WARNING] %v, ignoring", err)
						}

						if err := ioutil.WriteFile(outputFileName+checksum.Checksummer.FileExt(), []byte(hex.EncodeToString(cs)), 0555); err != nil {
							err := fmt.Errorf("failed to write local binary checksum file: %s", err)
							errs = multierror.Append(errs, err)
							log.Printf("[WARNING] %v, ignoring", err)
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

	if errs.Len() == 0 {
		err := fmt.Errorf("could not find a local nor a remote checksum for plugin %q %q", pr.Identifier, pr.VersionConstraints)
		errs = multierror.Append(errs, err)
	}

	errs = multierror.Append(errs, fmt.Errorf("could not install any compatible version of plugin %q", pr.Identifier))

	return nil, errs
}
