package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

var additionsVersionMap = map[string]string{
	"4.2.1":  "4.2.0",
	"4.1.23": "4.1.22",
}

type guestAdditionsUrlTemplate struct {
	Version string
}

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
//
// Produces:
//   guest_additions_path string - Path to the guest additions.
type StepDownloadGuestAdditions struct {
	GuestAdditionsMode   string
	GuestAdditionsURL    string
	GuestAdditionsSHA256 string
	Ctx                  interpolate.Context
}

func (s *StepDownloadGuestAdditions) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var action multistep.StepAction
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// If we've disabled guest additions, don't download
	if s.GuestAdditionsMode == GuestAdditionsModeDisable {
		log.Println("Not downloading guest additions since it is disabled.")
		return multistep.ActionContinue
	}

	// Get VBox version
	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for guest additions download: %s", err))
		return multistep.ActionHalt
	}

	if newVersion, ok := additionsVersionMap[version]; ok {
		log.Printf("Rewriting guest additions version: %s to %s", version, newVersion)
		version = newVersion
	}

	additionsName := fmt.Sprintf("VBoxGuestAdditions_%s.iso", version)

	// Use provided version or get it from virtualbox.org
	var checksum string

	checksumType := "sha256"

	// Initialize the template context so we can interpolate some variables..
	s.Ctx.Data = &guestAdditionsUrlTemplate{
		Version: version,
	}

	// Interpolate any user-variables specified within the guest_additions_url
	url, err := interpolate.Render(s.GuestAdditionsURL, &s.Ctx)
	if err != nil {
		err := fmt.Errorf("Error preparing guest additions url: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// If this resulted in an empty url, then ask the driver about it.
	if url == "" {
		log.Printf("guest_additions_url is blank; querying driver for iso.")
		url, err = driver.Iso()

		if err == nil {
			checksumType = "none"
		} else {
			ui.Error(err.Error())
			url = fmt.Sprintf(
				"https://download.virtualbox.org/virtualbox/%s/%s",
				version,
				additionsName)
		}
	}

	// The driver couldn't even figure it out, so fail hard.
	if url == "" {
		err := fmt.Errorf("Couldn't detect guest additions URL.\n" +
			"Please specify `guest_additions_url` manually.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Figure out a default checksum here
	if checksumType != "none" {
		if s.GuestAdditionsSHA256 != "" {
			checksum = s.GuestAdditionsSHA256
		} else {
			checksum, action = s.downloadAdditionsSHA256(ctx, state, version, additionsName)
			if action != multistep.ActionContinue {
				return action
			}
		}
	}

	log.Printf("Guest additions URL: %s", url)

	// We're good, so let's go ahead and download this thing..
	downStep := &commonsteps.StepDownload{
		Checksum:    checksum,
		Description: "Guest additions",
		ResultKey:   "guest_additions_path",
		Url:         []string{url},
		Extension:   "iso",
	}

	return downStep.Run(ctx, state)
}

func (s *StepDownloadGuestAdditions) Cleanup(state multistep.StateBag) {}

func (s *StepDownloadGuestAdditions) downloadAdditionsSHA256(ctx context.Context, state multistep.StateBag, additionsVersion string, additionsName string) (string, multistep.StepAction) {
	// First things first, we get the list of checksums for the files available
	// for this version.
	checksumsUrl := fmt.Sprintf(
		"https://download.virtualbox.org/virtualbox/%s/SHA256SUMS",
		additionsVersion)

	checksumsFile, err := tmp.File("packer")
	if err != nil {
		state.Put("error", fmt.Errorf(
			"Failed creating temporary file to store guest addition checksums: %s",
			err))
		return "", multistep.ActionHalt
	}
	defer os.Remove(checksumsFile.Name())
	checksumsFile.Close()

	downStep := &commonsteps.StepDownload{
		Description: "Guest additions checksums",
		ResultKey:   "guest_additions_checksums_path",
		Url:         []string{checksumsUrl},
	}

	action := downStep.Run(ctx, state)
	if action == multistep.ActionHalt {
		return "", action
	}

	// Next, we find the checksum for the file we're looking to download.
	// It is an error if the checksum cannot be found.
	checksumsF, err := os.Open(state.Get("guest_additions_checksums_path").(string))
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening guest addition checksums: %s", err))
		return "", multistep.ActionHalt
	}
	defer checksumsF.Close()

	// We copy the contents of the file into memory. In general this file
	// is quite small so that is okay. In the future, we probably want to
	// use bufio and iterate line by line.
	var contents bytes.Buffer
	io.Copy(&contents, checksumsF)

	checksum := ""
	for _, line := range strings.Split(contents.String(), "\n") {
		parts := strings.Fields(line)
		log.Printf("Checksum file parts: %#v", parts)
		if len(parts) != 2 {
			// Bogus line
			continue
		}

		if strings.HasSuffix(parts[1], additionsName) {
			checksum = parts[0]
			log.Printf("Guest additions checksum: %s", checksum)
			break
		}
	}

	if checksum == "" {
		state.Put("error", fmt.Errorf(
			"The checksum for the file '%s' could not be found.", additionsName))
		return "", multistep.ActionHalt
	}

	return checksum, multistep.ActionContinue

}
