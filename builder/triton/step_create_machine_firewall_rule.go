package triton

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepCreateMachineFirewallRule struct{}

func (s *StepCreateMachineFirewallRule) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if config.MachineFirewallDetails.Empty() {
		ui.Say("No firewall to configure!")
		return multistep.ActionContinue
	}

	firewallRule := createFirewallRuleString(config.MachineFirewallDetails.SourceAddress, state.Get("machine").(string), config.MachineFirewallDetails.Port)
	ui.Say(fmt.Sprintf("Creating Firewall Rule: %s", firewallRule))

	ruleId, err := driver.CreateFirewallRule(firewallRule)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem creating firewall rule: %s", err))
		return multistep.ActionHalt
	}

	state.Put("firewall_rule_id", ruleId)

	return multistep.ActionContinue
}

func createFirewallRuleString(sourceAddress, machineId string, port int) string {
	sourceType := ""
	if sourceAddress != "any" {
		sourceType = "ip"
	}
	return fmt.Sprintf("FROM %s %s TO vm %s ALLOW tcp PORT %d", sourceType, sourceAddress, machineId, port)
}

func (s *StepCreateMachineFirewallRule) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)
	config := state.Get("config").(Config)

	if config.MachineFirewallDetails.Empty() {
		return
	}

	ui.Say("Deleting Firewall Rule")

	ruleId, ok := state.GetOk("firewall_rule_id")
	if !ok {
		return
	}

	err := driver.DeleteFirewallRule(ruleId.(string))
	if err != nil {
		state.Put("error", fmt.Errorf("Problem deleting firewall rule: %s", err))
		return
	}

	return
}
