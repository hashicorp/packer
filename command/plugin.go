package command

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/packer/builder/amazon/chroot"
	"github.com/mitchellh/packer/builder/amazon/ebs"
	"github.com/mitchellh/packer/builder/amazon/instance"
	"github.com/mitchellh/packer/builder/digitalocean"
	"github.com/mitchellh/packer/builder/docker"
	filebuilder "github.com/mitchellh/packer/builder/file"
	"github.com/mitchellh/packer/builder/googlecompute"
	"github.com/mitchellh/packer/builder/null"
	"github.com/mitchellh/packer/builder/openstack"
	parallelsiso "github.com/mitchellh/packer/builder/parallels/iso"
	parallelspvm "github.com/mitchellh/packer/builder/parallels/pvm"
	"github.com/mitchellh/packer/builder/qemu"
	virtualboxiso "github.com/mitchellh/packer/builder/virtualbox/iso"
	virtualboxovf "github.com/mitchellh/packer/builder/virtualbox/ovf"
	vmwareiso "github.com/mitchellh/packer/builder/vmware/iso"
	vmwarevmx "github.com/mitchellh/packer/builder/vmware/vmx"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/artifice"
	"github.com/mitchellh/packer/post-processor/atlas"
	"github.com/mitchellh/packer/post-processor/compress"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"github.com/mitchellh/packer/post-processor/docker-push"
	"github.com/mitchellh/packer/post-processor/docker-save"
	"github.com/mitchellh/packer/post-processor/docker-tag"
	"github.com/mitchellh/packer/post-processor/vagrant"
	"github.com/mitchellh/packer/post-processor/vagrant-cloud"
	"github.com/mitchellh/packer/post-processor/vsphere"
	"github.com/mitchellh/packer/provisioner/ansible-local"
	"github.com/mitchellh/packer/provisioner/chef-client"
	"github.com/mitchellh/packer/provisioner/chef-solo"
	fileprovisioner "github.com/mitchellh/packer/provisioner/file"
	"github.com/mitchellh/packer/provisioner/powershell"
	"github.com/mitchellh/packer/provisioner/puppet-masterless"
	"github.com/mitchellh/packer/provisioner/puppet-server"
	"github.com/mitchellh/packer/provisioner/salt-masterless"
	"github.com/mitchellh/packer/provisioner/shell"
	shelllocal "github.com/mitchellh/packer/provisioner/shell-local"
	"github.com/mitchellh/packer/provisioner/windows-restart"
	windowsshell "github.com/mitchellh/packer/provisioner/windows-shell"
)

type PluginCommand struct {
	Meta
}

var Builders = map[string]packer.Builder{
	"amazon-chroot":   new(chroot.Builder),
	"amazon-ebs":      new(ebs.Builder),
	"amazon-instance": new(instance.Builder),
	"digitalocean":    new(digitalocean.Builder),
	"docker":          new(docker.Builder),
	"file":            new(filebuilder.Builder),
	"googlecompute":   new(googlecompute.Builder),
	"null":            new(null.Builder),
	"openstack":       new(openstack.Builder),
	"parallels-iso":   new(parallelsiso.Builder),
	"parallels-pvm":   new(parallelspvm.Builder),
	"qemu":            new(qemu.Builder),
	"virtualbox-iso":  new(virtualboxiso.Builder),
	"virtualbox-ovf":  new(virtualboxovf.Builder),
	"vmware-iso":      new(vmwareiso.Builder),
	"vmware-vmx":      new(vmwarevmx.Builder),
}

var Provisioners = map[string]packer.Provisioner{
	"ansible-local":     new(ansiblelocal.Provisioner),
	"chef-client":       new(chefclient.Provisioner),
	"chef-solo":         new(chefsolo.Provisioner),
	"file":              new(fileprovisioner.Provisioner),
	"powershell":        new(powershell.Provisioner),
	"puppet-masterless": new(puppetmasterless.Provisioner),
	"puppet-server":     new(puppetserver.Provisioner),
	"salt-masterless":   new(saltmasterless.Provisioner),
	"shell":             new(shell.Provisioner),
	"shell-local":       new(shelllocal.Provisioner),
	"windows-restart":   new(restart.Provisioner),
	"windows-shell":     new(windowsshell.Provisioner),
}

var PostProcessors = map[string]packer.PostProcessor{
	"artifice":      new(artifice.PostProcessor),
	"atlas":         new(atlas.PostProcessor),
	"compress":      new(compress.PostProcessor),
	"docker-import": new(dockerimport.PostProcessor),
	"docker-push":   new(dockerpush.PostProcessor),
	"docker-save":   new(dockersave.PostProcessor),
	"docker-tag":    new(dockertag.PostProcessor),
	"vagrant":       new(vagrant.PostProcessor),
	"vagrant-cloud": new(vagrantcloud.PostProcessor),
	"vsphere":       new(vsphere.PostProcessor),
}

func (c *PluginCommand) Run(args []string) int {
	// This is an internal call so we're not going to do much error checking.
	// If there's a problem we'll usually just crash.
	log.Printf("args: %#v", args)
	if len(args) != 1 {
		c.Ui.Error("Wrong number of args")
		os.Exit(1)
	}

	// Plugin should be called like "packer-builder-amazon-ebs" so we'll take it
	// apart.
	parts := strings.Split(args[0], "-")
	pluginType := parts[1]
	pluginName := ""
	// Post-processor is split so we'll so some magic here. We could use a
	// regexp but this is simpler.
	if pluginType == "post" {
		pluginType = strings.Join(parts[1:2], "-")
		pluginName = strings.Join(parts[3:], "-")
	} else {
		pluginName = strings.Join(parts[2:], "-")
	}

	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	if pluginType == "builder" {
		builder, found := Builders[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load builder: %s", pluginName))
		}
		server.RegisterBuilder(builder)
	} else if pluginType == "provisioner" {
		provisioner, found := Provisioners[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load provisioner: %s", pluginName))
		}
		server.RegisterProvisioner(provisioner)
	} else if pluginType == "post-processor" {
		postProcessor, found := PostProcessors[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load post-processor: %s", pluginName))
		}
		server.RegisterPostProcessor(postProcessor)
	}

	server.Serve()

	return 0
}

func (*PluginCommand) Help() string {
	helpText := `
Usage: packer plugin PLUGIN

  Runs an internally-compiled version of a plugin from the packer binary. Note
  that this is an internal command and you should not call it yourself.
`

	return strings.TrimSpace(helpText)
}

func (c *PluginCommand) Synopsis() string {
	return "call an internal plugin"
}
