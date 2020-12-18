package commonsteps

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	gcs "github.com/hashicorp/go-getter/gcs/v2"
	s3 "github.com/hashicorp/go-getter/s3/v2"
	getter "github.com/hashicorp/go-getter/v2"
	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"

	"github.com/hashicorp/packer-plugin-sdk/filelock"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepDownload downloads a remote file using the download client within
// this package. This step handles setting up the download configuration,
// progress reporting, interrupt handling, etc.
//
// Uses:
//   cache packer.Cache
//   ui    packersdk.Ui
type StepDownload struct {
	// The checksum and the type of the checksum for the download
	Checksum string

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

var defaultGetterClient = getter.Client{
	Getters: getter.Getters,
}

func init() {
	defaultGetterClient.Getters = append(defaultGetterClient.Getters, new(gcs.Getter))
	defaultGetterClient.Getters = append(defaultGetterClient.Getters, new(s3.Getter))
}

func (s *StepDownload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Url) == 0 {
		log.Printf("No URLs were provided to Step Download. Continuing...")
		return multistep.ActionContinue
	}

	defer log.Printf("Leaving retrieve loop for %s", s.Description)

	ui := state.Get("ui").(packersdk.Ui)
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

func (s *StepDownload) UseSourceToFindCacheTarget(source string) (*url.URL, string, error) {
	u, err := parseSourceURL(source)
	if err != nil {
		return nil, "", fmt.Errorf("url parse: %s", err)
	}
	if checksum := u.Query().Get("checksum"); checksum != "" {
		s.Checksum = checksum
	}
	if s.Checksum != "" && s.Checksum != "none" {
		// add checksum to url query params as go getter will checksum for us
		q := u.Query()
		q.Set("checksum", s.Checksum)
		u.RawQuery = q.Encode()
	}

	// store file under sha1(hash) if set
	// hash can sometimes be a checksum url
	// otherwise, use sha1(source_url)
	var shaSum [20]byte
	if s.Checksum != "" && s.Checksum != "none" {
		shaSum = sha1.Sum([]byte(s.Checksum))
	} else {
		shaSum = sha1.Sum([]byte(u.String()))
	}
	shaSumString := hex.EncodeToString(shaSum[:])

	targetPath := s.TargetPath
	if targetPath == "" {
		targetPath = shaSumString
		if s.Extension != "" {
			targetPath += "." + s.Extension
		}
		targetPath, err = packersdk.CachePath(targetPath)
		if err != nil {
			return nil, "", fmt.Errorf("CachePath: %s", err)
		}
	} else if filepath.Ext(targetPath) == "" {
		// When an absolute path is provided
		// this adds the file to the targetPath
		if !strings.HasSuffix(targetPath, "/") {
			targetPath += "/"
		}
		targetPath += shaSumString
		if s.Extension != "" {
			targetPath += "." + s.Extension
		} else {
			targetPath += ".iso"
		}
	}
	return u, targetPath, nil
}

func (s *StepDownload) download(ctx context.Context, ui packersdk.Ui, source string) (string, error) {
	u, targetPath, err := s.UseSourceToFindCacheTarget(source)
	if err != nil {
		return "", err
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
	req := &getter.Request{
		Dst:              targetPath,
		Src:              src,
		ProgressListener: ui,
		Pwd:              wd,
		Mode:             getter.ModeFile,
		Inplace:          true,
	}

	switch op, err := defaultGetterClient.Get(ctx, req); err.(type) {
	case nil: // success !
		ui.Say(fmt.Sprintf("%s => %s", u.String(), op.Dst))
		return op.Dst, nil
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

func parseSourceURL(source string) (*url.URL, error) {
	if runtime.GOOS == "windows" {
		// Check that the user specified a UNC path, and promote it to an smb:// uri.
		if strings.HasPrefix(source, "\\\\") && len(source) > 2 && source[2] != '?' {
			source = filepath.ToSlash(source[2:])
			source = fmt.Sprintf("smb://%s", source)
		}
	}

	u, err := urlhelper.Parse(source)
	return u, err
}

func (s *StepDownload) Cleanup(multistep.StateBag) {}
