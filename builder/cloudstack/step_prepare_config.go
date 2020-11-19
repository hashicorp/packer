package cloudstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepPrepareConfig struct{}

func (s *stepPrepareConfig) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Preparing config...")

	var err error
	var errs *packersdk.MultiError

	// First get the project and zone UUID's so we can use them in other calls when needed.
	if config.Project != "" && !isUUID(config.Project) {
		config.Project, _, err = client.Project.GetProjectID(config.Project)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"project", config.Project, err})
		}
	}

	if config.UserDataFile != "" {
		userdata, err := ioutil.ReadFile(config.UserDataFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("problem reading user data file: %s", err))
		}
		config.UserData = string(userdata)
	}

	if !isUUID(config.Zone) {
		config.Zone, _, err = client.Zone.GetZoneID(config.Zone)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"zone", config.Zone, err})
		}
	}

	// Then try to get the remaining UUID's.
	if config.DiskOffering != "" && !isUUID(config.DiskOffering) {
		config.DiskOffering, _, err = client.DiskOffering.GetDiskOfferingID(config.DiskOffering)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"disk offering", config.DiskOffering, err})
		}
	}

	if config.PublicIPAddress != "" {
		if isUUID(config.PublicIPAddress) {
			ip, _, err := client.Address.GetPublicIpAddressByID(config.PublicIPAddress)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Failed to retrieve IP address: %s", err))
			}
			state.Put("ipaddress", ip.Ipaddress)
		} else {
			// Save the public IP address before replacing it with it's UUID.
			state.Put("ipaddress", config.PublicIPAddress)

			p := client.Address.NewListPublicIpAddressesParams()
			p.SetIpaddress(config.PublicIPAddress)

			if config.Project != "" {
				p.SetProjectid(config.Project)
			}

			ipAddrs, err := client.Address.ListPublicIpAddresses(p)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"IP address", config.PublicIPAddress, err})
			}
			if err == nil && ipAddrs.Count != 1 {
				errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"IP address", config.PublicIPAddress, ipAddrs})
			}
			if err == nil && ipAddrs.Count == 1 {
				config.PublicIPAddress = ipAddrs.PublicIpAddresses[0].Id
			}
		}
	}

	if !isUUID(config.Network) {
		config.Network, _, err = client.Network.GetNetworkID(config.Network, cloudstack.WithProject(config.Project))
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"network", config.Network, err})
		}
	}

	// Then try to get the SG's UUID's.
	if len(config.SecurityGroups) > 0 {
		for i := range config.SecurityGroups {
			if !isUUID(config.SecurityGroups[i]) {
				config.SecurityGroups[i], _, err = client.SecurityGroup.GetSecurityGroupID(config.SecurityGroups[i], cloudstack.WithProject(config.Project))
				if err != nil {
					errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"network", config.SecurityGroups[i], err})
				}
			}
		}
	}

	if !isUUID(config.ServiceOffering) {
		config.ServiceOffering, _, err = client.ServiceOffering.GetServiceOfferingID(config.ServiceOffering)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"service offering", config.ServiceOffering, err})
		}
	}

	if config.SourceISO != "" {
		if isUUID(config.SourceISO) {
			state.Put("source", config.SourceISO)
		} else {
			isoID, _, err := client.ISO.GetIsoID(config.SourceISO, "executable", config.Zone)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"ISO", config.SourceISO, err})
			}
			state.Put("source", isoID)
		}
	}

	if config.SourceTemplate != "" {
		if isUUID(config.SourceTemplate) {
			state.Put("source", config.SourceTemplate)
		} else {
			templateID, _, err := client.Template.GetTemplateID(config.SourceTemplate, "executable", config.Zone)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"template", config.SourceTemplate, err})
			}
			state.Put("source", templateID)
		}
	}

	if !isUUID(config.TemplateOS) {
		p := client.GuestOS.NewListOsTypesParams()
		p.SetDescription(config.TemplateOS)

		types, err := client.GuestOS.ListOsTypes(p)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"OS type", config.TemplateOS, err})
		}
		if err == nil && types.Count != 1 {
			errs = packersdk.MultiErrorAppend(errs, &retrieveErr{"OS type", config.TemplateOS, types})
		}
		if err == nil && types.Count == 1 {
			config.TemplateOS = types.OsTypes[0].Id
		}
	}

	// This is needed because nil is not always nil. When returning *packersdk.MultiError(nil)
	// as an error interface, that interface will no longer be equal to nil but it will be
	// an interface with type *packersdk.MultiError and value nil which is different then a
	// nil interface.
	if errs != nil && len(errs.Errors) > 0 {
		state.Put("error", errs)
		ui.Error(errs.Error())
		return multistep.ActionHalt
	}

	ui.Message("Config has been prepared!")
	return multistep.ActionContinue
}

func (s *stepPrepareConfig) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}

type retrieveErr struct {
	name   string
	value  string
	result interface{}
}

func (e *retrieveErr) Error() string {
	if err, ok := e.result.(error); ok {
		e.result = err.Error()
	}
	return fmt.Sprintf("Error retrieving UUID of %s %s: %v", e.name, e.value, e.result)
}

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func isUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
