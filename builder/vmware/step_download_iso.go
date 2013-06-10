package vmware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// This step downloads the ISO specified.
//
// Uses:
//   cache packer.Cache
//   config *config
//   ui     packer.Ui
//
// Produces:
//   iso_path string
type stepDownloadISO struct{}

func (s stepDownloadISO) Run(state map[string]interface{}) multistep.StepAction {
	cache := state["cache"].(packer.Cache)
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	log.Printf("Acquiring lock to download the ISO.")
	cachePath := cache.Lock(config.ISOUrl)
	defer cache.Unlock(config.ISOUrl)

	err := s.checkMD5(cachePath, config.ISOMD5)
	haveFile := err == nil
	if err != nil {
		if !os.IsNotExist(err) {
			ui.Say(fmt.Sprintf("Error validating MD5 of ISO: %s", err))
			return multistep.ActionHalt
		}
	}

	if !haveFile {
		url, err := url.Parse(config.ISOUrl)
		if err != nil {
			ui.Error(fmt.Sprintf("Error parsing iso_url: %s", err))
			return multistep.ActionHalt
		}

		// Start the download in a goroutine so that we cancel it and such.
		var progress uint
		downloadComplete := make(chan bool, 1)
		go func() {
			ui.Say("Copying or downloading ISO. Progress will be shown periodically.")
			cachePath, err = s.downloadUrl(cachePath, url, &progress)
			downloadComplete <- true
		}()

		progressTimer := time.NewTicker(15 * time.Second)
		defer progressTimer.Stop()

	DownloadWaitLoop:
		for {
			select {
			case <-downloadComplete:
				log.Println("Download of ISO completed.")
				break DownloadWaitLoop
			case <-progressTimer.C:
				ui.Say(fmt.Sprintf("Download progress: %d%%", progress))
			case <-time.After(1 * time.Second):
				if _, ok := state[multistep.StateCancelled]; ok {
					ui.Say("Interrupt received. Cancelling download...")
					return multistep.ActionHalt
				}
			}
		}

		if err != nil {
			ui.Error(fmt.Sprintf("Error downloading ISO: %s", err))
			return multistep.ActionHalt
		}

		if err = s.checkMD5(cachePath, config.ISOMD5); err != nil {
			ui.Say(fmt.Sprintf("Error validating MD5 of ISO: %s", err))
			return multistep.ActionHalt
		}
	}

	log.Printf("Path to ISO on disk: %s", cachePath)
	state["iso_path"] = cachePath

	return multistep.ActionContinue
}

func (stepDownloadISO) Cleanup(map[string]interface{}) {}

func (stepDownloadISO) checkMD5(path string, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	hash := md5.New()
	io.Copy(hash, f)
	result := strings.ToLower(hex.EncodeToString(hash.Sum(nil)))
	if result != expected {
		return fmt.Errorf("result != expected: %s != %s", result, expected)
	}

	return nil
}

func (stepDownloadISO) downloadUrl(path string, url *url.URL, progress *uint) (string, error) {
	if url.Scheme == "file" {
		// If it is just a file URL, then we already have the ISO
		return url.Path, nil
	}

	// Otherwise, it is an HTTP URL, and we must download it.
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	log.Printf("Beginning download of ISO: %s", url.String())
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	var buffer [4096]byte
	var totalRead int64
	for {
		n, err := resp.Body.Read(buffer[:])
		if err != nil && err != io.EOF {
			return "", err
		}

		totalRead += int64(n)
		*progress = uint((float64(totalRead) / float64(resp.ContentLength)) * 100)

		if _, werr := f.Write(buffer[:n]); werr != nil {
			return "", werr
		}

		if err == io.EOF {
			break
		}
	}

	return path, nil
}
