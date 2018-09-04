package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer/fix"
	"github.com/hashicorp/packer/template"

	"github.com/posener/complete"
)

type FixCommand struct {
	Meta
}

func (c *FixCommand) Run(args []string) int {
	var flagValidate bool
	flags := c.Meta.FlagSet("fix", FlagSetNone)
	flags.BoolVar(&flagValidate, "validate", true, "")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Read the file for decoding
	tplF, err := os.Open(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error opening template: %s", err))
		return 1
	}
	defer tplF.Close()

	// Decode the JSON into a generic map structure
	var templateData map[string]interface{}
	decoder := json.NewDecoder(tplF)
	if err := decoder.Decode(&templateData); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing template: %s", err))
		return 1
	}

	// Close the file since we're done with that
	tplF.Close()

	input := templateData
	for _, name := range fix.FixerOrder {
		var err error
		fixer, ok := fix.Fixers[name]
		if !ok {
			panic("fixer not found: " + name)
		}

		log.Printf("Running fixer: %s", name)
		input, err = fixer.Fix(input)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error fixing: %s", err))
			return 1
		}
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	if err := encoder.Encode(input); err != nil {
		c.Ui.Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, output.Bytes(), "", "  "); err != nil {
		c.Ui.Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	result := indented.String()
	result = strings.Replace(result, `\u003c`, "<", -1)
	result = strings.Replace(result, `\u003e`, ">", -1)
	c.Ui.Say(result)

	if flagValidate {
		// Attempt to parse and validate the template
		tpl, err := template.Parse(strings.NewReader(result))
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Error! Fixed template fails to parse: %s\n\n"+
					"This is usually caused by an error in the input template.\n"+
					"Please fix the error and try again.",
				err))
			return 1
		}
		if err := tpl.Validate(); err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Error! Fixed template failed to validate: %s\n\n"+
					"This is usually caused by an error in the input template.\n"+
					"Please fix the error and try again.",
				err))
			return 1
		}
	}

	return 0
}

func (*FixCommand) Help() string {
	helpText := `
Usage: packer fix [options] TEMPLATE

  Reads the JSON template and attempts to fix known backwards
  incompatibilities. The fixed template will be outputted to standard out.

  If the template cannot be fixed due to an error, the command will exit
  with a non-zero exit status. Error messages will appear on standard error.

Fixes that are run:

  iso-md5                    Replaces "iso_md5" in builders with newer
                             "iso_checksum"
  createtime                 Replaces ".CreateTime" in builder configs with
                             "{{timestamp}}"
  virtualbox-gaattach        Updates VirtualBox builders using
                             "guest_additions_attach" to use
                             "guest_additions_mode"
  pp-vagrant-override        Replaces old-style provider overrides for the
                             Vagrant post-processor to new-style as of Packer
                             0.5.0.
  virtualbox-rename          Updates "virtualbox" builders to "virtualbox-iso"
  vmware-rename              Updates "vmware" builders to "vmware-iso"
  parallels-headless         Removes unused "headless" setting from Parallels
                             builders
  parallels-deprecations     Removes deprecated "parallels_tools_host_path" from
                             Parallels builders
  sshkeypath                 Updates builders using "ssh_key_path" to use
                             "ssh_private_key_file"
  sshdisableagent            Updates builders using "ssh_disable_agent" to use
                             "ssh_disable_agent_forwarding"
  manifest-filename          Updates "manifest" post-processor so any "filename"
                             field is renamed to "output"
  amazon-shutdown_behavior   Changes "shutdown_behaviour" to "shutdown_behavior"
                             in Amazon builders
  amazon-enhanced-networking Replaces "enhanced_networking" in builders with
                             "ena_support"
  amazon-private-ip          Replaces "ssh_private_ip": true in amazon builders
                             with "ssh_interface": "private_ip"
  docker-email               Removes "login_email" from the Docker builder
  powershell-escapes         Removes PowerShell escapes from user env vars and
                             elevated username and password strings
  hyperv-deprecations        Removes the deprecated "vhd_temp_path" setting from
                             Hyper-V ISO builder templates
  hyperv-vmxc-typo           Corrects a typo in the "clone_from_vmxc_path"
                             setting. Replaces with "clone_from_vmcx_path".
  vmware-compaction          Adds "skip_compaction = true" to "vmware-iso"
                             builders with incompatible disk_type_id

Options:

  -validate=true      If true (default), validates the fixed template.
`

	return strings.TrimSpace(helpText)
}

func (c *FixCommand) Synopsis() string {
	return "fixes templates from old versions of packer"
}

func (c *FixCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *FixCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-validate": complete.PredictNothing,
	}
}
