package common

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	getter "github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/packer/common/filelock"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepDownload downloads a remote file using the download client within
// this package. This step handles setting up the download configuration,
// progress reporting, interrupt handling, etc.
//
// Uses:
//   cache packer.Cache
//   ui    packer.Ui
type StepDownload struct {
	// The checksum and the type of the checksum for the download
	Checksum     string
	ChecksumType string

	// A short description of the type of download being done. Example:
	// "ISO" or "Guest Additions"
	Description string

	// The name of the key where the final path of the ISO will be put
	// into the state.
	ResultKey string

	// The path where the result should go, otherwise it goes to the
	// cache directory.
	TargetPath string

	// A list of URLs to attempt to download this thing.
	Url []string

	// Extension is the extension to force for the file that is downloaded.
	// Some systems require a certain extension. If this isn't set, the
	// extension on the URL is used. Otherwise, this will be forced
	// on the downloaded file for every URL.
	Extension string
}

func (s *StepDownload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Url) == 0 {
		log.Printf("No URLs were provided to Step Download. Continuing...")
		return multistep.ActionContinue
	}

	defer log.Printf("Leaving retrieve loop for %s", s.Description)

	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Retrieving %s", s.Description))

	var errs []error

	for _, source := range s.Url {
		if ctx.Err() != nil {
			state.Put("error", fmt.Errorf("Download cancelled: %v", errs))
			return multistep.ActionHalt
		}
		ui.Say(fmt.Sprintf("Trying %s", source))
		var err error
		var dst string
		if s.Description == "OVF/OVA" && strings.HasSuffix(source, ".ovf") {
			// TODO(adrien): make go-getter allow using files in place.
			// ovf files usually point to a file in the same directory, so
			// using them in place is the only way.
			ui.Say(fmt.Sprintf("Using ovf inplace"))
			dst = source
		} else {
			dst, err = s.download(ctx, ui, source)
		}
		if err == nil {
			state.Put(s.ResultKey, dst)
			return multistep.ActionContinue
		}
		// may be another url will work
		errs = append(errs, err)
	}

	err := fmt.Errorf("error downloading %s: %v", s.Description, errs)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

var (
	getters = getter.Getters
)

func init() {
	if runtime.GOOS == "windows" {
		getters["file"] = &getter.FileGetter{
			// always copy local files instead of symlinking to fix GH-7534. The
			// longer term fix for this would be to change the go-getter so that it
			// can leave the source file where it is & tell us where it is.
			Copy: true,
		}
		getters["smb"] = &getter.FileGetter{
			Copy: true,
		}
	}
}

func (s *StepDownload) download(ctx context.Context, ui packer.Ui, source string) (string, error) {
	if runtime.GOOS == "windows" {
		// Check that the user specified a UNC path, and promote it to an smb:// uri.
		if strings.HasPrefix(source, "\\\\") && len(source) > 2 && source[2] != '?' {
			source = filepath.ToSlash(source[2:])
			source = fmt.Sprintf("smb://%s", source)
		}
	}

	u, err := urlhelper.Parse(source)
	if err != nil {
		return "", fmt.Errorf("url parse: %s", err)
	}
	if checksum := u.Query().Get("checksum"); checksum != "" {
		s.Checksum = checksum
	}
	if s.ChecksumType != "" && s.ChecksumType != "none" {
		// add checksum to url query params as go getter will checksum for us
		q := u.Query()
		q.Set("checksum", s.ChecksumType+":"+s.Checksum)
		u.RawQuery = q.Encode()
	} else if s.Checksum != "" {
		q := u.Query()
		q.Set("checksum", s.Checksum)
		u.RawQuery = q.Encode()
	}

	targetPath := s.TargetPath
	if targetPath == "" {
		// store file under sha1(hash) if set
		// hash can sometimes be a checksum url
		// otherwise, use sha1(source_url)
		var shaSum [20]byte
		if s.Checksum != "" {
			shaSum = sha1.Sum([]byte(s.Checksum))
		} else {
			shaSum = sha1.Sum([]byte(u.String()))
		}
		targetPath = hex.EncodeToString(shaSum[:])
		if s.Extension != "" {
			targetPath += "." + s.Extension
		}
		targetPath, err = packer.CachePath(targetPath)
		if err != nil {
			return "", fmt.Errorf("CachePath: %s", err)
		}
	}

	lockFile := targetPath + ".lock"

	log.Printf("Acquiring lock for: %s (%s)", u.String(), lockFile)
	lock := filelock.New(lockFile)
	lock.Lock()
	defer lock.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("get working directory: %v", err)
		// here we ignore the error in case the
		// working directory is not needed.
		// It would be better if the go-getter
		// could guess it only in cases it is
		// necessary.
	}
	src := u.String()
	if u.Scheme == "" || strings.ToLower(u.Scheme) == "file" {
		// If a local filepath, then we need to preprocess to make sure the
		// path doens't have any multiple successive path separators; if it
		// does, go-getter will read this as a specialized go-getter-specific
		// subdirectory command, which it most likely isn't.
		src = filepath.Clean(u.String())
		if _, err := os.Stat(filepath.Clean(u.Path)); err != nil {
			// Cleaned path isn't present on system so it must be some other
			// scheme. Don't error right away; see if go-getter can figure it
			// out.
			src = u.String()
		}
	}

	ui.Say(fmt.Sprintf("Trying %s", u.String()))
	gc := getter.Client{
		Ctx:              ctx,
		Dst:              targetPath,
		Src:              src,
		ProgressListener: ui,
		Pwd:              wd,
		Dir:              false,
		Getters:          getters,
	}

	switch err := gc.Get(); err.(type) {
	case nil: // success !
		ui.Say(fmt.Sprintf("%s => %s", u.String(), targetPath))
		return targetPath, nil
	case *getter.ChecksumError:
		ui.Say(fmt.Sprintf("Checksum did not match, removing %s", targetPath))
		if err := os.Remove(targetPath); err != nil {
			ui.Error(fmt.Sprintf("Failed to remove cache file. Please remove manually: %s", targetPath))
		}
		return "", err
	default:
		ui.Say(fmt.Sprintf("Download failed %s", err))
		return "", err
	}
}

func (s *StepDownload) Cleanup(multistep.StateBag) {}
