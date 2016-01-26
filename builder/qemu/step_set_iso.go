package qemu

import (
	"fmt"
	"net/http"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step set iso_patch to available url
type stepSetISO struct {
	ResultKey string
	Url       []string
}

func (s *stepSetISO) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	iso_path := ""

	for _, url := range s.Url {
		req, err := http.NewRequest("HEAD", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "Packer")

		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}

		res, err := httpClient.Do(req)
		if err == nil && (res.StatusCode >= 200 && res.StatusCode < 300) {
			if res.Header.Get("Accept-Ranges") == "bytes" {
				iso_path = url
			}
		}
	}

	if iso_path == "" {
		err := fmt.Errorf("No byte serving support. The HTTP server must support Accept-Ranges=bytes")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, iso_path)

	return multistep.ActionContinue
}

func (s *stepSetISO) Cleanup(state multistep.StateBag) {}
