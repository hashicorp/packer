package cloudstack

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepSetupNetworking struct {
	privatePort int
	publicPort  int
}

func (s *stepSetupNetworking) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Setup networking...")

	if config.UseLocalIPAddress {
		ui.Message("Using the local IP address...")
		state.Put("commPort", config.Comm.Port())
		ui.Message("Networking has been setup!")
		return multistep.ActionContinue
	}

	if config.PublicPort != 0 {
		s.publicPort = config.PublicPort
	} else {
		// Generate a random public port used to configure our port forward.
		rand.Seed(time.Now().UnixNano())
		s.publicPort = 50000 + rand.Intn(10000)
	}
	state.Put("commPort", s.publicPort)

	// Set the currently configured port to be the private port.
	s.privatePort = config.Comm.Port()

	// Retrieve the instance ID from the previously saved state.
	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		err := fmt.Errorf("Could not retrieve instance_id from state!")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	network, _, err := client.Network.GetNetworkByID(
		config.Network,
		cloudstack.WithProject(config.Project),
	)
	if err != nil {
		err := fmt.Errorf("Failed to retrieve the network object: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if config.PublicIPAddress == "" {
		ui.Message("Associating public IP address...")
		p := client.Address.NewAssociateIpAddressParams()

		if config.Project != "" {
			p.SetProjectid(config.Project)
		}

		if network.Vpcid != "" {
			p.SetVpcid(network.Vpcid)
		} else {
			p.SetNetworkid(network.Id)
		}

		p.SetZoneid(config.Zone)

		// Associate a new public IP address.
		ipAddr, err := client.Address.AssociateIpAddress(p)
		if err != nil {
			err := fmt.Errorf("Failed to associate public IP address: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the IP address and it's ID.
		config.PublicIPAddress = ipAddr.Id
		state.Put("ipaddress", ipAddr.Ipaddress)

		// Store the IP address ID.
		state.Put("ip_address_id", ipAddr.Id)
	}

	ui.Message("Creating port forward...")
	p := client.Firewall.NewCreatePortForwardingRuleParams(
		config.PublicIPAddress,
		s.privatePort,
		"TCP",
		s.publicPort,
		instanceID,
	)

	// Configure the port forward.
	p.SetNetworkid(network.Id)
	p.SetOpenfirewall(false)

	// Create the port forward.
	forward, err := client.Firewall.CreatePortForwardingRule(p)
	if err != nil {
		err := fmt.Errorf("Failed to create port forward: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Store the port forward ID.
	state.Put("port_forward_id", forward.Id)

	if config.PreventFirewallChanges {
		ui.Message("Networking has been setup (without firewall changes)!")
		return multistep.ActionContinue
	}

	if network.Vpcid != "" {
		ui.Message("Creating network ACL rule...")

		if network.Aclid == "" {
			err := fmt.Errorf("Failed to configure the firewall: no ACL connected to the VPC network")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Create a new parameter struct.
		p := client.NetworkACL.NewCreateNetworkACLParams("TCP")

		// Configure the network ACL rule.
		p.SetAclid(network.Aclid)
		p.SetAction("allow")
		p.SetCidrlist(config.CIDRList)
		p.SetStartport(s.privatePort)
		p.SetEndport(s.privatePort)
		p.SetTraffictype("ingress")

		// Create the network ACL rule.
		aclRule, err := client.NetworkACL.CreateNetworkACL(p)
		if err != nil {
			err := fmt.Errorf("Failed to create network ACL rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Store the network ACL rule ID.
		state.Put("network_acl_rule_id", aclRule.Id)
	} else {
		ui.Message("Creating firewall rule...")

		// Create a new parameter struct.
		p := client.Firewall.NewCreateFirewallRuleParams(config.PublicIPAddress, "TCP")

		// Configure the firewall rule.
		p.SetCidrlist(config.CIDRList)
		p.SetStartport(s.publicPort)
		p.SetEndport(s.publicPort)

		fwRule, err := client.Firewall.CreateFirewallRule(p)
		if err != nil {
			err := fmt.Errorf("Failed to create firewall rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Store the firewall rule ID.
		state.Put("firewall_rule_id", fwRule.Id)
	}

	ui.Message("Networking has been setup!")
	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepSetupNetworking) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Cleanup networking...")

	if fwRuleID, ok := state.Get("firewall_rule_id").(string); ok && fwRuleID != "" {
		// Create a new parameter struct.
		p := client.Firewall.NewDeleteFirewallRuleParams(fwRuleID)

		ui.Message("Deleting firewall rule...")
		if _, err := client.Firewall.DeleteFirewallRule(p); err != nil {
			// This is a very poor way to be told the ID does no longer exist :(
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"Invalid parameter id value=%s due to incorrect long value format, "+
					"or entity does not exist", fwRuleID)) {
				ui.Error(fmt.Sprintf("Error deleting firewall rule: %s", err))
			}
		}
	}

	if aclRuleID, ok := state.Get("network_acl_rule_id").(string); ok && aclRuleID != "" {
		// Create a new parameter struct.
		p := client.NetworkACL.NewDeleteNetworkACLParams(aclRuleID)

		ui.Message("Deleting network ACL rule...")
		if _, err := client.NetworkACL.DeleteNetworkACL(p); err != nil {
			// This is a very poor way to be told the ID does no longer exist :(
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"Invalid parameter id value=%s due to incorrect long value format, "+
					"or entity does not exist", aclRuleID)) {
				ui.Error(fmt.Sprintf("Error deleting network ACL rule: %s", err))
			}
		}
	}

	if forwardID, ok := state.Get("port_forward_id").(string); ok && forwardID != "" {
		// Create a new parameter struct.
		p := client.Firewall.NewDeletePortForwardingRuleParams(forwardID)

		ui.Message("Deleting port forward...")
		if _, err := client.Firewall.DeletePortForwardingRule(p); err != nil {
			// This is a very poor way to be told the ID does no longer exist :(
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"Invalid parameter id value=%s due to incorrect long value format, "+
					"or entity does not exist", forwardID)) {
				ui.Error(fmt.Sprintf("Error deleting port forward: %s", err))
			}
		}
	}

	if ipAddrID, ok := state.Get("ip_address_id").(string); ok && ipAddrID != "" {
		// Create a new parameter struct.
		p := client.Address.NewDisassociateIpAddressParams(ipAddrID)

		ui.Message("Releasing public IP address...")
		if _, err := client.Address.DisassociateIpAddress(p); err != nil {
			// This is a very poor way to be told the ID does no longer exist :(
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"Invalid parameter id value=%s due to incorrect long value format, "+
					"or entity does not exist", ipAddrID)) {
				ui.Error(fmt.Sprintf("Error releasing public IP address: %s", err))
			}
		}
	}

	ui.Message("Networking has been cleaned!")
	return
}
