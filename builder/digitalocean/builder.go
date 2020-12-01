// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"golang.org/x/oauth2"
)

// The unique id for the builder
const BuilderId = "pearkes.digitalocean"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	client := godo.NewClient(oauth2.NewClient(context.TODO(), &apiTokenSource{
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

		validRegions := make(map[string]struct{})
		for _, val := range regions {
			validRegions[val.Slug] = struct{}{}
		}

		for _, region := range append(b.config.SnapshotRegions, b.config.Region) {
			if _, ok := validRegions[region]; !ok {
				return nil, fmt.Errorf("DigitalOcean: Invalid region, %s", region)
			}
		}
	}

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
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
			Host:      communicator.CommHost(b.config.Comm.Host(), "droplet_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(commonsteps.StepProvision),
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		new(stepShutdown),
		new(stepPowerOff),
		&stepSnapshot{
			snapshotTimeout: b.config.SnapshotTimeout,
		},
	}

	// Run the steps
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		log.Println("Failed to find snapshot_name in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		SnapshotName: state.Get("snapshot_name").(string),
		SnapshotId:   state.Get("snapshot_image_id").(int),
		RegionNames:  state.Get("regions").([]string),
		Client:       client,
		StateData:    map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}
