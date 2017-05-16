// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"golang.org/x/oauth2"
)

// The unique id for the builder
const BuilderId = "pearkes.digitalocean"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = *c

	return nil, nil
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	client := godo.NewClient(oauth2.NewClient(oauth2.NoContext, &apiTokenSource{
		AccessToken: b.config.APIToken,
	}))
	if b.config.APIURL != "" {
		u, err := url.Parse(b.config.APIURL)
		if err != nil {
			return nil, fmt.Errorf("DigitalOcean: Invalid API URL, %s.", err)
		}
		client.BaseURL = u
	}

	if len(b.config.SnapshotRegions) > 0 {
		opt := &godo.ListOptions{
			Page:    1,
			PerPage: 200,
		}
		regions, _, err := client.Regions.List(context.TODO(), opt)
		if err != nil {
			return nil, fmt.Errorf("DigitalOcean: Unable to get regions, %s", err)
		}

		var dcs []string
		for _, val := range regions {
			dcs = append(dcs, val.Slug)
		}

		regionSet := make(map[string]struct{})
		regionsMap := make([]string, 0, len(b.config.SnapshotRegions))
		regionSet[b.config.Region] = struct{}{}
		for _, region := range b.config.SnapshotRegions {
			// If we already saw the region, then don't look again
			if _, ok := regionSet[region]; ok {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = struct{}{}

			regionsMap = append(regionsMap, region)
		}

		for _, val := range regionsMap {
			if contains(dcs, val) {
				continue
			} else {
				return nil, fmt.Errorf("DigitalOcean: Invalid region, %s", val)
			}
		}
	}

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("do_%s.pem", b.config.PackerBuildName),
		},
		new(stepCreateDroplet),
		new(stepDropletInfo),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
		},
		new(common.StepProvision),
		new(stepShutdown),
		new(stepPowerOff),
		new(stepSnapshot),
	}

	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		log.Println("Failed to find snapshot_name in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		snapshotName: state.Get("snapshot_name").(string),
		snapshotId:   state.Get("snapshot_image_id").(int),
		regionNames:  state.Get("regions").([]string),
		client:       client,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
