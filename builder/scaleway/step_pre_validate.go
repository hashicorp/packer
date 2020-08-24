package scaleway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type stepPreValidate struct {
	Force        bool
	ImageName    string
	SnapshotName string
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.Force {
		ui.Say("Force flag found, skipping prevalidating image name")
		return multistep.ActionContinue
	}

	client := state.Get("client").(*api.ScalewayAPI)
	ui.Say(fmt.Sprintf("Prevalidating image name: %s", s.ImageName))

	images, err := getImages(client)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, im := range images.Images {
		if im.Name == s.ImageName {
			err := fmt.Errorf("Error: image name: '%s' is used by existing image with ID %s",
				s.ImageName, im.Identifier)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say(fmt.Sprintf("Prevalidating snapshot name: %s", s.SnapshotName))

	snapshots, err := client.GetSnapshots()
	if err != nil {
		err := fmt.Errorf("Error: getting snapshot list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, sn := range *snapshots {
		if sn.Name == s.SnapshotName {
			err := fmt.Errorf("Error: snapshot name: '%s' is used by existing snapshot with ID %s",
				s.SnapshotName, sn.Identifier)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {
}

// getImages returns a list of all the images in region belonging to the organization
// configured in client.
// The (deprecated) package github.com/scaleway/scaleway-cli/pkg/api used by Packer
// doesn't have this function (it has instead a confusing GetImages() that returns all
// the market images plus the private images, but the private images are missing the
// majority of fields...).
func getImages(client *api.ScalewayAPI) (*api.ScalewayImages, error) {
	var computeAPI string
	switch client.Region {
	case "par1":
		computeAPI = api.ComputeAPIPar1
	case "ams1":
		computeAPI = api.ComputeAPIAms1
	}

	values := url.Values{}
	values.Set("organization", client.Organization)
	resp, err := client.GetResponsePaginate(computeAPI, "images", values)
	if err != nil {
		return nil, fmt.Errorf("getting image list: %s", err)
	}
	defer resp.Body.Close()

	body, err := handleHTTPError(resp)
	if err != nil {
		return nil, fmt.Errorf("reading image list body: %s", err)
	}

	var images api.ScalewayImages
	if err = json.Unmarshal(body, &images); err != nil {
		err = fmt.Errorf("parsing image list: %s", err)
		return nil, err
	}
	return &images, nil
}

// Copied from scaleway-cli@v0.0.0-20180921094345-7b12c9699d70/pkg/api/api.go because
// not exported.
// handleHTTPError checks the statusCode and displays the error
func handleHTTPError(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, errors.New(string(body))
	}
	if resp.StatusCode != http.StatusOK {
		var scwError api.ScalewayAPIError
		if err := json.Unmarshal(body, &scwError); err != nil {
			return nil, err
		}
		scwError.StatusCode = resp.StatusCode
		return nil, scwError
	}
	return body, nil
}
